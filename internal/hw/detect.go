package hw

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Probes carries everything Detect needs, so tests can fake the host.
type Probes struct {
	Run     Runner
	SysRoot string // root of /proc and /sys on Linux; "/" if empty
	GOOS    string
	GOARCH  string
	Vulkan  bool // treat Intel/unsupported AMD GPUs as experimental Vulkan devices
}

// MergeGPUs pools supported GPUs of one vendor into a single entry (sum of
// VRAM, minimum bandwidth). With several supported vendors, only the vendor
// group with the most VRAM is kept, with a warning. Unsupported GPUs follow,
// untouched.
func MergeGPUs(gpus []GPU) ([]GPU, []string) {
	var supported, unsupported []GPU
	for _, g := range gpus {
		if g.Supported {
			supported = append(supported, g)
		} else {
			unsupported = append(unsupported, g)
		}
	}
	var warns []string
	var out []GPU
	if len(supported) > 0 {
		groups := map[string][]GPU{}
		var order []string
		total := map[string]uint64{}
		for _, g := range supported {
			if _, ok := groups[g.Vendor]; !ok {
				order = append(order, g.Vendor)
			}
			groups[g.Vendor] = append(groups[g.Vendor], g)
			total[g.Vendor] += g.VRAMBytes
		}
		best := order[0]
		for _, v := range order[1:] {
			if total[v] > total[best] {
				best = v
			}
		}
		if len(order) > 1 {
			warns = append(warns, fmt.Sprintf("mixed vendors detected (%s), using only the %s GPUs", strings.Join(order, ", "), best))
		}
		out = append(out, poolGPUs(groups[best]))
	}
	return append(out, unsupported...), warns
}

func poolGPUs(g []GPU) GPU {
	if len(g) == 1 {
		return g[0]
	}
	p := g[0]
	names := make([]string, len(g))
	p.VRAMBytes, p.FreeVRAMBytes = 0, 0
	allFree := true
	p.BandwidthLowConfidence = false
	for i, x := range g {
		names[i] = x.Name
		p.VRAMBytes += x.VRAMBytes
		p.FreeVRAMBytes += x.FreeVRAMBytes
		if x.FreeVRAMBytes == 0 {
			allFree = false
		}
		if x.BandwidthGBs < p.BandwidthGBs {
			p.BandwidthGBs = x.BandwidthGBs
		}
		p.BandwidthLowConfidence = p.BandwidthLowConfidence || x.BandwidthLowConfidence
	}
	if !allFree {
		p.FreeVRAMBytes = 0
	}
	p.Name = fmt.Sprintf("%dx GPU (%s)", len(g), strings.Join(names, " + "))
	return p
}

// Detect profiles the host. Detector failures become Warnings.
func Detect(ctx context.Context, p Probes) Hardware {
	hw := Hardware{OS: p.GOOS, Arch: p.GOARCH}
	root := p.SysRoot
	if root == "" {
		root = "/"
	}
	var gpus []GPU
	add := func(g []GPU, w []string) {
		gpus = append(gpus, g...)
		hw.Warnings = append(hw.Warnings, w...)
	}
	warn := func(format string, a ...any) { hw.Warnings = append(hw.Warnings, fmt.Sprintf(format, a...)) }

	switch p.GOOS {
	case "linux":
		add(DetectNvidia(ctx, p.Run))
		add(DetectAMD(ctx, p.Run, root))
		add(DetectIntel(ctx, p.Run))
		if b, err := os.ReadFile(filepath.Join(root, "proc", "meminfo")); err != nil {
			warn("reading meminfo failed (%v)", err)
		} else if n, err := parseMemTotal(b); err != nil {
			warn("meminfo: %v", err)
		} else {
			hw.RAMBytes = n
		}
		if b, err := os.ReadFile(filepath.Join(root, "proc", "cpuinfo")); err != nil {
			warn("reading cpuinfo failed (%v)", err)
		} else {
			hw.CPU = ParseCPUInfo(b)
		}
	case "windows":
		add(DetectNvidia(ctx, p.Run))
		if adapters, err := DetectWindowsRegistry(); err != nil {
			warn("windows registry GPU detection failed (%v)", err)
		} else {
			add(ParseWindowsAdapters(adapters))
		}
		if n, err := detectWindowsRAM(); err != nil {
			warn("windows RAM detection failed (%v)", err)
		} else {
			hw.RAMBytes = n
		}
		c, err := detectWindowsCPU()
		if err != nil {
			warn("windows CPU detection failed (%v)", err)
		}
		hw.CPU = c
	case "darwin":
		g, ram, chip, w := DetectApple(ctx, p.Run, p.GOARCH)
		add(g, w)
		hw.RAMBytes = ram
		hw.CPU = detectDarwinCPU(ctx, p.Run)
		if hw.CPU.Model == "" {
			hw.CPU.Model = chip
		}
		hw.UnifiedMemory = len(g) > 0 && g[0].Backend == "metal"
	default:
		warn("unsupported OS %q, hardware detection skipped", p.GOOS)
	}

	// Intel Macs run on CPU; Vulkan is not an option there.
	if p.Vulkan && p.GOOS != "darwin" {
		var vw []string
		before := gpus
		gpus, vw = EnableVulkan(gpus, hw.RAMBytes, RAMBandwidthFallback(p.GOOS, p.GOARCH))
		hw.Warnings = dropUnsupportedWarnings(hw.Warnings, before, gpus)
		hw.Warnings = append(hw.Warnings, vw...)
	}

	merged, w := MergeGPUs(gpus)
	hw.GPUs = merged
	hw.Warnings = append(hw.Warnings, w...)

	if hw.UnifiedMemory && len(hw.GPUs) > 0 {
		hw.RAMBandwidthGBs = hw.GPUs[0].BandwidthGBs
	} else {
		hw.RAMBandwidthGBs = RAMBandwidthFallback(p.GOOS, p.GOARCH)
	}
	return hw
}
