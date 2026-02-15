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
potter build (Auto-analysis + AI implementation) ← Fully automated!
   ↓
Microservice Implementation (100% Contract compliant)
   ↓
potter verify (Verification)
   ↓
potter run (Startup)
```

### Next Milestones

- [ ] potter CLI enhancements
  - [x] Support complex dependency graphs (topological sort)
  - [ ] Implementation plan visualization
  - [x] Cycle detection (circular dependency detection)
  - [ ] Retry logic
  - [ ] Partial re-execution
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

### CLI Commands
- **potter** - Unified command-line interface
  - `potter new` - Service template generation
  - `potter build` - Contract parsing, AI implementation
  - `potter verify` - Contract verification, test execution
  - `potter run` - Service startup

## Contributing

Currently in PoC phase. Ideas and feedback are welcome.

## License

MIT License (planned)

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

**Status:** ✅ **Unified CLI Complete**
**Version:** 0.5.0
**Latest Achievement:** Unified CLI implementation complete (all operations possible with `potter` command)

**Implemented:**
- ✅ **potter CLI** - Unified command-line interface
  - `potter new` - Service template generation
  - `potter build` - Contract parsing, AI implementation
  - `potter verify` - Contract verification, test execution
  - `potter run` - Service startup
  - Concurrency control (`--concurrency`)
  - Multi-wave support via topological sort
  - Claude API integration
  - Prompt generation feature
- ✅ Pot (entire application): tsubo-todo-app
- ✅ 2 solid objects (AI parallel implementation):
  - user-service (Wave 0) - User management
  - todo-service (Wave 1) - TODO management
- ✅ Inter-domain communication (service-to-service)
- ✅ Orchestration with Docker Compose
- ✅ 100% Contract compliant
- ✅ Complete integration tests

**Everything from service creation to implementation, verification, and startup is complete with a single `potter` command!** 🎉
