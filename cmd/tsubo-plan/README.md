# tsubo-plan

Tsubo Implementation Planning Tool - Analyzes contracts and generates implementation plans for AI-driven parallel development.

## Overview

`tsubo-plan` is a command-line tool that:
1. Parses Tsubo (application) and Object (service) contracts
2. Analyzes dependencies between services
3. Determines implementation order (waves)
4. Generates a JSON implementation plan for AI agents

## Installation

### Build from source

```bash
# From the project root
go build -o tsubo-plan ./cmd/tsubo-plan

# Or install to $GOPATH/bin
go install ./cmd/tsubo-plan
```

### Usage

```bash
tsubo-plan <path-to-tsubo.yaml>
```

Example:
```bash
tsubo-plan ./poc/contracts/tsubo-todo-app.tsubo.yaml
```

## What it does

### Step 0: Parse Tsubo file
Reads the `.tsubo.yaml` file to understand the application structure.

### Step 1: Verify context files
Checks for the presence of:
- `PHILOSOPHY.md` - Tsubo's philosophy
- `docs/DEVELOPMENT_PRINCIPLES.md` - Development principles
- `docs/WHY_GO.md` - Why Go language
- `docs/CONTRACT_DESIGN.md` - Contract design guide

### Step 2: Enumerate objects
Lists all objects (microservices) defined in the tsubo.

### Step 3: Analyze dependencies
Parses each `.object.yaml` file and extracts service dependencies (not database dependencies).

### Step 4: Determine implementation order
Groups objects into waves:
- **Wave 0**: Objects with no dependencies (can run in parallel)
- **Wave 1**: Objects with dependencies (run after Wave 0 completes)

### Step 5: Generate implementation plan
Creates a JSON file at `/tmp/tsubo-implementation-plan.json` containing:
- Context files for AI agents to read
- Waves with implementation order
- Object details (name, contract path, dependencies)

## Output

The generated plan is a JSON file with this structure:

```json
{
  "tsubo": "tsubo-todo-app",
  "tsubo_file": "./poc/contracts/tsubo-todo-app.tsubo.yaml",
  "contracts_dir": "poc/contracts",
  "project_root": ".",
  "context_files": [
    "PHILOSOPHY.md",
    "docs/DEVELOPMENT_PRINCIPLES.md",
    "docs/WHY_GO.md",
    "docs/CONTRACT_DESIGN.md"
  ],
  "waves": [
    {
      "wave": 0,
      "parallel": true,
      "objects": [
        {
          "name": "user-service",
          "contract": "poc/contracts/user-service.object.yaml",
          "dependencies": null
        }
      ]
    },
    {
      "wave": 1,
      "parallel": true,
      "objects": [
        {
          "name": "todo-service",
          "contract": "poc/contracts/todo-service.object.yaml",
          "dependencies": ["user-service"]
        }
      ]
    }
  ]
}
```

## Next Steps

After generating the plan:

1. **Review the plan**:
   ```bash
   cat /tmp/tsubo-implementation-plan.json | jq
   ```

2. **Start AI implementation**:
   Use the plan to guide AI agents in parallel implementation of services.

## Architecture

```
cmd/tsubo-plan/
├── main.go              # CLI entry point
internal/
├── parser/
│   ├── tsubo.go        # Tsubo file parser
│   └── object.go       # Object file parser
├── analyzer/
│   └── dependency.go   # Dependency analyzer
└── planner/
    ├── wave.go         # Wave generator
    └── plan.go         # Plan generator
pkg/
└── types/
    ├── tsubo.go        # Tsubo type definitions
    ├── object.go       # Object type definitions
    └── plan.go         # Plan type definitions
```

## Philosophy

This tool embodies the Tsubo philosophy:
- **Human defines "what"**: Contracts specify requirements
- **AI implements "how"**: AI agents use the plan to implement services
- **Parallel execution**: Independent services can be implemented simultaneously
- **Dependency awareness**: Services with dependencies are implemented in order

## License

MIT
