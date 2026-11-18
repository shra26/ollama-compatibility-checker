package hw

import (
	"fmt"
	"regexp"
	"strings"
)

// sharedMemFrac is the share of system RAM an integrated GPU may use
// (Windows lets the GPU borrow about half of RAM as shared memory).
const sharedMemFrac = 0.5

// integratedName matches integrated GPU marketing names, which have no
// dedicated VRAM and read system memory.
var integratedName = regexp.MustCompile(`(?i)\b(iris|uhd|hd graphics|xe graphics|radeon graphics|radeon\(tm\) graphics|vega \d+ graphics)\b`)

// EnableVulkan re-labels GPUs that Ollama's native backends do not cover
// (Intel, and AMD cards off the ROCm list) as Vulkan devices, for Ollama's
// experimental Vulkan backend. Integrated GPUs get a share of system RAM as
// VRAM and the system RAM bandwidth, both flagged low confidence. GPUs that
// already have a supported backend are left alone.
func EnableVulkan(gpus []GPU, ramBytes uint64, ramBandwidthGBs float64) ([]GPU, []string) {
	var warns []string
	out := make([]GPU, len(gpus))
	copy(out, gpus)
	for i, g := range out {
		if g.Supported || (g.Vendor != "intel" && g.Vendor != "amd") {
			continue
		}
		g.Backend, g.Supported = "vulkan", true
		if integratedName.MatchString(g.Name) {
			g.VRAMBytes = uint64(float64(ramBytes) * sharedMemFrac)
			g.FreeVRAMBytes = 0
			g.BandwidthGBs = ramBandwidthGBs
			g.BandwidthLowConfidence = true
		}
		out[i] = g
		warns = append(warns, fmt.Sprintf("--vulkan: estimating %q on Ollama's experimental Vulkan backend (set OLLAMA_VULKAN=1); speeds are rough", g.Name))
	}
	return out, warns
}

// dropUnsupportedWarnings removes "treating as unsupported" warnings for GPUs
// that EnableVulkan has since made usable.
func dropUnsupportedWarnings(warns []string, before, after []GPU) []string {
	var kept []string
next:
	for _, w := range warns {
		if strings.Contains(w, "treating as unsupported") {
			for i, g := range before {
				if !g.Supported && after[i].Supported && strings.Contains(w, fmt.Sprintf("%q", g.Name)) {
					continue next
				}
			}
		}
		kept = append(kept, w)
	}
	return kept
}
