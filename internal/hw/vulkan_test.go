package hw

import "testing"

func TestEnableVulkan(t *testing.T) {
	in := []GPU{
		{Vendor: "intel", Name: "Intel(R) Iris(R) Xe Graphics", VRAMBytes: 128 << 20, Backend: "cpu"},
		{Vendor: "intel", Name: "Intel(R) Arc(TM) A770", VRAMBytes: 16 * gib, BandwidthGBs: 560, Backend: "cpu"},
		{Vendor: "amd", Name: "AMD Radeon RX 7800 XT", VRAMBytes: 16 * gib, Backend: "rocm", Supported: true},
	}
	out, warns := EnableVulkan(in, 32*gib, 50)
	if len(warns) != 2 {
		t.Fatalf("warns = %v", warns)
	}
	igpu := out[0]
	if igpu.Backend != "vulkan" || !igpu.Supported || igpu.VRAMBytes != 16*gib || igpu.BandwidthGBs != 50 || !igpu.BandwidthLowConfidence {
		t.Errorf("iGPU wrong: %+v", igpu)
	}
	arc := out[1]
	if arc.Backend != "vulkan" || !arc.Supported || arc.VRAMBytes != 16*gib || arc.BandwidthGBs != 560 {
		t.Errorf("arc wrong: %+v", arc)
	}
	if out[2].Backend != "rocm" {
		t.Errorf("rocm GPU changed: %+v", out[2])
	}
	if in[0].Backend != "cpu" {
		t.Error("input mutated")
	}
}

func TestDropUnsupportedWarnings(t *testing.T) {
	before := []GPU{{Name: "Intel Iris"}, {Name: "Old AMD"}}
	after := []GPU{{Name: "Intel Iris", Supported: true}, {Name: "Old AMD"}}
	warns := []string{
		`windows registry: Intel GPU "Intel Iris" has no supported Ollama backend, treating as unsupported`,
		`windows registry: "Old AMD" is not on the list, treating as unsupported`,
		"other warning",
	}
	got := dropUnsupportedWarnings(warns, before, after)
	if len(got) != 2 || got[0] != warns[1] || got[1] != warns[2] {
		t.Errorf("got %v", got)
	}
}
