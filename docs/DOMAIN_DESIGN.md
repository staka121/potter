# Tsubo Domain Design

## The Pot and Solid Objects Metaphor

### Core Concept

```
┌─────────────────────────────────────────┐
│   Tsubo (Pot) = Application             │  ← Humans define
│                                         │
│  ┌──────────┐  ┌──────────┐            │
│  │  TODO    │  │   User   │   ...      │  ← Solid Objects
│  │  Domain  │  │  Domain  │            │     (Microservices)
│  └──────────┘  └──────────┘            │
│                                         │
└─────────────────────────────────────────┘

One Pot (Application) = Multiple Solid Objects (Domains/Microservices)
```

### Physical Analogy

**The Pot (Tsubo):**
- Container representing the entire application
- Humans define its shape (boundaries, interfaces)
- Holds multiple solid objects

**Solid Objects (Domains):**
- Tangible business concepts
- Each becomes one microservice
- Independent and self-contained
- AI implements their internals

## Domain Identification

### What is a Domain?

A domain is a **tangible business concept** with:
- ✅ **Clear boundaries**: Well-defined responsibilities
- ✅ **Ubiquitous language**: Domain-specific terminology
- ✅ **Independence**: Can change without affecting others
- ✅ **Cohesion**: Related concepts grouped together

### Examples

#### Good Domain Separation

```
✅ TODO Application (Pot)
├─ User Domain → user-service
│  - User registration
│  - User authentication
│  - User profile management
│
├─ TODO Domain → todo-service
│  - TODO creation
│  - TODO management
│  - TODO status tracking
│
└─ Notification Domain → notification-service
   - Email notifications
   - Push notifications
   - Notification preferences
```

Each domain:
- Has clear business purpose
- Uses domain-specific language
- Is independently deployable
- Maintains its own data

#### Bad Domain Separation

```
❌ Mixed Responsibilities
└─ app-service (Everything in one service)
   - User management + TODO management + Notifications
   → Multiple domains mixed, no clear boundaries
```

## Domain Size Guidelines

### Too Small (Over-fragmentation)

```
❌ Excessive Division
├─ user-creation-service
├─ user-authentication-service
├─ user-profile-service
└─ user-deletion-service

Problems:
- Domain logic scattered
- Increased network overhead
- Complex orchestration
- Hard for AI to understand the big picture
```

### Appropriate Size

```
✅ Balanced Domain
└─ user-service
   ├─ User creation
   ├─ User authentication
   ├─ User profile management
   └─ User deletion

Benefits:
- Clear business boundary
- Domain independence
- AI can understand at once
- Single deployment unit
```

### Too Large (Multiple Domains)

```
❌ Multiple Domains Mixed
└─ application-service
   ├─ User management
   ├─ TODO management
   ├─ Notifications
   └─ Analytics

Problems:
- Unclear boundaries
- AI loses context
- Increased hallucinations
- Tight coupling
```

## Identifying Domain Boundaries

### 1. Ubiquitous Language Test

Does the domain have its own vocabulary?

```
✅ User Domain:
- User, Registration, Authentication, Profile
- Login, Logout, Password Reset

✅ TODO Domain:
- TODO, Task, Completion, Status
- Due Date, Priority, Assignment
```

### 2. Independence Test

Can this domain change without affecting others?

```
✅ Independent:
- Changing TODO status logic doesn't affect User domain
- Adding notification types doesn't affect TODO domain

❌ Dependent:
- Changing user authentication affects all domains
- → Indicates tight coupling, needs restructuring
```

### 3. Cohesion Test

Are related concepts grouped together?

```
✅ High Cohesion (User Domain):
- User registration
- User authentication
- User profile
→ All related to "User" concept

❌ Low Cohesion:
- User registration
- Email sending
- Payment processing
→ Unrelated concepts mixed
```

### 4. Data Ownership Test

Does this domain own specific data?

```
✅ Clear Ownership:
- User domain owns user data
- TODO domain owns todo data
- Each has its own database

❌ Shared Ownership:
- Multiple services modifying same user table
- → Indicates poor domain separation
```

## Domain Relationships

### 1. Independent Domains (No Dependencies)

```
User Domain  ←→  Independent
TODO Domain  ←→  Independent
```

**Characteristic:**
- Can be implemented in parallel
- No coordination needed
- Wave 0 in implementation plan

### 2. Dependent Domains

```
User Domain  →  TODO Domain
    (Wave 0)        (Wave 1)
```

**Characteristic:**
- TODO depends on User
- Must implement User first
- Sequential implementation

### 3. Circular Dependencies (Anti-pattern)

```
❌ User Domain  ⇄  TODO Domain

This indicates:
- Poor domain separation
- Needs redesign
- Tsubo will detect and warn
```

## Domain-to-Service Mapping

### One Domain = One Service

**Principle:**
```
Domain  →  Microservice  →  Docker Container

User Domain → user-service → user-service:8080
TODO Domain → todo-service → todo-service:8081
```

**Benefits:**
- Clear 1:1 mapping
- Easy to understand
- Simple deployment
- AI-friendly scope

### Contract per Domain

Each domain has one Contract (`.object.yaml`):

```yaml
# user-service.object.yaml
service:
  name: user-service
  description: User domain management

# Defines everything about User domain:
# - API endpoints
# - Data models
# - Business rules
# - Dependencies
```

## Example: TODO Application Design

### Application (Pot)

```yaml
# tsubo-todo-app.tsubo.yaml
tsubo:
  name: tsubo-todo-app
  description: TODO management application

objects:
  - name: user-service
    contract: ./user-service.object.yaml

  - name: todo-service
    contract: ./todo-service.object.yaml
    dependencies:
      - user-service
```

### Domains (Solid Objects)

**User Domain:**
- **Purpose**: Manage users
- **Responsibilities**: Registration, authentication, profiles
- **Data**: Users table
- **API**: `/users/*` endpoints

**TODO Domain:**
- **Purpose**: Manage tasks
- **Responsibilities**: Create, update, complete TODOs
- **Data**: TODOs table
- **API**: `/todos/*` endpoints
- **Dependencies**: User service (to verify user ownership)

### Dependency Graph

```
Wave 0: user-service (no dependencies)
Wave 1: todo-service (depends on user-service)
```

## Best Practices

### 1. Start with Domains, Not Services

❌ **Bad:** "Let's create a REST API service"
✅ **Good:** "We need a User domain for managing users"

### 2. Use Business Language

❌ **Bad:** data-service, api-service, controller-service
✅ **Good:** user-service, todo-service, notification-service

### 3. One Database per Domain

```
✅ Good:
├─ user-service → user_db
└─ todo-service → todo_db

❌ Bad:
├─ user-service ─┐
└─ todo-service ─┴→ shared_db
```

### 4. Define Clear Interfaces

```yaml
# User domain interface
api:
  - GET /users/{id}
  - POST /users
  - PUT /users/{id}

# TODO domain interface
api:
  - GET /todos
  - POST /todos
  - PUT /todos/{id}
```

## Anti-Patterns

### 1. God Service

❌ One service doing everything
✅ Multiple focused domains

### 2. Chatty Services

❌ Too many inter-service calls
✅ Proper domain boundaries reduce chattiness

### 3. Data Coupling

❌ Services sharing database tables
✅ Each service owns its data

### 4. Circular Dependencies

❌ Services depending on each other
✅ Unidirectional dependencies only

## Validation

Tsubo automatically validates domain design:

```bash
potter build app.tsubo.yaml

# Checks:
# - Circular dependencies (error)
# - Missing dependencies (error)
# - Duplicate service names (error)
# - Optimal domain size (warning)
```

## See Also

- [PHILOSOPHY.md](./PHILOSOPHY.md) - The pot metaphor
- [CONTRACT_DESIGN.md](./CONTRACT_DESIGN.md) - How to define domains
- [DEVELOPMENT_PRINCIPLES.md](./DEVELOPMENT_PRINCIPLES.md) - Development workflow

---

> "Good domains are like LEGO blocks: independent, composable, and fit together perfectly."
