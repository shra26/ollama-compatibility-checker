package main

import (
	"context"
	"io"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/shra26/ollama-compatibility-checker/internal/hw"
)

// version is set at build time via -ldflags "-X main.version=...".
var version = "dev"

// vulkanFlag is set by --vulkan and read by hostDetect.
var vulkanFlag bool

// detectFunc profiles the host. Commands take it as a parameter so tests can
// inject synthetic hardware.
type detectFunc func(context.Context) hw.Hardware

// hostDetect profiles the real machine.
func hostDetect(ctx context.Context) hw.Hardware {
	return hw.Detect(ctx, hw.Probes{Run: hw.DefaultRunner, SysRoot: "/", GOOS: runtime.GOOS, GOARCH: runtime.GOARCH, Vulkan: vulkanFlag})
}

// newRootCmd builds the root command wired to the real host.
func newRootCmd(stdout, stderr io.Writer) *cobra.Command {
	return newRootCmdWith(stdout, stderr, hostDetect)
}

// newRootCmdWith builds the root command with an injected detector. With no
// subcommand it runs recommend.
func newRootCmdWith(stdout, stderr io.Writer, detect detectFunc) *cobra.Command {
	return newRootCmdEnv(stdout, stderr, detect, hostBenchEnv())
}

// newRootCmdEnv also injects the bench environment (Ollama client, TTY check).
func newRootCmdEnv(stdout, stderr io.Writer, detect detectFunc, env benchEnv) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "ollama-compat",
		Short:        "Recommend Ollama models that fit your hardware",
		Version:      version,
		SilenceUsage: true,
		Args:         cobra.NoArgs,
		RunE:         recommendRunE(detect),
	}
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)
	cmd.SetVersionTemplate("ollama-compat {{.Version}}\n")

	pf := cmd.PersistentFlags()
	pf.Bool("json", false, "emit machine-readable JSON")
	pf.String("catalog", "", "path to a custom model catalog file")
	pf.Bool("no-color", false, "disable colored output")
	pf.BoolVar(&vulkanFlag, "vulkan", false, "estimate Intel and unsupported AMD GPUs on Ollama's experimental Vulkan backend")
	addRecommendFlags(cmd)

	cmd.AddCommand(newScanCmd(detect), newRecommendCmd(detect), newCatalogCmd(), newBenchCmd(detect, env))
	return cmd
}
