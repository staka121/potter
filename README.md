> [!NOTE]
> This project is designed based on the concept of "What design patterns are possible when AI-implemented code operates without any human intervention?"
>
> As of February 2026, I am critical of the stance of abandoning or refusing to understand code.
> This stems from my pride as an engineer, but also because engineers currently bear the ultimate responsibility for quality assurance.
> Paradoxically, when engineers can relinquish quality responsibility, we will truly be able to accept full AI automation.

# Potter Framework

> A microservices framework for AI-driven development
>
> **Potter (craftsman) creates Tsubo (pots) - AI implements each solid object inside the pot**

[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Status](https://img.shields.io/badge/status-proof%20of%20concept-green.svg)]()

**Language / 言語:**
- 🇺🇸 English (this file)
- 🇯🇵 [日本語](./README.ja.md)

## Overview

**Potter** is a microservices development framework designed to accelerate parallel implementation by AI (LLMs) and reduce hallucinations.

Potter (craftsman) creates pots (Tsubo), and AI implements each solid object (domain) inside the pot.

### The Pot Metaphor

```
   ┌─────────────────────────────────────┐
   │  Tsubo (Pot) = Entire Application   │  ← Humans decide
   │                                     │
   │  ┌──────────┐  ┌──────────┐         │
   │  │  TODO    │  │   User   │  ...    │  ← Solid Objects
   │  │ Contract │  │ Contract │         │     (Domains)
   │  │  ┌────┐  │  │  ┌────┐  │         │
   │  │  │Impl│  │  │  │Impl│  │         │  ← AI decides
   │  │  └────┘  │  │  └────┘  │         │
   │  └──────────┘  └──────────┘         │
   │       ↓              ↓              │
   │  todo-service   user-service        │  ← Microservices
   └─────────────────────────────────────┘
```

**Humans decide the pot shape (application) and solid objects (domains) to put in. AI creates the internal structure of each object.**

- **Pot**: Entire application (container)
- **Solid Objects (Domains)**: Concrete business concepts (tangible things)
- **Microservices**: Implementation of each solid object
- **Implementation Details**: Internal structure of objects (AI decides)

### Why Potter?

Modern AI-driven development faces challenges:
- AI loses context easily in large codebases
- Parallel development is difficult with monolithic implementations
- Without clear boundaries, AI generates inconsistent code

Potter solves these challenges with the concept of **"pot = context boundary"**.

**Potter's Philosophy:**
- Humans focus on **"what to do"** (Contract definitions, domain boundaries)
- AI focuses on **"how to implement"** (implementation details)
- **Potter (craftsman) creates pots, AI implements solid objects inside**
- **One pot (application) contains multiple solid objects (domains/microservices)**
- Each solid object is independent, achieving loose coupling

## Core Idea

```
Small services → AI can understand → Reduced hallucinations
     ↓
Clear contracts → Parallel implementation → Faster development
     ↓
Auto verification → Quality assurance → Reliable code
```

## Key Features

### 🎯 Standardized Service Definition
Declarative YAML format for microservice specifications that AI can easily understand.

### 🔄 Parallel Implementation Orchestration
Multiple AI agents implement services in parallel, considering dependencies.

### ✅ Automatic Verification & Testing
Automatically execute contract tests, type checking, and integration tests for quality assurance.

### 🚀 Fast Development Cycle
Implement microservices 3-5x faster than traditional methods.

### 🔄 Contract-Driven Migration
Detect contract changes and migrate only affected services — no full rebuild needed:
- **`potter migrate plan`**: Show what changed and what will be rebuilt (dry run)
- **`potter migrate apply`**: Apply changes with breaking-change warnings
- **`potter migrate history`**: Audit trail of all contract changes
- **`potter refactor`**: Regenerate services cleanly from current Contract (remove patchwork)

### ☸️ Multi-Environment Support (Docker Compose ⇄ Kubernetes)
Single Tsubo definition deploys to both local development and production:
- **Local**: Docker Compose + gateway-service (simple, fast)
- **Production**: Kubernetes + Ingress (scalable, resilient)
- Automatic K8s manifest generation with `potter deploy generate`

## Quick Start

### Installation

```bash
# Clone repository
git clone https://github.com/staka121/tsubo.git
cd tsubo

# Build Potter CLI
go build -o potter ./cmd/potter

# Or install
go install ./cmd/potter
```

### API Key Setup (for AI Auto-Implementation)

To use Potter's AI-driven implementation (`potter build` command), you need a Claude API key.

#### 1. Get API Key

1. Visit [Anthropic Console](https://console.anthropic.com/)
2. Create account or login
3. Create new API key in **API Keys** section
4. Copy the key (starts with `sk-ant-`)

#### 2. Set API Key

Set the API key using one of these methods:

**Method 1: Environment Variable (Recommended)**

```bash
# Temporary (current session only)
export ANTHROPIC_API_KEY=sk-ant-xxxxx

# Permanent (add to ~/.bashrc or ~/.zshrc)
echo 'export ANTHROPIC_API_KEY=sk-ant-xxxxx' >> ~/.bashrc
source ~/.bashrc
```

**Method 2: .env File (Per Project)**

```bash
# Create .env file in project root
echo "ANTHROPIC_API_KEY=sk-ant-xxxxx" > .env

# Load and run
source .env
potter build app.tsubo.yaml
```

#### 3. Verify

```bash
# Check if API key is set
echo $ANTHROPIC_API_KEY

# Test AI implementation
potter build ./poc/contracts/tsubo-todo-app.tsubo.yaml
```

#### Security Notes

⚠️ **Important**: API keys are sensitive information. Follow these guidelines:

- ✅ Add `.env` files to `.gitignore`
- ✅ Never hardcode API keys in source code
- ✅ Don't commit API keys to public repositories
- ✅ Disable unused keys in Anthropic Console
- ✅ Rotate API keys regularly

#### API Pricing

Claude API uses pay-as-you-go pricing. See [Anthropic Pricing](https://www.anthropic.com/pricing) for details.

- **Model**: claude-sonnet-4-5-20250929 (default)
- **Estimated Cost**: ~$0.50-2.00 for medium service (~1000 lines)
- **Concurrency Control**: Manage costs with `--concurrency` option

### AI-Driven Service Implementation (Fully Automated)

```bash
# 1. Create new service template
potter new user-service

# 2. Create/edit .tsubo.yaml file to add services
# (Example: see poc/contracts/tsubo-todo-app.tsubo.yaml)

# 3. AI-driven auto-implementation (default)
export ANTHROPIC_API_KEY=your-api-key
potter build ./poc/contracts/tsubo-todo-app.tsubo.yaml

# Limit parallel execution
potter build --concurrency 4 ./poc/contracts/tsubo-todo-app.tsubo.yaml

# Generate prompts only (for manual execution)
potter build --prompt-only ./poc/contracts/tsubo-todo-app.tsubo.yaml

# 4. Start services after implementation
potter run ./poc/contracts/tsubo-todo-app.tsubo.yaml -d

# 5. Run tests
potter verify ./poc/contracts/tsubo-todo-app.tsubo.yaml
```

### Run PoC (Tsubo TODO Application)

```bash
# Clone repository
git clone https://github.com/staka121/tsubo.git
cd tsubo

# Build Potter CLI
go build -o potter ./cmd/potter

# Generate implementation with AI
potter build ./poc/contracts/tsubo-todo-app.tsubo.yaml

# Start implemented services
potter run ./poc/contracts/tsubo-todo-app.tsubo.yaml -d

# Integration tests
potter verify ./poc/contracts/tsubo-todo-app.tsubo.yaml
```

**Included Domains (Solid Objects):**
- User Domain (user-service: port 8080)
- TODO Domain (todo-service: port 8081)

### Incremental Development (migrate & refactor)

Once services are built, use `migrate` and `refactor` for ongoing development:

```bash
# Check what changed since last migration
potter migrate plan ./poc/contracts/tsubo-todo-app.tsubo.yaml

# Example output:
# [+] notification-service (NEW)        → Will implement
# [~] user-service (MODIFIED - BREAKING) → Endpoint removed
# [*] Update infrastructure

# Apply changes (prompts for confirmation if breaking)
potter migrate apply ./poc/contracts/tsubo-todo-app.tsubo.yaml

# View migration history
potter migrate history ./poc/contracts/tsubo-todo-app.tsubo.yaml

# Regenerate a single service cleanly from its current Contract
potter refactor --service todo-service ./poc/contracts/tsubo-todo-app.tsubo.yaml

# Regenerate all services (full clean slate from Contract)
potter refactor ./poc/contracts/tsubo-todo-app.tsubo.yaml
```

**Why refactor?** Contract is the Single Source of Truth. When implementations accumulate patches and drift from the Contract, regenerating from scratch is the cleanest solution.

### Production Deployment (Kubernetes)

Potter provides seamless deployment to Kubernetes with automatic manifest generation and deployment.

#### Prerequisites

- Kubernetes cluster (local: minikube/kind, cloud: GKE/EKS/AKS)
- kubectl configured to access your cluster
- Docker images pushed to a registry (optional for local clusters)

#### Deployment Steps

**1. Generate Kubernetes Manifests**

```bash
# Basic generation (uses default namespace)
potter deploy generate ./poc/contracts/tsubo-todo-app.tsubo.yaml

# Production with custom configuration
potter deploy generate \
  --namespace production \
  --ingress-host api.example.com \
  --registry docker.io/your-org \
  --tag v1.0.0 \
  --replicas 3 \
  ./poc/contracts/tsubo-todo-app.tsubo.yaml
```

This generates:
- `k8s/namespace.yaml` - Kubernetes namespace
- `k8s/deployment-*.yaml` - Service deployments with health probes
- `k8s/service-*.yaml` - ClusterIP services for internal communication
- `k8s/ingress.yaml` - Ingress for external access (replaces gateway-service)

**2. Deploy to Cluster**

```bash
# Apply manifests with automatic rollout monitoring
potter deploy apply

# Or with custom options
potter deploy apply --namespace production --timeout 10m
```

The `apply` command will:
- ✓ Check kubectl availability
- ✓ Apply all manifests to the cluster
- ✓ Wait for deployment rollout completion
- ✓ Display pod and service status

**3. Verify Deployment**

```bash
# Check pods
kubectl get pods -n production

# Check services
kubectl get svc -n production

# Check ingress
kubectl get ingress -n production

# View logs
kubectl logs -f deployment/user-service -n production
```

**4. Access Your Services**

If using Ingress with a hostname:
```bash
# Add to /etc/hosts for local testing
echo "127.0.0.1 api.example.com" | sudo tee -a /etc/hosts

# Access via Ingress
curl http://api.example.com/api/v1/users
curl http://api.example.com/api/v1/todos
```

For production with real DNS:
```bash
# Services are accessible via your configured domain
curl https://api.example.com/api/v1/users
```

#### Complete Workflow Example

```bash
# 1. Define services (already done in PoC)
ls ./poc/contracts/
# → tsubo-todo-app.tsubo.yaml
# → user-service.object.yaml
# → todo-service.object.yaml

# 2. AI implements services
potter build ./poc/contracts/tsubo-todo-app.tsubo.yaml

# 3. Test locally with Docker Compose
potter run ./poc/contracts/tsubo-todo-app.tsubo.yaml
curl http://localhost:8080/api/v1/users

# 4. Generate K8s manifests for production
potter deploy generate \
  --namespace prod \
  --ingress-host api.prod.example.com \
  --registry gcr.io/my-project \
  --tag $(git rev-parse --short HEAD) \
  ./poc/contracts/tsubo-todo-app.tsubo.yaml

# 5. Deploy to production cluster
potter deploy apply --namespace prod

# 6. Verify deployment
kubectl get all -n prod

# Done! 🎉
```

#### Key Differences: Local vs Production

| Aspect | Local (Docker Compose) | Production (Kubernetes) |
|--------|----------------------|------------------------|
| **Gateway** | gateway-service (auto-started) | Ingress (nginx/traefik) |
| **Access** | localhost:8080 | api.example.com |
| **Scaling** | Manual | Auto-scaling (HPA) |
| **Monitoring** | Docker logs | K8s native (Prometheus/Grafana) |
| **Deployment** | `potter run` | `potter deploy apply` |

#### Best Practices

- **Image Registry**: Always use a registry for production (Docker Hub, GCR, ECR, ACR)
- **Versioning**: Tag images with git commit SHA or semantic version
- **Namespaces**: Use separate namespaces for dev/staging/production
- **Resource Limits**: Review and adjust resource requests/limits in generated manifests
- **Monitoring**: Set up health checks and observability tools
- **Secrets**: Use Kubernetes Secrets for sensitive data (not yet automated by Potter)

For detailed Kubernetes integration guide, see [KUBERNETES.md](./docs/KUBERNETES.md).

## Architecture

```
┌─────────────────────────────────────────┐
│    Contract Definitions (YAML)          │
│  - tsubo-todo-app.tsubo.yaml            │
│  - user-service.object.yaml             │
│  - todo-service.object.yaml             │
└────────────┬────────────────────────────┘
             │
             ▼
┌─────────────────────────────────────────┐
│          potter CLI (Go)                │
│  - Contract parsing                     │
│  - Dependency analysis                  │
│  - Wave generation (execution order)    │
│  - AI auto-implementation (Claude API)  │
│  - Verification, testing, startup       │
└────────────┬────────────────────────────┘
             │
             ▼ (Default: AI implementation)
┌──────────┬──────────┬──────────┬────────┐
│ Claude   │ Claude   │ Claude   │  ...   │
│ (Wave 0) │ (Wave 0) │ (Wave 1) │        │
│          │          │          │        │
│ user-    │ other-   │ todo-    │        │
│ service  │ service  │ service  │        │
└────┬─────┴────┬─────┴────┬─────┴────┬───┘
     │          │          │          │
     ▼          ▼          ▼          ▼
┌────────────────────────────────────────┐
│       Generated Services (Go)          │
│  - 100% Contract compliant             │
│  - Docker-ready                        │
│  - Tests included                      │
└────────────────────────────────────────┘
```

## Tech Stack

- **CLI Framework:** Go 1.22
  - potter: Unified command-line interface
  - Contract parsing & planning
  - Claude API integration
  - Type-safe YAML parsing
  - Dependency analysis (topological sort)
  - Single binary distribution

- **Generated Services:** Go 1.22 (recommended)
  - Simple and consistent code
  - Reduced hallucinations
  - Standard library focused
  - Future support for TypeScript, Python

- **Contract Definition:** YAML
  - `.tsubo.yaml`: Pot (application) definition
  - `.object.yaml`: Object (service) definition
  - Human and AI readable

- **Deployment:** Docker & Docker Compose
  - Docker First principle
  - Complete environment isolation
  - Reproducibility guarantee

### Why Go?

**Go's "same code regardless of who writes it" characteristic dramatically reduces AI hallucinations.**

See [WHY_GO.md](./docs/WHY_GO.md) for details.

## Project Status

**Current Status: ✅ Fully Automated Pipeline Complete**

- [x] **Establish core philosophy**
- [x] **Define service definition format** (Contract Design)
- [x] **Establish development principles** (Docker First & Questioning timing)
- [x] **Establish file formats** (.tsubo.yaml / .object.yaml)
- [x] **Complete PoC** (TODO Application)
  - [x] Design pot (entire application)
  - [x] User Domain (solid object 1)
    - [x] Contract definition
    - [x] **AI-driven Go implementation**
    - [x] Dockerization
    - [x] Tests (100% Contract compliant)
  - [x] TODO Domain (solid object 2)
    - [x] Contract definition
    - [x] **AI-driven Go implementation**
    - [x] Dockerization
    - [x] User domain integration
    - [x] Tests (100% Contract compliant)
  - [x] Orchestration with docker-compose
  - [x] Integration tests (verify inter-domain communication)
- [x] **Unified CLI complete**
  - [x] **potter** (Go) - All-in-one command-line tool
    - [x] `potter new` - Service template generation
    - [x] `potter build` - Contract parsing, AI implementation, prompt generation
    - [x] `potter verify` - Contract verification, test execution
    - [x] `potter run` - Service startup (Docker Compose)
    - [x] `potter migrate` - Contract change detection & incremental migration
    - [x] `potter refactor` - Clean regeneration from current Contract
    - [x] Automatic dependency analysis (topological sort)
    - [x] Automatic Wave (execution order) determination (multi-wave support)
    - [x] Claude API client implementation
    - [x] Concurrency control (`--concurrency`)
    - [x] Wave-based parallel execution
    - [x] Real-time progress display
    - [x] Error handling

### Completed Automation Pipeline

```
Contract Definition (Human)
   ↓
potter build (Auto-analysis + AI implementation) ← Full rebuild
   ↓
Microservice Implementation (100% Contract compliant)
   ↓
potter verify (Verification)
   ↓
potter run (Startup)
   ↓
[Contract changes over time]
   ↓
potter migrate plan  → Detect changes (breaking / non-breaking)
potter migrate apply → Rebuild only affected services ← Incremental!
   or
potter refactor      → Regenerate cleanly from Contract ← Clean slate!
```

### Next Milestones

- [ ] potter CLI enhancements
  - [x] Support complex dependency graphs (topological sort)
  - [ ] Implementation plan visualization
  - [x] Cycle detection (circular dependency detection)
  - [ ] Retry logic
  - [x] Partial re-execution (`potter migrate apply` / `potter refactor --service`)
  - [ ] Multiple model support
- [ ] Verification engine implementation
  - [ ] Automate Contract compliance checking
  - [ ] Performance testing
  - [ ] Security scanning
- [ ] Multi-language support
  - [ ] TypeScript service generation
  - [ ] Python service generation

## Documentation

**Language / 言語:**
- 🇺🇸 [English (Master)](./docs/) - AI implementation uses English docs
- 🇯🇵 [日本語](./docs/ja/) - Japanese documentation

### Core Philosophy
- [PHILOSOPHY.md](./docs/PHILOSOPHY.md) ([日本語](./docs/ja/PHILOSOPHY.md)) - Potter's core philosophy
- [DOMAIN_DESIGN.md](./docs/DOMAIN_DESIGN.md) ([日本語](./docs/ja/DOMAIN_DESIGN.md)) - Relationship between pots and solid objects

### Development Guide
- [DEVELOPMENT_PRINCIPLES.md](./docs/DEVELOPMENT_PRINCIPLES.md) ([日本語](./docs/ja/DEVELOPMENT_PRINCIPLES.md)) - Docker First & questioning timing
- [CONTRACT_DESIGN.md](./docs/CONTRACT_DESIGN.md) ([日本語](./docs/ja/CONTRACT_DESIGN.md)) - Contract format details
- [WHY_GO.md](./docs/WHY_GO.md) ([日本語](./docs/ja/WHY_GO.md)) - Why Go language
- [KUBERNETES.md](./docs/KUBERNETES.md) ([日本語](./docs/ja/KUBERNETES.md)) - Kubernetes integration and multi-environment deployment

### CLI Commands
- **potter** - Unified command-line interface
  - `potter new` - Service template generation
  - `potter build` - Contract parsing, AI implementation (full rebuild)
  - `potter verify` - Contract verification, test execution
  - `potter run` - Service startup (Docker Compose)
  - `potter deploy` - Kubernetes deployment tools
    - `potter deploy generate` - Generate K8s manifests with Ingress
    - `potter deploy apply` - Apply manifests to K8s cluster
  - `potter migrate` - Contract change detection and incremental migration
    - `potter migrate plan` - Show pending changes (dry run)
    - `potter migrate apply` - Apply changes (rebuilds only affected services)
    - `potter migrate history` - Show migration audit trail
  - `potter refactor` - Regenerate services cleanly from current Contract
    - `--service <name>` - Refactor a single service

## Contributing

Currently in PoC phase. Ideas and feedback are welcome.

## License

MIT License - see the [LICENSE](LICENSE) file for details

## Name Origin

**Potter** and **Tsubo (壺)** carry deep meaning:

> **The pot (Tsubo) is a container representing the entire application.**
>
> **Domains are solid objects placed inside the pot.**
>
> One pot contains multiple solid objects (domains).
> Each solid object becomes an independent microservice.
>
> Humans decide **which solid objects (domains) to put in the pot**
> and define each object's **interface (Contract)**.
> How the internal structure of objects works is **determined by AI**.
>
> **Potter (craftsman) creates the pot structure,
> AI implements the contents of each solid object.**
>
> **One Pot (Application) = Collection of Multiple Solid Objects (Domains/Microservices)**

**New Meaning of Encapsulation:**
- Traditional encapsulation: Hide internal implementation from outside
- Potter's encapsulation: **Hide implementation details from humans**, delegate to AI
- **Independence of solid objects**: Each domain (microservice) exists independently in the pot

A collection of solid objects (domains) in the pot creates a robust application.

## Development Principles

Potter is developed based on these principles:

### 🐳 Docker First
- All implementations run in virtual environments (Docker)
- Zero impact on local environment
- Reproducibility guarantee

### 🤐 Questioning Timing
- **Before implementation (Contract stage)**: Eliminate ambiguity, security confirmation
- **During implementation**: No questions, AI implements autonomously

### 📝 Contract is Everything
- Contract is the Single Source of Truth
- Humans define "what to do"
- AI decides "how to implement"

See [DEVELOPMENT_PRINCIPLES.md](./docs/DEVELOPMENT_PRINCIPLES.md) for details.

---

**Status:** ✅ **Contract-Driven Migration Complete**
**Version:** 0.6.0
**Latest Achievement:** `potter migrate` & `potter refactor` — incremental, Contract-driven development

**Implemented:**
- ✅ **potter CLI** - Unified command-line interface
  - `potter new` - Service template generation
  - `potter build` - Contract parsing, AI implementation (full rebuild)
  - `potter verify` - Contract verification, test execution
  - `potter run` - Service startup (Docker Compose)
  - `potter deploy` - Kubernetes deployment
    - `potter deploy generate` - K8s manifest generation with Ingress
  - `potter migrate` - Contract change detection & incremental migration
    - `potter migrate plan` - Dry-run: show what changed and what will be rebuilt
    - `potter migrate apply` - Apply: rebuild only affected services
    - `potter migrate history` - Audit trail of all contract changes
  - `potter refactor` - Regenerate services cleanly from current Contract
  - Concurrency control (`--concurrency`)
  - Multi-wave support via topological sort
  - Claude API integration
- ✅ **Multi-Environment Support**
  - Docker Compose (local development with gateway-service)
  - Kubernetes (production with Ingress)
  - Same Tsubo definition for both environments
  - Auto-generated K8s manifests (Deployment, Service, Ingress)
- ✅ Pot (entire application): tsubo-todo-app
- ✅ 2 solid objects (AI parallel implementation):
  - user-service (Wave 0) - User management
  - todo-service (Wave 1) - TODO management
- ✅ Inter-domain communication (service-to-service)
- ✅ 100% Contract compliant
- ✅ Complete integration tests

**Contract is the Single Source of Truth — Potter manages everything else.** 🎉
