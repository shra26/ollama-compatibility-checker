# ollama-compat

[![CI](https://github.com/shra26/ollama-compatibility-checker/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/shra26/ollama-compatibility-checker/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/shra26/ollama-compatibility-checker)](https://github.com/shra26/ollama-compatibility-checker/releases/latest)
[![Downloads](https://img.shields.io/github/downloads/shra26/ollama-compatibility-checker/total)](https://github.com/shra26/ollama-compatibility-checker/releases)
[![Go](https://img.shields.io/github/go-mod/go-version/shra26/ollama-compatibility-checker)](go.mod)
[![License: MIT](https://img.shields.io/github/license/shra26/ollama-compatibility-checker)](LICENSE)
[![Platforms](https://img.shields.io/badge/platforms-macOS%20%7C%20Linux%20%7C%20Windows-blue)](#install)
[![Homebrew](https://img.shields.io/badge/brew-shra26%2Ftap-orange)](#homebrew-macos-linux)
[![Telemetry: none](https://img.shields.io/badge/telemetry-none-brightgreen)](#privacy)

**Run one command. See which Ollama models your machine can actually run, and how fast.**
No forms to fill in, no account, no network.

https://github.com/user-attachments/assets/e693299b-c213-4b11-858c-289b6352fafd


*Click to watch the 20-second intro (with sound). The RTX 4090 scene is
real output for an example hardware profile.*

```bash
brew install --cask shra26/tap/ollama-compat   # or see Install for Windows, Linux, curl
ollama-compat
```

## Why ollama-compat

- **It detects your hardware.** No picking your GPU from a list or guessing
  your VRAM. It reads the GPU, VRAM (free VRAM too, on NVIDIA and AMD), memory
  bandwidth, RAM and CPU itself: NVIDIA, AMD, Apple Silicon, Intel or CPU only.
- **It works offline.** A normal run makes no network calls and sends no
  telemetry. The model catalog is built into a single static binary, so it
  works on a plane or an air-gapped box. Only the optional `bench` talks to
  your local Ollama, and it asks before downloading anything.
- **It sizes the context, not just the weights.** It works out the KV cache
  for your context length, splits layers between GPU and CPU when a model does
  not fully fit, and shrinks the context so the model you pick runs fully on
  your GPU.
- **It hands you the config.** Copy-paste `ollama pull`, environment variables
  for your OS and a Modelfile, or a `llama-server` / `lms load` command with
  `--runtime`.
- **It is honest about speed.** Every estimate carries its expected error
  (`~20 tok/s ±25%`), fitted from real bench runs people share. `bench`
  measures the real number on your machine.
- **It fits in scripts.** `--json` output, one binary for macOS, Linux and
  Windows on amd64 and arm64, MIT licensed.

Picks come in three flavors: **Fastest**, **Balanced** (best quality that
still reaches 20 tok/s) and **Largest** that runs at a usable speed. Add
`--task coding`, `reasoning` or `vision` to narrow them.

New here? The **[Getting started guide](docs/getting-started.md)** walks
through installing Ollama and this tool on macOS, Linux and Windows, reading
the output, applying the config, and troubleshooting.

## Quick start

```bash
# 1. Install (or download a release, or build from source; see Install below)
brew install --cask shra26/tap/ollama-compat

# 2. Get recommendations (offline, about one second)
ollama-compat

# 3. Use the Balanced pick (copy the exact commands from your output)
ollama pull <tag from the output>
# save the Modelfile lines from the output to a file named Modelfile, then:
ollama create <name from the output> -f Modelfile
ollama run <name from the output> --verbose
```

## Install

### Homebrew (macOS, Linux)

```bash
brew install --cask shra26/tap/ollama-compat
```

### Scoop (Windows)

```powershell
scoop bucket add shra26 https://github.com/shra26/scoop-bucket
scoop install ollama-compat
```

### Install script

macOS and Linux (installs to `~/.local/bin`; set `INSTALL_DIR` or `VERSION`
to override). It verifies the SHA-256 checksum before installing:

```bash
curl -fsSL https://raw.githubusercontent.com/shra26/ollama-compatibility-checker/main/scripts/install.sh | sh
```

Windows PowerShell (installs to `%LOCALAPPDATA%\Programs\ollama-compat`; save
the script and run `./install.ps1 -Version v0.1.1 -AddToPath` for options):

```powershell
irm https://raw.githubusercontent.com/shra26/ollama-compatibility-checker/main/scripts/install.ps1 | iex
```

### Prebuilt binaries

Download the archive for your platform from
[GitHub Releases](https://github.com/shra26/ollama-compatibility-checker/releases),
unpack it and put `ollama-compat` on your `PATH`. Builds exist for Linux,
macOS and Windows on amd64 and arm64 (use `darwin_arm64` for Apple Silicon
Macs and `windows_amd64` for most Windows PCs). If no release is listed yet,
build from source.

On macOS, a downloaded binary may be blocked by Gatekeeper. Clear the
quarantine flag once:

```bash
xattr -d com.apple.quarantine ollama-compat
```

On Windows, SmartScreen may warn about an unrecognized app; choose "More
info", then "Run anyway".

### From source

Requires [Go 1.24 or newer](https://go.dev/dl/).

```bash
git clone https://github.com/shra26/ollama-compatibility-checker.git
cd ollama-compatibility-checker
go build -o ollama-compat ./cmd/ollama-compat
```

Or install into your Go bin directory (`~/go/bin`, which must be on your
`PATH`):

```bash
go install github.com/shra26/ollama-compatibility-checker/cmd/ollama-compat@latest
```

### Ollama

`recommend`, `scan` and `catalog` do not need Ollama. To run the recommended
model, or to use `bench`, install Ollama from <https://ollama.com/download>
and make sure it is running (`curl http://127.0.0.1:11434/api/version`).

## Usage

```
ollama-compat [flags]
ollama-compat [command]

Commands:
  recommend   Scan hardware and recommend models (default command)
  scan        Report detected hardware
  catalog     List models in the catalog
  bench       Benchmark the shortlist on a local Ollama (opt-in, may download models)

Flags:
      --catalog string   path to a custom model catalog file
      --ctx int          context length to size the KV cache for (default 8192)
      --json             emit machine-readable JSON
      --min-tps float    Balanced pick must reach this estimated tok/s (default 20)
      --no-color         disable colored output
      --runtime string   runtime to print config for: ollama, llamacpp, lmstudio (default "ollama")
      --task string      only recommend models tagged for this task: coding, reasoning, vision
      --top int          how many alternatives to list per pick (default 3)
  -v, --version          version for ollama-compat
```

Running `ollama-compat` with no command is the same as `ollama-compat recommend`.

### recommend

Scans the hardware, estimates every catalog variant at the chosen context
length, and prints three picks, each with runners-up:

- **Fastest**: highest est. tok/s among models with a quality score of at least 40 that fit fully on the GPU (falls back to partial offload if none fit).
- **Balanced**: highest quality reaching `--min-tps`.
- **Largest**: highest quality reaching at least 5 tok/s, partial offload allowed.

It then prints the Ollama config for the Balanced pick: the `ollama pull`
command, environment variables, a Modelfile and the `ollama create` command.
Environment variable syntax is shown for your OS. How to apply each part on
macOS, Linux (systemd) and Windows is in
[Getting started, step 4](docs/getting-started.md#4-apply-the-recommended-config).

Example output on an Apple M1 with 8 GB of RAM (your numbers will differ):

```
Hardware
  GPU 0  Apple M1  5.7 GB VRAM  68 GB/s  Metal
  RAM    8.6 GB  ~68 GB/s
  CPU    Apple M1  8C/8T

Recommendations  (ctx 8192, estimates ± expected error)
  Note: Apple Silicon speeds are fitted from 1 bench run on 1 machine. Add yours: `ollama-compat bench --share`.
  Fastest   granite4:7b-a1b-h                 ~64 tok/s ±25%  4.9 GB  100% GPU
      also  lfm3:1.7b-q4_K_M                  ~35 tok/s ±25%  1.7 GB  100% GPU
      also  qwen3.5:2b-q4_K_M                 ~30 tok/s ±25%  3.0 GB  100% GPU
      also  gemma4:e2b-it-q4_K_M              ~25 tok/s ±25%  5.1 GB  100% GPU
  Balanced  lfm2.5:8b-a1b-q4_K_M              ~62 tok/s ±40%  5.7 GB  79% GPU
      also  granite4:7b-a1b-h                 ~64 tok/s ±25%  4.9 GB  100% GPU
      also  gemma4:e2b-it-q4_K_M              ~25 tok/s ±25%  5.1 GB  100% GPU
      also  lfm3:1.7b-q8_0                    ~21 tok/s ±25%  2.4 GB  100% GPU
  Largest   deepseek-r1:8b-0528-qwen3-q4_K_M  ~7 tok/s ±40%   6.4 GB  75% GPU
      also  qwen3:4b-thinking-2507-q8_0       ~9 tok/s ±40%   5.4 GB  89% GPU
      also  qwen3:8b-q4_K_M                   ~7 tok/s ±40%   6.4 GB  75% GPU
      also  qwen3:4b-instruct-2507-q8_0       ~9 tok/s ±40%   5.4 GB  89% GPU
  With 16 GB unified memory, Largest would be qwen3.5:9b-q4_K_M (~7 tok/s ±25%, needs 8.5 GB).

Config for Balanced
  ollama pull lfm2.5:8b-a1b-q4_K_M
  env:
    export OLLAMA_FLASH_ATTENTION=1
    export OLLAMA_KV_CACHE_TYPE=q8_0
    # OLLAMA_KV_CACHE_TYPE is server-wide and applies to all models.
  env (for the Ollama.app):
    launchctl setenv OLLAMA_FLASH_ATTENTION 1
    launchctl setenv OLLAMA_KV_CACHE_TYPE q8_0
    # OLLAMA_KV_CACHE_TYPE is server-wide and applies to all models.
  Restart Ollama after setting these.
  Modelfile:
    FROM lfm2.5:8b-a1b-q4_K_M
    PARAMETER num_ctx 8192
    PARAMETER num_gpu 19
  ollama create lfm2.5-8b-a1b-q4_K_M-tuned -f Modelfile

Run `ollama-compat bench` to measure these (downloads up to ~14.6 GB, asks first).
```

Useful flags: `--ctx 16384` sizes the KV cache for a longer context,
`--min-tps 30` raises the speed bar for the Balanced pick, `--top 5` lists
more runners-up, `--json` prints machine-readable output.

#### Picks for a task

`--task` limits the picks to catalog models tagged for one job:

| Task | Models | Meaning |
|---|---|---|
| `coding` | Qwen2.5 Coder 1.5B to 32B | code-tuned |
| `reasoning` | Qwen3, DeepSeek-R1 distills | has a thinking mode |
| `vision` | Gemma 3 4B/12B/27B, Mistral Small 3.2 | accepts images in Ollama |

```bash
ollama-compat --task coding
```

The same rules apply inside the smaller set, so a pick can come back empty
(for example no vision model reaches 20 tok/s on an 8 GB Mac). Vision sizes
cover the language model only; the image encoder adds a little memory.

#### Other runtimes

The config block is for Ollama by default. `--runtime` prints the same
choices (context, GPU layers, flash attention, KV cache type) for another
runtime:

```bash
ollama-compat --runtime llamacpp   # a llama-server command
ollama-compat --runtime lmstudio   # an lms load command
```

The catalog only knows Ollama tags, so the model path is a placeholder such
as `/path/to/Qwen2.5-7B-Instruct-Q4_K_M.gguf`; any GGUF of that model at that
quant works. `lms load` has no flags for flash attention or KV cache
quantization, so those are listed as settings to apply in the LM Studio app.
With `--json`, the config object gains `runtime`, and for llama.cpp and
LM Studio also `runtime_command`, `runtime_settings` and `runtime_notes`.

### scan

Prints only the detected hardware (the first block above). Accepts the global
flags, including `--json`.

### catalog

Lists every model variant in the catalog with its parameter count,
quantization, download size, maximum context and quality score. Use
`--catalog path/to/models.json` on any command to use your own catalog file.

### bench

Measures real tokens per second for the recommended models through a local
Ollama server. This is the only command that downloads data. It shows the
total size and asks before pulling anything.

```
ollama-compat bench [flags]

Flags:
      --allow-remote          allow benchmarking a non-local OLLAMA_HOST
      --installed-only        benchmark only models already installed; downloads nothing
      --keep                  keep models downloaded for the benchmark
      --max-download string   refuse if the total download exceeds this size, for example 20GB
      --runs int              timed runs per model (the median is reported) (default 3)
      --yes                   skip the download confirmation (for scripts)
```

`bench` also accepts `--ctx`, `--min-tps`, `--top` and the global flags
(`--catalog`, `--json`, `--no-color`). It talks to `http://127.0.0.1:11434`,
or to `OLLAMA_HOST` if set. If `OLLAMA_HOST` is not a loopback address,
`bench` refuses unless you pass `--allow-remote`; the measurements then
describe the remote machine, not the one you scanned.

Consent and cleanup:

- Nothing is downloaded without showing the total size and asking
  `[y/N]` (default No). Non-interactive runs refuse to download unless you pass `--yes`.
- Afterwards, on a terminal, it asks "Remove the N models downloaded for this
  benchmark ... [Y/n]" and deletes only models it pulled in that run. Models
  you already had are never removed.
- With `--yes`, or when there is no terminal, it never deletes anything and
  prints the `ollama rm <tag>` commands instead.
- `--keep` skips the question and keeps the downloaded models.
- Ctrl+C at the download prompt downloads nothing; at the cleanup prompt nothing is deleted and the downloaded models are listed with `ollama rm` commands.

Examples:

```bash
ollama-compat bench --installed-only          # measure what you already have
ollama-compat bench --max-download 10GB       # refuse if more than 10 GB is needed
ollama-compat bench --json > bench.json       # keep the results
ollama-compat bench --installed-only --share  # save a calibration record to contribute
```

## How estimates work

Token generation (decode) is limited by memory bandwidth: every generated
token reads the active weights once. The tool computes:

1. **Memory needed** = model file size + KV cache for your context length + 0.5 GB runtime overhead.
2. **GPU fraction** = how much of that fits in usable VRAM (VRAM minus 10 percent headroom), rounded to whole layers. On Apple Silicon the GPU shares system RAM, so usable memory is about 66 percent of RAM (75 percent above 36 GB).
3. **Speed** = 1 / (time on GPU + time on CPU), where each time is the bytes read per token divided by memory bandwidth times an efficiency factor. GPU factors are fitted from shared bench records (see [Contributing calibration data](#contributing-calibration-data)); a backend with no records uses 0.65, and CPU uses 0.55. For mixture-of-experts models only the active parameters count.

Bandwidth comes from lookup tables of known GPUs and chips, not from a
measurement. Models that cannot fit in VRAM plus 80 percent of RAM are
excluded.

Estimates are a rough guide (target within about plus or minus 30 percent,
not yet met on every machine). Apple Silicon is fitted from bench records
(only an 8 GB M1 so far, where the estimate lands within a few percent);
NVIDIA and AMD have no records yet. Offloaded models, unknown GPUs (marked
low confidence) and unusual drivers can be further off. Run
`ollama-compat bench` to measure the real
number on your machine. See `catalog/README.md` for the calibration log.

Each speed carries its expected error, for example `~20 tok/s ±25%`:

| Case | Range |
|---|---|
| Fully on GPU, backend with bench records (Apple Silicon and NVIDIA today) | the fitted error: at least ±25% until 3 machines have shared data, then at least ±15% |
| Fully on GPU, backend without records, or CPU only | ±30% |
| Partial offload (part GPU, part CPU) | ±40% |
| Unknown GPU (bandwidth guessed, marked low confidence) | ±50% |

With `--json`, each estimate has `est_tok_s_low`, `est_tok_s_high` and
`error_pct`.

When twice the memory that limits you (unified memory on a Mac, VRAM on a
discrete GPU, RAM without a GPU) would unlock a better Largest pick, a line
under the picks says so, for example
`With 24 GB VRAM, Largest would be gemma2:27b-instruct-q4_K_M (~20 tok/s ±50%, needs 18.1 GB).`
It also appears when the current Largest pick would run fully on GPU at
least 25 percent faster. Bandwidth is kept the same, so on a bandwidth-bound
machine (such as an 8 GB M1) more memory often unlocks nothing usable and
the line is left out. `--json` reports it as `near_miss`.

Quality scores in the catalog are maintainer-curated estimates, not benchmark
results; see `catalog/README.md`.

## Privacy

A default run (`recommend`, `scan`, `catalog`) makes no network calls. It reads
local hardware information and a catalog compiled into the binary. `bench`
only talks to your local Ollama server and asks before downloading anything.
Nothing is uploaded anywhere.

## Supported hardware

| Hardware | Backend | Notes |
|---|---|---|
| NVIDIA | CUDA | Compute capability 5.0 or newer, detected through `nvidia-smi` |
| AMD | ROCm | Only ROCm-supported gfx targets are treated as GPU |
| Apple Silicon | Metal | Unified memory; M1 through current chips |
| Intel GPUs, unsupported AMD GPUs | none | Estimated as CPU; pass `--vulkan` to estimate on Ollama's experimental Vulkan backend (needs `OLLAMA_VULKAN=1`) |
| CPU only | none | Estimated from system RAM bandwidth |

Platforms: Linux, macOS and Windows, on amd64 and arm64. If a GPU is not in the
bandwidth table, the tool uses a heuristic by vendor and VRAM size and marks
the result low confidence.

## Contributing calibration data

Speed estimates get better with every shared measurement. To add yours:

1. Close memory-heavy apps, then run `ollama-compat bench --share` (add
   `--installed-only` to download nothing). It saves
   `ollama-compat-calibration.json`; use `--share=path.json` for another name.
2. Open a [calibration issue](https://github.com/shra26/ollama-compatibility-checker/issues/new?template=calibration.yml)
   and paste the file. It holds hardware model names, memory sizes and
   speeds; no hostname, username or paths. The tool never uploads anything.

Records are fitted into per-backend speed factors and error ranges that ship
in the next release. See [calibration/README.md](calibration/README.md) for
how the fit works.

Issues adding missing GPUs to the bandwidth table are welcome too.

## Regenerating the catalog

`internal/catalog/models.json` is generated from the hand-maintained
`catalog/sources.yaml`. The generator needs network access to
`registry.ollama.ai` and is a development tool only:

```bash
go run ./scripts/gen-catalog -in catalog/sources.yaml -out internal/catalog/models.json
```

See `catalog/README.md` for where each field comes from.

## Troubleshooting

Common problems (GPU not detected, Ollama not reachable, slow speeds, Windows
colors, macOS Gatekeeper) are covered in
[Getting started: Troubleshooting](docs/getting-started.md#troubleshooting).

## Development

```bash
go test ./...
go vet ./...
go build -o ollama-compat ./cmd/ollama-compat
```

Project layout:

| Path | Contents |
|---|---|
| `cmd/ollama-compat` | CLI commands (cobra) |
| `internal/hw` | Hardware detection per vendor and OS |
| `internal/catalog` | Embedded model catalog and loader |
| `internal/estimate` | Memory fit and tok/s estimator |
| `internal/recommend` | Picks and Ollama config generation |
| `internal/output` | Text and JSON rendering |
| `internal/bench` | Ollama API client and the bench flow |
| `scripts/gen-catalog` | Catalog generator (dev only) |
| `docs/` | Getting started guide, design spec and implementation plan |

Releases are built by GoReleaser when a `v*` tag is pushed:

```bash
git tag v0.1.0
git push origin v0.1.0
```

## License

MIT. See `LICENSE`.
