# Tsubo Contract Design

## Overview

Tsubo Contracts are **structured specifications** that define microservices with semantic information that AI can understand and humans can read.

## Contract Format

### File Extensions

- **`.tsubo.yaml`**: Application (pot) definition
- **`.object.yaml`**: Service (solid object) definition

### Basic Structure

```yaml
version: "1.0"

service:
  name: service-name
  description: Service description
  runtime:
    language: go
    version: "1.22"

api:
  type: rest
  port: 8080
  endpoints:
    - path: /resource
      method: GET
      description: Endpoint description
      response:
        type: object
        properties:
          field: type

dependencies:
  services:
    - name: other-service
      endpoint: http://other-service:8080

  databases:
    - type: postgres
      name: db_name
      schema:
        tables:
          - name: table_name
            columns:
              - name: column_name
                type: data_type
```

## Key Sections

### 1. Service Metadata

```yaml
service:
  name: user-service
  description: User management service
  runtime:
    language: go
    version: "1.22"
```

**Purpose:**
- Identifies the service
- Specifies implementation language
- Sets runtime requirements

### 2. API Definition

```yaml
api:
  type: rest
  port: 8080
  endpoints:
    - path: /users/{id}
      method: GET
      description: Get user by ID
      response:
        type: object
        properties:
          id:
            type: string
          name:
            type: string
          email:
            type: string
```

**Purpose:**
- Define HTTP endpoints
- Specify request/response schemas
- Document expected behavior

### 3. Dependencies

```yaml
dependencies:
  services:
    - name: auth-service
      endpoint: http://auth-service:8081
      description: Authentication and authorization

  databases:
    - type: postgres
      name: user_db
      schema:
        tables:
          - name: users
            columns:
              - name: id
                type: uuid
                primary_key: true
              - name: email
                type: varchar(255)
                required: true
```

**Purpose:**
- Declare service dependencies
- Define database schemas
- Enable dependency graph analysis

## Contract Principles

### 1. Semantic Richness

Contracts must include **why**, not just **what**:

```yaml
api:
  endpoints:
    - path: /users
      method: POST
      description: Create a new user
      semantics:
        intent: Register new user in the system
        behavior:
          success: Returns created user with generated ID
          edge_cases:
            - case: Email already exists
              response: 409 Conflict
              reason: Email must be unique across all users
```

### 2. Single Source of Truth

Contracts serve three purposes:
1. **For Humans**: Service specification and agreement
2. **For AI**: Implementation instructions and context
3. **For Tests**: Validation criteria

### 3. Language Agnostic

Contracts are defined in YAML and are language-independent:
- Can generate Go services
- Future support for TypeScript, Python
- Compatible with OpenAPI/Protobuf

## Example: Complete Contract

```yaml
version: "1.0"

service:
  name: user-service
  description: User management and authentication service
  runtime:
    language: go
    version: "1.22"

api:
  type: rest
  port: 8080
  endpoints:
    - path: /health
      method: GET
      description: Health check endpoint
      response:
        type: object
        properties:
          status:
            type: string
            example: "healthy"

    - path: /users
      method: POST
      description: Create a new user
      request:
        type: object
        properties:
          name:
            type: string
            required: true
          email:
            type: string
            required: true
      response:
        type: object
        properties:
          id:
            type: string
          name:
            type: string
          email:
            type: string

    - path: /users/{id}
      method: GET
      description: Get user by ID
      response:
        type: object
        properties:
          id:
            type: string
          name:
            type: string
          email:
            type: string

dependencies:
  services: []

  databases:
    - type: postgres
      name: user_db
      schema:
        tables:
          - name: users
            columns:
              - name: id
                type: uuid
                primary_key: true
              - name: name
                type: varchar(255)
                required: true
              - name: email
                type: varchar(255)
                required: true
              - name: created_at
                type: timestamp
                default: now()
```

## Best Practices

### 1. Be Explicit

❌ **Bad:**
```yaml
- path: /users
  method: POST
```

✅ **Good:**
```yaml
- path: /users
  method: POST
  description: Create a new user account
  request:
    type: object
    properties:
      name:
        type: string
        required: true
      email:
        type: string
        required: true
        pattern: "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"
  response:
    type: object
    properties:
      id: string
      name: string
      email: string
```

### 2. Include Edge Cases

```yaml
endpoints:
  - path: /users
    method: POST
    semantics:
      edge_cases:
        - case: Email already exists
          response: 409 Conflict
          reason: Email must be unique
        - case: Invalid email format
          response: 400 Bad Request
          reason: Email must match RFC 5322 format
```

### 3. Document Intent

```yaml
service:
  context:
    purpose: |
      User management service handles user registration,
      authentication, and profile management.
    responsibilities:
      - User CRUD operations
      - Email uniqueness validation
      - Password hashing and verification
    constraints:
      - Emails must be unique
      - Passwords must be hashed with bcrypt
      - User IDs are UUIDs
```

## Contract Validation

Contracts are automatically validated by `tsubo-plan`:

```bash
# Validate contract
potter build app.tsubo.yaml

# Contract validation checks:
# - YAML syntax
# - Required fields
# - Type consistency
# - Dependency resolution
```

## See Also

- [PHILOSOPHY.md](./PHILOSOPHY.md) - Core philosophy
- [DEVELOPMENT_PRINCIPLES.md](./DEVELOPMENT_PRINCIPLES.md) - Development rules
- [WHY_GO.md](./WHY_GO.md) - Why Go language

---

> "A good Contract makes implementation trivial. A great Contract makes AI implementation perfect."
