package recommend

import (
	"fmt"
	"sort"
	"strings"

	"github.com/shra26/ollama-compatibility-checker/internal/estimate"
	"github.com/shra26/ollama-compatibility-checker/internal/hw"
)

const (
	minCtx = 2048
	kvNote = "OLLAMA_KV_CACHE_TYPE is server-wide and applies to all models."
)

// Config is copy-paste Ollama setup for one estimate.
type Config struct {
	PullCmd   string
	Modelfile string
	NumCtx    int
	NumGPU    int
	CtxNote   string
	Env       map[string]string // OLLAMA_FLASH_ATTENTION, OLLAMA_KV_CACHE_TYPE
	EnvSyntax map[string]string // "posix", "powershell", "windows-persistent", "macos-app"

	Runtime   string // "ollama" (default), "llamacpp" or "lmstudio"
	ModelName string // catalog display name, e.g. "Qwen2.5 7B"
	Quant     string // catalog quant, e.g. "q4_K_M"
	Layers    int    // total model layers, to turn NumGPU into a ratio
}

// BuildConfig builds the pull command, Modelfile and environment for e.
// If e is not a full GPU fit, the context is halved (powers of two, down to
// 2048) until the model fits fully on GPU; otherwise the requested context
// is kept with e's partial layer count.
func BuildConfig(h hw.Hardware, e estimate.Estimate, o Options) Config {
	final := e
	var notes []string
	if e.CtxReduced {
		notes = append(notes, fmt.Sprintf("capped at the model's maximum context of %d", e.Ctx))
	}
	if e.GPUFraction < 1 {
		for ctx := pow2AtOrBelow(e.Ctx); ctx >= minCtx; ctx /= 2 {
			if ctx >= e.Ctx {
				continue
			}
			p := estimate.DefaultParams()
			p.Ctx = ctx
			r := estimate.EstimateVariant(h, e.Model, e.Variant, p)
			if r.GPUFraction == 1 {
				final = r
				notes = append(notes, fmt.Sprintf("reduced from %d to %d so the model fits fully on GPU", e.Ctx, ctx))
				break
			}
		}
	}

	c := Config{
		PullCmd: "ollama pull " + final.Variant.Tag,
		NumCtx:  final.Ctx,
		NumGPU:  final.NumGPU,
		CtxNote: strings.Join(notes, "; "),
		Env:     map[string]string{},

		Runtime:   o.Runtime,
		ModelName: final.Model.Name,
		Quant:     final.Variant.Quant,
		Layers:    final.Model.Layers,
	}
	if c.Runtime == "" {
		c.Runtime = RuntimeOllama
	}
	c.Modelfile = fmt.Sprintf("FROM %s\nPARAMETER num_ctx %d\n", final.Variant.Tag, c.NumCtx)
	if c.NumGPU > 0 {
		c.Modelfile += fmt.Sprintf("PARAMETER num_gpu %d\n", c.NumGPU)
	}

	if c.NumGPU > 0 && usesVulkan(h) {
		c.Env["OLLAMA_VULKAN"] = "1"
	}
	if c.NumGPU > 0 && usesAccelerator(h) {
		c.Env["OLLAMA_FLASH_ATTENTION"] = "1"
		if final.KVQuant {
			c.Env["OLLAMA_KV_CACHE_TYPE"] = "q8_0"
		}
	}
	c.EnvSyntax = envSyntax(c.Env)
	return c
}

func pow2AtOrBelow(n int) int {
	p := 1
	for p*2 <= n {
		p *= 2
	}
	return p
}

// usesAccelerator reports whether the first supported GPU is on cuda, metal or rocm.
func usesAccelerator(h hw.Hardware) bool {
	for _, g := range h.GPUs {
		if g.Supported {
			return g.Backend == "cuda" || g.Backend == "metal" || g.Backend == "rocm"
		}
	}
	return false
}

// usesVulkan reports whether the first supported GPU is on the experimental Vulkan backend.
func usesVulkan(h hw.Hardware) bool {
	for _, g := range h.GPUs {
		if g.Supported {
			return g.Backend == "vulkan"
		}
	}
	return false
}

func envSyntax(env map[string]string) map[string]string {
	if len(env) == 0 {
		return nil
	}
	keys := make([]string, 0, len(env))
	for k := range env {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	_, hasKV := env["OLLAMA_KV_CACHE_TYPE"]

	build := func(line func(k, v string) string, comment string) string {
		var b strings.Builder
		for _, k := range keys {
			b.WriteString(line(k, env[k]))
			b.WriteString("\n")
		}
		if hasKV {
			b.WriteString(comment + " " + kvNote + "\n")
		}
		return b.String()
	}
	return map[string]string{
		"posix":              build(func(k, v string) string { return "export " + k + "=" + v }, "#"),
		"powershell":         build(func(k, v string) string { return "$env:" + k + `="` + v + `"` }, "#"),
		"windows-persistent": build(func(k, v string) string { return "setx " + k + " " + v }, "REM"),
		"macos-app":          build(func(k, v string) string { return "launchctl setenv " + k + " " + v }, "#"),
	}
}
