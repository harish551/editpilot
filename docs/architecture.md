# EditPilot Architecture

## Core Pipeline

Prompt -> Planner -> Plan JSON -> Validator -> FFmpeg Builder -> Runner -> Output

## Principles

- AI does not directly execute shell commands
- Only validated operations are compiled into FFmpeg arguments
- Rendering should be deterministic and inspectable
- CLI comes first, web UI later
- AI provider/model must be config-driven, not hardcoded into command logic

## Initial Modules

- `internal/media`: ffprobe integration
- `internal/planner`: prompt to plan generation
- `internal/validator`: plan safety checks
- `internal/ffmpeg`: FFmpeg argument generation
- `internal/runner`: FFmpeg execution
- `internal/cli`: Cobra commands
- `internal/config`: env-driven configuration
- `internal/ai`: model/provider abstraction scaffold and planner service

## Current State

Implemented:
- probe command
- plan command
- validate command
- render command
- run command
- heuristic prompt parsing for trim/resize/concat
- FFmpeg arg generation for trim/resize/concat/mute
- filtered concat pipeline when concat is combined with transforms
- tests for planner, validator, and ffmpeg builder
- config-driven AI planning entry path with heuristic fallback

## Render Strategy

There are currently two concat paths:

1. **Simple concat demuxer path**
   - used when inputs are directly concatenated with no per-input transforms
   - fast and simple

2. **Filtered concat path**
   - used when concat is combined with transforms like trim/resize
   - builds a `filter_complex` graph with per-input transforms before concat

This makes EditPilot more correct for real chained edit requests.

## Next

1. support text overlay and speed changes in builder
2. add richer media-aware validation using ffprobe
3. integrate real provider calls in `internal/ai/client.go`
4. add end-to-end CLI tests with fixture media
5. improve prompt parsing for multi-input targeted edits
