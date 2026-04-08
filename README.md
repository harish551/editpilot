# EditPilot

EditPilot is a CLI-first AI-assisted video editing orchestrator powered by FFmpeg.

## What this version is

This version focuses on proving the core workflow:

1. accept one or more local video inputs
2. accept a natural-language edit request
3. convert the request into a structured edit plan
4. validate the plan against real media
5. compile FFmpeg commands
6. render or dry-run the result

The goal is not to be a full NLE replacement. It is to establish a reliable prompt-to-plan-to-FFmpeg pipeline that can later power richer CLI and web experiences.

## Design

### CLI-first approach

EditPilot starts as a CLI instead of a web app because the hardest part is not UI — it is the editing pipeline itself:

- prompt interpretation
- edit plan generation
- plan validation
- FFmpeg command compilation
- safe execution

CLI-first keeps iteration fast and makes the core engine easier to test.

### Planner model

EditPilot does **not** execute arbitrary AI-generated shell commands.

Instead it uses this pipeline:

```text
Prompt -> Plan JSON -> Validation -> FFmpeg Builder -> Runner
```

This keeps the system inspectable and safer.

### AI behavior

Prompt handling supports two paths:

1. **AI planner path**
   - uses configured provider/model/API key
   - currently supports an OpenAI-style planning flow
2. **heuristic fallback path**
   - used when AI is not configured or AI planning fails
   - keeps the CLI usable without external model access

So `--prompt` is AI-capable, but not AI-dependent.

### Media-aware validation

Validation is not limited to checking files exist.

EditPilot uses `ffprobe` to validate:
- input contains a video stream
- trim ranges fit within media duration
- concat inputs are dimension-compatible
- output path does not overwrite an input
- operation parameters are structurally valid

### Render strategy

EditPilot currently uses two render paths:

1. **simple concat path**
   - for direct concat with no transforms
2. **filtered concat path**
   - for trim/resize/text/speed workflows that need `filter_complex`

This allows EditPilot to handle both simple and chained edit operations more correctly.

## Implemented features

### CLI commands

- `editpilot probe`
- `editpilot plan`
- `editpilot validate`
- `editpilot render`
- `editpilot run`
- `editpilot config show`
- `editpilot config init`
- `editpilot config set provider <provider>`
- `editpilot config set model <model>`
- `editpilot config set api-key <key>`

### Supported edit operations

- `trim`
- `resize`
- `concat`
- `mute`
- `text_overlay`
- `speed`

### Config support

Config can be managed through CLI and/or environment variables.

Default config file:

```bash
$HOME/.config/editpilot/config.env
```

Supported settings:

```bash
EDITPILOT_AI_PROVIDER
EDITPILOT_AI_MODEL
EDITPILOT_AI_API_KEY
```

Environment variables override file-based config.

## Build requirements

- Go
- `ffmpeg`
- `ffprobe`

## Build and install

```bash
make build
make test
make install
```

Default install path:

```bash
$HOME/.local/bin/editpilot
```

## Example usage

### Probe input media

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

### Validate a saved plan

```bash
editpilot validate --plan examples/sample-plan.json
```

### Render from an existing plan

```bash
editpilot render \
  --plan examples/sample-plan.json \
  --save-command command.sh
```

### Dry-run without rendering

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

### Merge clips into a vertical reel

```bash
editpilot run \
  --input a.mp4 \
  --input b.mp4 \
  --prompt 'merge these clips, make them vertical, and add title "Trip Recap"' \
  --output reel.mp4
```

### Speed up and mute a clip

```bash
editpilot run \
  --input clip.mp4 \
  --prompt 'make it 1.5x speed and mute audio' \
  --output fast.mp4
```

## Testing

```bash
make test
```

Current test coverage includes:
- planner tests
- validator tests
- FFmpeg builder tests
- end-to-end CLI dry-run test with generated fixture media

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

## Current limitations

The current version does not yet include:
- crop
- subtitles
- background music
- timeline editing model
- advanced multi-clip semantic targeting

