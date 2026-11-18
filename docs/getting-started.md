# Getting started

This guide takes you from nothing installed to running a recommended model in
Ollama with the right settings. It covers macOS, Linux and Windows.

The steps are:

1. [Install Ollama](#1-install-ollama) (needed to run models, and for `bench`)
2. [Get ollama-compat](#2-get-ollama-compat)
3. [Run your first check](#3-run-your-first-check)
4. [Apply the recommended config](#4-apply-the-recommended-config)
5. [Optional: measure real speed with bench](#5-optional-measure-real-speed-with-bench)

If something goes wrong, see [Troubleshooting](#troubleshooting) at the end.

## 1. Install Ollama

`ollama-compat recommend`, `scan` and `catalog` work without Ollama. You need
Ollama to actually run the recommended model and to use `bench`.

| OS | How |
|---|---|
| macOS | Download the app from <https://ollama.com/download>, or `brew install ollama` |
| Linux | `curl -fsSL https://ollama.com/install.sh \| sh` (installs a systemd service) |
| Windows | Download and run the installer from <https://ollama.com/download> |

Check it is running:

```bash
ollama --version
curl http://127.0.0.1:11434/api/version
```

If the second command fails, start the server: open the Ollama app (macOS,
Windows), run `sudo systemctl start ollama` (Linux), or run `ollama serve` in
a separate terminal.

## 2. Get ollama-compat

### Option A: prebuilt binary

Quickest: `brew install --cask shra26/tap/ollama-compat` on macOS or Linux,
or `scoop bucket add shra26 https://github.com/shra26/scoop-bucket` then
`scoop install ollama-compat` on Windows. Otherwise:

Download the archive for your platform from
[GitHub Releases](https://github.com/shra26/ollama-compatibility-checker/releases):

| Your machine | Archive |
|---|---|
| Mac with Apple Silicon (M1 and later) | `ollama-compat_<version>_darwin_arm64.tar.gz` |
| Mac with Intel | `ollama-compat_<version>_darwin_amd64.tar.gz` |
| Linux PC or server | `ollama-compat_<version>_linux_amd64.tar.gz` |
| Linux on ARM (Raspberry Pi 64-bit, Graviton) | `ollama-compat_<version>_linux_arm64.tar.gz` |
| Windows PC | `ollama-compat_<version>_windows_amd64.zip` |
| Windows on ARM | `ollama-compat_<version>_windows_arm64.zip` |

Then unpack it and put the binary on your `PATH`.

macOS and Linux:

```bash
tar -xzf ollama-compat_*_darwin_arm64.tar.gz
sudo mv ollama-compat /usr/local/bin/
```

On macOS, Gatekeeper blocks unsigned downloaded binaries. Clear the
quarantine flag once:

```bash
xattr -d com.apple.quarantine /usr/local/bin/ollama-compat
```

Windows (PowerShell): unzip the archive, move `ollama-compat.exe` into a
folder such as `C:\Tools`, and add that folder to your `PATH`:

```powershell
Expand-Archive .\ollama-compat_*_windows_amd64.zip -DestinationPath C:\Tools
setx PATH "$env:PATH;C:\Tools"
```

Open a new terminal after `setx` so the new `PATH` is picked up. Windows
SmartScreen may warn about an unrecognized app; choose "More info", then
"Run anyway".

### Option B: build from source

Use this if you want the latest unreleased code. You
need [Go 1.24 or newer](https://go.dev/dl/) and git.

```bash
git clone https://github.com/shra26/ollama-compatibility-checker.git
cd ollama-compatibility-checker
go build -o ollama-compat ./cmd/ollama-compat
./ollama-compat --version
```

On Windows use `go build -o ollama-compat.exe ./cmd/ollama-compat` and run
`.\ollama-compat.exe`.

Or install straight into your Go bin directory (`~/go/bin` by default; make
sure it is on your `PATH`):

```bash
go install github.com/shra26/ollama-compatibility-checker/cmd/ollama-compat@latest
```

## 3. Run your first check

```bash
ollama-compat
```

This is the same as `ollama-compat recommend`. It makes no network calls and
finishes in about a second. You get three blocks:

- **Hardware**: what was detected. Check the GPU line first. If it says
  `No supported GPU` but you have one, see [Troubleshooting](#troubleshooting).
- **Recommendations**: three picks, each with runners-up marked `also`.
  - **Fastest**: the quickest model of reasonable quality that fits fully on your GPU.
  - **Balanced**: the best quality that still reaches 20 tok/s. This is the main recommendation.
  - **Largest**: the best quality that is still usable (at least 5 tok/s), even if part of it runs on the CPU.
- **Config for Balanced**: copy-paste commands for Ollama (next step).

Every speed is an estimate with its expected error, such as `~20 tok/s ±25%`. Read `100% GPU` as "fits
entirely in GPU memory"; a lower percentage means part of the model runs on
the CPU, which is much slower.

Common variations:

```bash
ollama-compat --ctx 16384      # you need a longer context (uses more memory)
ollama-compat --min-tps 30     # you want a faster Balanced pick
ollama-compat --top 5          # show more alternatives
ollama-compat scan             # hardware only
ollama-compat catalog          # every model the tool knows about
ollama-compat --json > out.json  # machine-readable output
```

## 4. Apply the recommended config

The "Config for Balanced" block has up to four parts. Do them in order.

### 4.1 Download the model

```bash
ollama pull qwen2.5:3b-instruct-q4_K_M
```

Use the exact tag from your own output; the one above is an example.

### 4.2 Set the environment variables (GPU machines only)

The block shows `OLLAMA_FLASH_ATTENTION=1` and, for most models,
`OLLAMA_KV_CACHE_TYPE=q8_0`. They reduce memory use so more of the model
fits on the GPU. They are settings of the Ollama server, so they apply to all
models, and the server must be restarted to pick them up.

macOS, Ollama app:

```bash
launchctl setenv OLLAMA_FLASH_ATTENTION 1
launchctl setenv OLLAMA_KV_CACHE_TYPE q8_0
```

Then quit Ollama from the menu bar and open it again. `launchctl setenv`
does not survive a reboot; repeat it after restarting the Mac.

macOS or Linux, running `ollama serve` yourself:

```bash
export OLLAMA_FLASH_ATTENTION=1
export OLLAMA_KV_CACHE_TYPE=q8_0
ollama serve
```

Linux, systemd service (the default after the install script):

```bash
sudo systemctl edit ollama.service
```

Add these lines in the editor, save, then restart:

```ini
[Service]
Environment="OLLAMA_FLASH_ATTENTION=1"
Environment="OLLAMA_KV_CACHE_TYPE=q8_0"
```

```bash
sudo systemctl daemon-reload
sudo systemctl restart ollama
```

Windows (persistent, PowerShell or cmd):

```powershell
setx OLLAMA_FLASH_ATTENTION 1
setx OLLAMA_KV_CACHE_TYPE q8_0
```

Then quit Ollama from the system tray and start it again from the Start
menu.

If the block has no `env` part, the model runs on the CPU and there is
nothing to set.

### 4.3 Create a tuned model from the Modelfile

The Modelfile fixes the context length (`num_ctx`) and, on GPU machines, puts
all layers on the GPU (`num_gpu`). Save the three lines from your output to a
file named `Modelfile`, for example:

```
FROM qwen2.5:3b-instruct-q4_K_M
PARAMETER num_ctx 8192
PARAMETER num_gpu 99
```

Then run the `ollama create` line from your output, in the same folder:

```bash
ollama create qwen2.5-3b-instruct-q4_K_M-tuned -f Modelfile
```

### 4.4 Run it

```bash
ollama run qwen2.5-3b-instruct-q4_K_M-tuned --verbose
```

`--verbose` prints the real `eval rate` (tokens per second) after each
answer, so you can compare it with the estimate.

## 5. Optional: measure real speed with bench

`bench` runs the recommended models through your local Ollama and reports
measured tokens per second next to the estimate. Ollama must be running.

Start with models you already have; this downloads nothing:

```bash
ollama-compat bench --installed-only
```

To benchmark the full shortlist, run `ollama-compat bench`. It lists each
model, shows the total download size and asks `[y/N]` before pulling
anything. Afterwards it offers to remove the models it downloaded. Limit the
download with `--max-download 10GB`. See the [README](../README.md#bench)
for every flag.

If you share the output of `ollama-compat bench --json`, it helps improve
the estimates for everyone; see
[Contributing calibration data](../README.md#contributing-calibration-data).

## Troubleshooting

**`No supported GPU` on an NVIDIA machine.** The tool needs `nvidia-smi`,
which ships with the NVIDIA driver. Run `nvidia-smi` yourself; if it is not
found or errors, install or update the driver. Cards older than compute
capability 5.0 are not used by Ollama.

**`No supported GPU` on an AMD machine (Linux).** Install ROCm so that
`rocm-smi` is available, or the tool falls back to `/sys` and may mark the
card unsupported. Only ROCm-supported cards (for example RX 6800 and newer,
RX 7000, RX 9000, Instinct MI) are treated as GPUs.

**Intel GPU or older AMD GPU shows `not used by Ollama`.** Ollama does not
use these through CUDA or ROCm, so the tool estimates them as CPU. Pass
`--vulkan` (for example `ollama-compat --vulkan`) to estimate them on
Ollama's experimental Vulkan backend instead; set `OLLAMA_VULKAN=1` as shown
in the config block, and run `bench` to measure real speed. Integrated GPUs
are assumed to share system RAM, so those estimates are rough.

**Speeds in practice are much lower than the estimate.** Estimates assume
the machine is otherwise idle. Close memory-heavy apps and try again, and use
`ollama run <model> --verbose` or `bench` to measure. On an M1 with 8 GB,
memory pressure from other apps once halved the measured speed (10.8 against
21 tok/s on an idle machine).

**A recommended model runs partly on the CPU.** Check with `ollama ps`: the
`PROCESSOR` column shows the GPU and CPU split. Lower the context with
`--ctx 4096`, pick a runner-up with a smaller size, or set the environment
variables from step 4.2.

**`cannot reach Ollama at http://127.0.0.1:11434`.** Ollama is not running.
Start it (step 1). If Ollama runs on another port or machine, set
`OLLAMA_HOST`, for example `OLLAMA_HOST=127.0.0.1:11500`.

**`bench` says `refusing to download without --yes`.** You ran it without
an interactive terminal (for example in a script). Pass `--yes` to allow the
download, or `--installed-only` to download nothing.

**`bench` refuses because `OLLAMA_HOST` is remote.** Measurements on another
machine do not describe the one you scanned. Pass `--allow-remote` if that
is what you want.

**Colors look like `←[1m` on Windows.** Use `--no-color`, or set
`NO_COLOR=1`. Windows Terminal and PowerShell 7 show colors correctly.

**macOS says the binary "cannot be opened".** Run
`xattr -d com.apple.quarantine <path-to-ollama-compat>` once.
