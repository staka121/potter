# Potter Framework Design Philosophy

## The Philosophy of Potter and Tsubo

### Origin of the Name

**Potter** is the craftsman who creates pots (Tsubo).
**Tsubo (壺)** represents the pot - a container for the entire application.

In the Potter framework, the craftsman (Potter) creates the pot (Tsubo), and AI implements the solid objects inside it.

```
      ┌─────────────────────────────────────────┐
      │   Tsubo (Pot) = Entire Application      │  ← Humans define the shape
      │                                         │
      │  ┌──────────┐  ┌──────────┐            │
      │  │  TODO    │  │   User   │   ...      │  ← Solid Objects
      │  │  Domain  │  │  Domain  │            │     (Microservices)
      │  │┌────────┐│  │┌────────┐│            │
      │  ││Contract││  ││Contract││            │  ← Contract (Shape)
      │  │└────────┘│  │└────────┘│            │
      │  │  ┌────┐ │  │  ┌────┐ │            │
      │  │  │Impl│ │  │  │Impl│ │            │  ← AI decides implementation
      │  │  └────┘ │  │  └────┘ │            │
      │  └──────────┘  └──────────┘            │
      │                                         │
      └─────────────────────────────────────────┘

      One Pot (Application) = Multiple Solid Objects (Domains/Microservices)
```

### New Meaning of Encapsulation

**Traditional Encapsulation (OOP):**
- Hide internal implementation from the outside
- Improve modularity through information hiding

**Potter's Encapsulation:**
- **Hide implementation details from humans**
- **Clearly separate what AI handles from what humans decide**
- Humans don't need to know how the contents of the pot interact
- Potter (craftsman) creates the pot, AI fills it with implementations

### AI-First Principle

**Human's Role:**
- ✅ **Define the shape of the pot** (interfaces, boundaries)
- ✅ **Decide what goes into the pot** (responsibilities, context)
- ✅ **Define what you see when you look into the pot** (expected behavior, I/O)
- ❌ No need to know **how the contents interact**

**AI's Role:**
- ✅ **Decide how the contents interact** (implementation details)
- ✅ Handle edge cases
- ✅ Error handling
- ✅ Performance optimization
- ❌ Leave interface and responsibility decisions **to humans**

### The Essence of Potter

> **Potter (craftsman) creates the pot (Tsubo) - the container representing the entire application.**
>
> **Domains are solid objects placed inside the pot.**
> **Each solid object becomes one microservice.**
>
> One pot contains multiple solid objects (domains).
> Humans (Potter) decide **the shape of this pot** (the overall application picture),
> and humans also decide **which objects (domains) to put in**.
>
> Each solid object (domain) is **tangible**, and humans create its concept.
> However, there's **no need to know** how the internal structure of each object works.
>
> Therefore, how these internal structures work is **determined by AI**.
>
> Humans should only care about **which objects (domains) to put in the pot**,
> and **what you see** when looking into each object—
> that is, **only the interface (Contract)**.
>
> **One Pot (Application) = Multiple Solid Objects (Domains/Microservices)**
>
> This ensures the independence of each domain and realizes a loosely-coupled system.

### Physical Imagery

```
Pot (Entire Application)
  ├─ Solid Object 1 = TODO Domain → todo-service
  ├─ Solid Object 2 = User Domain → user-service
  └─ Solid Object 3 = Auth Domain → auth-service

Each solid object is:
- Independent (doesn't mix with others)
- Tangible (concrete)
- Defined by Contract (shape)
- Implementation created by AI

Example: TODO Application
┌─────────────────────────────┐
│    Pot (TODO App)           │
│                             │
│  [TODO]  [User]  [Auth]     │ ← Solid Objects
│    ↓       ↓       ↓        │
│  todo-  user-  auth-        │
│  service service service    │
└─────────────────────────────┘
```

---

## Core Philosophy

### Why Create Potter?

**Challenges in Modern Software Development:**
- AI (LLMs) are powerful but prone to hallucinations in large codebases
- In monolithic implementations, AI cannot grasp the big picture, and context gets scattered
- Parallel development of multiple features is desired, but difficult with existing architectures
- Without clear boundaries and contracts, AI-generated code lacks consistency

**Potter's Approach:**
Use microservice boundaries as "context boundaries" for AI-driven development.
Potter (craftsman) creates pots (Tsubo), and AI fills them with implementations.
By treating each service as a small, clear, independent "pot":
- Split into scopes that AI can easily understand and implement
- Dramatically increase development speed through parallel implementation
- **Reduce hallucinations through Contract-Driven Testing (CDT)**
- Ensure quality through automated verification

**Important Concept:**
Even if microservices aren't the best architecture, **start with Contract definitions**.
Boundary definitions are always valuable.

Potter Contracts serve as **Single Source of Truth with three roles**:
1. **For Humans**: Agreement specification between services
2. **For AI**: Clear instruction on "what to do" (prompt context)
3. **For Testing**: Validation criteria

## Design Principles

### 0. AI First: Humans Concept, AI Implementation

**Potter's Most Fundamental Principle**: Clearly separate responsibilities between humans and AI.
Potter (craftsman/human) creates the pot structure, AI implements the contents.

**What Humans Decide (Tangible Things):**
- Shape of the pot (interfaces, boundaries)
- What to put in the pot (responsibilities, domain logic)
- What you see when looking into the pot (expected behavior, I/O)

**What AI Decides (Intangible Things):**
- How the contents of the pot interact (implementation details)
- Specific error handling methods
- Performance optimization
- Edge case handling

**Difference from Traditional Development:**

| Approach | Human's Role | AI's Role |
|----------|-------------|-----------|
| Traditional | Design + Implementation | Code completion, review |
| **Potter** | **Contract definition only** | **All implementation** |

**Rationale:**
- Humans can focus on "what should be done"
- AI can focus on "how to implement it"
- Division of labor reduces hallucinations (AI just follows clear instructions)
- Development speed dramatically increases

### 1. Contract is Everything

**Define Contracts before implementation. Contracts are "blueprints of the pot".**

Contracts are not just API definitions, but **define the shape and contents of the pot**:
- **Shape of the pot**: Interface (what to put in, what you see)
- **Contents of the pot**: Context (responsibilities, constraints, expected behavior)
- **How to use the pot**: Dependencies, performance requirements

Three roles of Contracts:
- **For Humans**: Blueprint of the pot, agreement specification between services
- **For AI**: Clear instruction on "how to create the contents of this pot" (prompt context)
- **For Testing**: Validation criteria for whether the pot functions correctly

What should be included:
- ✅ API schema (types, endpoints)
- ✅ **Business context** (purpose, responsibilities, domain)
- ✅ **Semantic information** (intent, behavior, edge cases)
- ✅ Dependencies (why dependencies exist)
- ✅ Performance requirements
- ✅ Constraints and invariants

**Rationale:**
- With clear contracts, multiple AI agents can implement in parallel while maintaining consistency
- Without semantic information ("why it should be so"), AI hallucinates
- One definition serves three purposes (DRY principle)

**Important:** OpenAPI or Protobuf alone is insufficient. Include not just type definitions, but **business purpose, behavioral intent, and expected behavior in edge cases**.

### 2. Boundary is Domain

**The pot is the boundary of the entire application.**
**Solid objects (domains) become independent microservices within it.**

**One Pot (Application) = Multiple Solid Objects (Domains/Microservices)**

Each solid object (domain) has:
- **Domain boundary**: Boundary of business concepts
- **Service boundary**: Implementation boundary (microservice)
- **Context boundary**: Scope that AI should understand

When these align:
- ✅ Independence of each domain is ensured
- ✅ Microservices become loosely coupled
- ✅ Fits within the scope AI can understand at once
- ✅ Each domain is independently testable and deployable

**Size of Solid Objects (Domains):**
```
Too small objects (excessive division):
  ❌ Domains become too fragmented, complexity increases overall
  ❌ Network overhead increases
  ❌ Domain logic gets scattered

Appropriately sized objects (1 domain = 1 service):
  ✅ Clear business concept boundaries
  ✅ Domain independence is ensured
  ✅ AI can understand at once
  ✅ Independently testable and deployable

Too large objects (multiple domains mixed):
  ❌ Domain boundaries are ambiguous
  ❌ AI loses context
  ❌ Hallucinations increase
```

**How to Identify Domains:**
- **Ubiquitous language**: Does it have domain-specific terminology?
- **Independence**: Can it change without depending on other domains?
- **Cohesion**: Are related concepts grouped together?
- **Boundary**: Is there a clear boundary of responsibilities?

**Example:**
```
❌ Bad example: Include TODO management in user-service
   → Multiple domains mixed in one solid object

✅ Good example (multiple solid objects in a pot):
   Pot (TODO Application)
   ├─ User solid (user-service)
   ├─ TODO solid (todo-service)
   └─ Auth solid (auth-service)
   → Each is an independent domain/microservice
```

**Important:** If you correctly define the boundaries of each solid object (domain), the independence and maintainability of the entire application improves.

### 3. Parallel by Default

Explicitly manage dependencies between services and implement in parallel whenever possible.
- Automatic analysis of dependency graphs
- Independent services implemented simultaneously by multiple AI agents
- Progress management by orchestrator
- Services under implementation don't interfere with each other

**Rationale:**
- Dramatic improvement in development speed (3-5x faster than traditional)
- Minimize developer wait time
- Efficient use of AI agents

### 4. Verify Continuously

Verify automatically while implementing. Especially **Contract-Driven Testing (CDT)** as the core.
- **Contract Testing**: Check conformance to Contract-defined specifications
- Type safety checks: Leverage Go/Rust type systems
- Integration tests: Verify interactions between services
- Performance tests: Confirm SLAs defined in contracts are met
- Security scans: Early detection of vulnerabilities

**Rationale:**
- Ensure quality of AI-generated code
- Early detection of bugs from hallucinations
- Prevent regressions
- Reduce human review burden

### 5. Go-First, Language-Agnostic Interface

**Contract definitions are language-agnostic, Go language recommended for implementation.**

- Contract definition: Language-agnostic (YAML)
- Compatible with OpenAPI/Protobuf
- **Recommended implementation language: Go**
- Future support for TypeScript, Python

**Why Go language:**
- ✅ **Same code regardless of who writes it** → Minimize AI hallucinations
- ✅ Simple language specification → Easy for AI to understand
- ✅ Explicit error handling → Prevent oversights
- ✅ Standard formatting (gofmt) → Unified code style
- ✅ Microservices ecosystem

**See [WHY_GO.md](./WHY_GO.md) for details.**

**Architecture Unification:**
- Orchestrator: Go
- Validator: Go (initially considered Rust, but unified to Go)
- Generated services: Go (recommended)
- CLI: Go

## What We're Building

### Core Components

#### 1. Potter Contract Format
Contract definition format including semantic information for defining microservices.

**See [CONTRACT_DESIGN.md](./CONTRACT_DESIGN.md) for details.**

Contracts are not just API definitions:
- Inherit the good parts of OpenAPI/Protobuf
- Describe business context, intent, and behavior
- Structured format that AI can easily understand
- Human-readable and understandable

```yaml
# Example: user-service.tsubo.yaml
service:
  name: user-service
  description: User management service

interface:
  api:
    - method: GET
      path: /users/{id}
      response: User
    - method: POST
      path: /users
      request: CreateUserRequest
      response: User

dependencies:
  - auth-service
  - database

types:
  User:
    id: string
    name: string
    email: string

tests:
  - name: User creation test
    given: Valid user data
    when: POST /users
    then: 201 Created
```

#### 2. Orchestrator
Manages multiple AI agents and coordinates parallel implementation.

**Features:**
- Parse service definitions
- Build dependency graphs
- Distribute tasks to AI agents
- Monitor implementation progress
- Automate integration and deployment

#### 3. Validator (Verification Engine)
Automatically verifies generated code.

**Features:**
- Check contract compliance
- Verify type safety
- Automatic test execution
- Performance measurement
- Security scanning

#### 4. Template System
Best practice templates for each language/framework.

**Features:**
- Generate initial service structure
- Auto-generate boilerplate code
- Embed best practices
- Support custom templates

#### 5. CLI Tool
Command-line tool for developers to operate Tsubo.

```bash
# Define a new service
potter new user-service

# Generate implementation from service definition (AI-driven)
potter build user-service

# Build all services in parallel
potter build --all --parallel

# Run verification
potter verify

# Start services
potter run
```

## Goals

### Short-term Goals (MVP)
- [x] Potter Contract Format specification
- [x] Contract validator (Go implementation)
- [x] Basic orchestrator (Go implementation)
- [x] AI prompt generation engine
- [x] Go service templates
- [x] CLI tool (`potter new`, `potter build`, `potter verify`)
- [x] Demo application (2 microservices)

### Mid-term Goals
- [ ] Support multiple AI providers (Claude, GPT-4, etc.)
- [ ] More language templates (TypeScript, Python, etc.)
- [ ] Web UI for orchestration
- [ ] Real-time monitoring dashboard
- [ ] Plugin system

### Long-term Goals
- [ ] De facto standard for AI-driven development
- [ ] Enterprise support (authentication, auditing, etc.)
- [ ] Cloud-native integration (K8s, Service Mesh, etc.)
- [ ] Ecosystem formation (community templates, etc.)

## Success Metrics

1. **Development Speed:** Implement microservices 3-5x faster than traditional methods
2. **Quality:** Reduce bug rate of AI-generated code by 50%+
3. **Parallelism:** Implement 5-10 services simultaneously on average
4. **Learning Curve:** New developers become productive within one day

## Non-Goals (What We Won't Do)

Tsubo does NOT aim for:

- ❌ **Runtime orchestration**: Not a replacement for Kubernetes, Service Mesh, or API Gateway
  - Tsubo is a **development-time tool**, not runtime infrastructure
  - Generated services run in any runtime environment

- ❌ **Complete automation**: Don't eliminate human review
  - AI-generated code must always be reviewed by humans
  - Contract definitions are also created and reviewed by humans
  - Tsubo accelerates development but doesn't replace developers

- ❌ **Monolithic application support**: Not applicable to codebases without boundaries
  - However, supports gradual migration from monolith to services

- ❌ **Support all languages/frameworks**: Initial focus on Go
  - Future expansion to TypeScript, Python, etc.
  - However, Contract definitions are language-agnostic
  - Go positioned as recommended language

- ❌ **Complete replacement of existing Contract formats**: Extend OpenAPI/Protobuf
  - Maintain compatibility with existing tools
  - Enable gradual adoption

---

> "Potter (craftsman) creates pots (Tsubo), each a small but complete space.
> A collection of well-crafted pots creates a beautiful garden.
> Similarly, a collection of well-defined microservices creates a robust system."
