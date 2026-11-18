// Package estimate computes memory fit and decode tokens/sec for catalog
// variants on a given machine. All functions are pure.
package estimate

import (
	"math"

	"github.com/shra26/ollama-compatibility-checker/internal/calib"
	"github.com/shra26/ollama-compatibility-checker/internal/catalog"
	"github.com/shra26/ollama-compatibility-checker/internal/hw"
)

const (
	runtimeOverhead = 512 << 20 // 512 MiB
	vramHeadroom    = 0.9       // keep 10% of VRAM free
	ramExcludeFrac  = 0.8       // models needing more than this share of RAM are excluded
	vulkanEffFactor = 0.6       // Vulkan reaches a smaller share of bandwidth than CUDA/Metal/ROCm
)

// Params tunes the estimator.
type Params struct {
	Ctx    int
	EffGPU float64 // GPU backends without a fitted factor
	EffCPU float64
	// Fitted holds per-backend factors fitted from shared bench records
	// (calibration/data); they replace EffGPU on their backend.
	Fitted calib.Table
}

// DefaultParams returns the default constants and the embedded fitted factors.
func DefaultParams() Params {
	return Params{Ctx: 8192, EffGPU: 0.65, EffCPU: 0.55, Fitted: calib.Factors()}
}

// gpuEff is the bandwidth efficiency for g's backend.
func (p Params) gpuEff(g hw.GPU) float64 {
	if f, ok := p.Fitted[g.Backend]; ok && f.Eff > 0 {
		return f.Eff
	}
	if g.Backend == "vulkan" {
		return p.EffGPU * vulkanEffFactor
	}
	return p.EffGPU
}

// Estimate is the fit and speed prediction for one variant.
type Estimate struct {
	Model         catalog.Model
	Variant       catalog.Variant
	Ctx           int
	CtxReduced    bool
	KVQuant       bool // q8_0 cache assumed
	NeedBytes     uint64
	GPUFraction   float64
	NumGPU        int
	TokS          float64
	Excluded      bool
	LowConfidence bool
	ErrPct        int // expected error of TokS in percent, see errPct
}

// Expected error of the speed estimate, in percent. A backend fitted from
// bench records uses its fitted error when the model runs fully on GPU;
// unmeasured backends and CPU get the README's plus or minus 30 percent
// target. Partial offload and unknown GPUs (guessed bandwidth) are less
// predictable.
const (
	errDefault = 30
	errPartial = 40
	errLowConf = 50
)

func (p Params) errPct(backend string, gpuFraction float64, lowConf bool) int {
	switch {
	case lowConf:
		return errLowConf
	case gpuFraction > 0 && gpuFraction < 1:
		return errPartial
	case gpuFraction == 1:
		if f, ok := p.Fitted[backend]; ok && f.ErrPct > 0 {
			return f.ErrPct
		}
	}
	return errDefault
}

// TokSRange is TokS minus and plus ErrPct.
func (e Estimate) TokSRange() (low, high float64) {
	f := float64(e.ErrPct) / 100
	return e.TokS * (1 - f), e.TokS * (1 + f)
}

// KVBytes is the KV cache size at ctx tokens (q8_0: 1 byte/elem, f16: 2).
// Hybrid and sliding-window models use their recorded layout: full layers
// grow with ctx, sliding-window layers stop at the window.
func KVBytes(m catalog.Model, ctx int, q8 bool) uint64 {
	per := uint64(2)
	if q8 {
		per = 1
	}
	if m.KVFullElems > 0 || m.KVSWAElems > 0 {
		swaCtx := min(ctx, m.SWAWindow)
		return (uint64(m.KVFullElems)*uint64(ctx) + uint64(m.KVSWAElems)*uint64(swaCtx)) * per
	}
	return 2 * uint64(m.Layers) * uint64(m.KVHeads) * uint64(m.HeadDim) * uint64(ctx) * per
}

// gpuPool returns the first supported GPU, if any.
func gpuPool(h hw.Hardware) (hw.GPU, bool) {
	for _, g := range h.GPUs {
		if g.Supported {
			return g, true
		}
	}
	return hw.GPU{}, false
}

// UsableVRAM is 90% of VRAM, or free VRAM when known and smaller. 0 if CPU-only.
func UsableVRAM(h hw.Hardware) uint64 {
	g, ok := gpuPool(h)
	if !ok {
		return 0
	}
	u := uint64(float64(g.VRAMBytes) * vramHeadroom)
	if g.FreeVRAMBytes != 0 && g.FreeVRAMBytes < u {
		u = g.FreeVRAMBytes
	}
	return u
}

// NumGPULayers is the number of whole layers to offload when the model does
// not fully fit: floor((usable - overhead - per_layer)/(per_layer+kv_per_layer)) - 1,
// clamped to [0, layers].
func NumGPULayers(m catalog.Model, v catalog.Variant, usable, kv uint64) int {
	if m.Layers <= 0 {
		return 0
	}
	layers := float64(m.Layers)
	perLayer := float64(v.SizeBytes) / (layers + 1)
	kvPerLayer := float64(kv) / layers
	n := math.Floor((float64(usable)-runtimeOverhead-perLayer)/(perLayer+kvPerLayer)) - 1
	if math.IsNaN(n) || n < 0 {
		return 0
	}
	if n > layers {
		return m.Layers
	}
	return int(n)
}

// EstimateVariant predicts fit and decode speed for one variant.
func EstimateVariant(h hw.Hardware, m catalog.Model, v catalog.Variant, p Params) Estimate {
	e := Estimate{Model: m, Variant: v, Ctx: p.Ctx}
	if m.MaxCtx > 0 && p.Ctx > m.MaxCtx {
		e.Ctx, e.CtxReduced = m.MaxCtx, true
	}
	g, hasGPU := gpuPool(h)
	if hasGPU {
		switch g.Backend {
		case "cuda", "metal", "rocm":
			e.KVQuant = m.KVQuantOK
		}
		e.LowConfidence = g.BandwidthLowConfidence
	}
	kv := KVBytes(m, e.Ctx, e.KVQuant)
	e.NeedBytes = v.SizeBytes + v.ExtraBytes + m.StateBytes + kv + runtimeOverhead
	usable := UsableVRAM(h)

	switch {
	case !hasGPU:
		e.GPUFraction, e.NumGPU = 0, 0
	case e.NeedBytes <= usable:
		e.GPUFraction, e.NumGPU = 1, 99
	default:
		room := usable // the projector sits on the GPU before any layer
		if v.ExtraBytes < room {
			room -= v.ExtraBytes
		} else {
			room = 0
		}
		e.NumGPU = NumGPULayers(m, v, room, kv)
		e.GPUFraction = float64(e.NumGPU) / float64(m.Layers)
	}

	ram := float64(h.RAMBytes)
	if h.UnifiedMemory {
		e.Excluded = float64(e.NeedBytes) > ramExcludeFrac*ram
	} else {
		e.Excluded = float64(e.NeedBytes) > float64(usable)+ramExcludeFrac*ram
	}

	bpt := float64(v.SizeBytes)
	if m.ParamsB > 0 {
		bpt *= m.ActiveParamsB / m.ParamsB
	}
	var tGPU, tCPU float64
	if e.GPUFraction > 0 {
		tGPU = bpt * e.GPUFraction / (g.BandwidthGBs * 1e9 * p.gpuEff(g))
	}
	if e.GPUFraction < 1 {
		tCPU = bpt * (1 - e.GPUFraction) / (h.RAMBandwidthGBs * 1e9 * p.EffCPU)
	}
	if t := tGPU + tCPU; t > 0 {
		e.TokS = 1 / t
	}
	e.ErrPct = p.errPct(g.Backend, e.GPUFraction, e.LowConfidence)
	return e
}

// EstimateAll returns one Estimate per variant of every model, in catalog
// order. Excluded estimates are included; callers filter.
func EstimateAll(h hw.Hardware, ms []catalog.Model, p Params) []Estimate {
	var out []Estimate
	for _, m := range ms {
		for _, v := range m.Variants {
			out = append(out, EstimateVariant(h, m, v, p))
		}
	}
	return out
}
