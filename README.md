# EditPilot

EditPilot is a CLI-first AI-assisted video editing orchestrator powered by FFmpeg.

## Vision

Turn natural-language editing instructions into structured edit plans and safe FFmpeg execution.

## v0 Goal

Build a strong CLI foundation before adding any web UI.

## Requirements

- Go
- `ffmpeg`
- `ffprobe`

## Build and install

```bash
make build
make test
make install
```

By default, install places the binary at:

```bash
$HOME/.local/bin/editpilot
```

If needed, add that to your PATH:

```bash
export PATH="$HOME/.local/bin:$PATH"
```

## Commands

- `editpilot probe`
- `editpilot plan`
- `editpilot validate`
- `editpilot render`
- `editpilot run`
- `editpilot config show`
- `editpilot config init`
- `editpilot config set model <model>`
- `editpilot config set provider <provider>`
- `editpilot config set api-key <key>`

## Implemented operations

- `trim`
- `resize`
- `concat`
- `mute`
- `text_overlay`
- `speed`

## AI Config

You can manage model settings through CLI now.

### Initialize config

```bash
editpilot config init
```

### Show current config

```bash
editpilot config show
```

### Set provider

```bash
editpilot config set provider openai
```

### Set model

```bash
editpilot config set model gpt-5.1-codex
```

### Set API key

```bash
editpilot config set api-key YOUR_KEY_HERE
```

Default config file:

```bash
$HOME/.config/editpilot/config.env
```

Environment variables still work and override file values:

```bash
EDITPILOT_AI_PROVIDER
EDITPILOT_AI_MODEL
EDITPILOT_AI_API_KEY
```

## Example usages

### Probe a video

```bash
editpilot probe --input clip.mp4
```

### Generate a plan only

```bash
editpilot plan \
  --input clip.mp4 \
  --prompt 'trim first 10 seconds and add title "Demo"' \
  --output out.mp4
```

### Validate an existing plan

```bash
editpilot validate --plan examples/sample-plan.json
```

### Render from a saved plan

```bash
editpilot render \
  --plan examples/sample-plan.json \
  --save-command command.sh
```

### Dry-run a prompt without rendering

```bash
editpilot run \
  --input clip.mp4 \
  --prompt 'make it 1.5x speed and add title "Fast Cut"' \
  --output fast.mp4 \
  --dry-run \
  --save-plan plan.json \
  --save-command command.sh
```

### Trim and title a clip

```bash
editpilot run \
  --input clip1.mp4 \
  --prompt 'trim first 10 seconds and add title "Demo"' \
  --output final.mp4
```

### Merge two clips into a vertical reel

```bash
editpilot run \
  --input a.mp4 \
  --input b.mp4 \
  --prompt 'merge these clips, make them vertical, and add title "Trip Recap"' \
  --output reel.mp4
```

### Speed up and mute audio

```bash
editpilot run \
  --input clip.mp4 \
  --prompt 'make it 1.5x speed and mute audio' \
  --output fast.mp4
```

## Helpful flags

- `--dry-run`
- `--save-plan plan.json`
- `--save-command command.sh`

## Make targets

```bash
make help
make build
make test
make fmt
make tidy
make install
make uninstall
make clean
```

## What is hardened now

- config-driven AI integration path
- heuristic fallback planner
- ffprobe-aware validation
- concat compatibility checks
- fixture-based end-to-end CLI test
- command/plan export support
- CLI config management for model/provider/API key

## Notes

- planner supports heuristic prompt parsing today
- AI provider path is config-driven and implemented for OpenAI-style chat planning
- media-aware validation is built in through ffprobe
- next big expansion area would be crop / bg music / subtitles / timeline model
