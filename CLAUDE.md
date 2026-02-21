# CLAUDE.md

This file provides context for Claude Code when working in the Potter repository.

## Project Overview

**Potter** is a Go CLI tool that orchestrates AI-driven microservice development. It reads declarative YAML contract definitions (`.tsubo.yaml` / `.object.yaml`) and uses the Claude API to automatically implement each microservice in parallel.

Key metaphor: Potter (craftsman) creates the pot structure, AI implements the solid objects inside.

## Build & Run

```bash
# Build the CLI
go build -o potter ./cmd/potter

# Install globally
go install ./cmd/potter

# Run tests
go test ./...
```

## CLI Commands

| Command | Description |
|---------|-------------|
| `potter new [service]` | Create a new service template |
| `potter build <tsubo-file>` | AI-driven implementation (full rebuild) |
| `potter build --prompt-only <tsubo-file>` | Generate prompts without calling AI |
| `potter build --concurrency N <tsubo-file>` | Limit parallel AI calls |
| `potter verify <tsubo-file>` | Run contract verification and tests |
| `potter run [-d] <tsubo-file>` | Start services via Docker Compose |
| `potter migrate plan <tsubo-file>` | Show pending contract changes (dry run) |
| `potter migrate apply <tsubo-file>` | Rebuild only changed services |
| `potter migrate history <tsubo-file>` | Show migration audit trail |
| `potter refactor [--service name] <tsubo-file>` | Regenerate services cleanly from Contract |
| `potter deploy generate <tsubo-file>` | Generate Kubernetes manifests |
| `potter deploy apply` | Apply K8s manifests to cluster |

## Repository Structure

```
cmd/potter/          # CLI entry points (main.go, build.go, run.go, ...)
internal/
  analyzer/          # Dependency analysis (topological sort)
  executor/          # Claude API client, prompt generation, runner
  parser/            # .tsubo.yaml / .object.yaml parsers
  planner/           # Wave generation (execution order planning)
pkg/
  diff/              # Contract diff for migrate
  k8s/               # Kubernetes manifest generators
  migration/         # Migration planner and executor
  state/             # State manager for tracking builds
  types/             # Shared types (Plan, Tsubo, Object, etc.)
docs/                # Philosophy, design, and development docs
poc/contracts/       # Example tsubo definitions and implementations
```

## Key Concepts

- **Tsubo** (`.tsubo.yaml`): Defines the entire application (pot) — lists all services and their dependencies.
- **Object** (`.object.yaml`): Defines a single microservice contract — endpoints, schemas, dependencies.
- **Wave**: Group of services that can be implemented in parallel (determined by topological sort of dependencies).
- **Contract**: Single source of truth. Humans define "what", AI decides "how".

## Environment

- **Language**: Go 1.22, no external dependencies except `gopkg.in/yaml.v3`
- **AI Model**: Claude API (requires `ANTHROPIC_API_KEY` env var)
- **Local runtime**: Docker + Docker Compose
- **Production**: Kubernetes

## Development Process

Always create a new branch from `main` and open a PR for every change.

```bash
# 1. Start from main
git checkout main
git pull origin main

# 2. Create a feature branch
git checkout -b feature/your-feature-name

# 3. Make changes, commit
git add <files>
git commit -m "your commit message"

# 4. Push and open a PR
git push -u origin feature/your-feature-name
gh pr create --base main
```

## Development Notes

- Generated services go into `<tsubo-dir>/implementations/<service-name>/`
- Prompts are saved to `/tmp/potter/<app-name>/<timestamp>/` when using `--prompt-only`
- State is tracked per tsubo file to enable incremental `migrate` / `refactor`
- Context files passed to AI agents: `docs/PHILOSOPHY.md`, `docs/DEVELOPMENT_PRINCIPLES.md`, `docs/WHY_GO.md`, `docs/CONTRACT_DESIGN.md`
