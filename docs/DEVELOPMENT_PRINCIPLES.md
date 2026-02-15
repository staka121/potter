# Tsubo Development Principles

## 🎯 Core Principles

### 1. Docker First (Virtual Environment Required)

**All implementations must run in a virtual environment (Docker).**

#### Rationale
- ✅ **Zero impact on local environment**: Dependencies, ports, processes don't affect host
- ✅ **Reproducibility guarantee**: Same behavior in any environment
- ✅ **Dependency isolation**: Lock library and tool versions
- ✅ **Easy cleanup**: Complete removal with `docker-compose down`

#### Implementation Rules
```bash
# All services start with Docker Compose
docker-compose up -d

# Development, testing, execution all within Docker
docker-compose exec service-name <command>

# Complete cleanup on exit
docker-compose down
```

### 2. Questioning Timing

**Ask questions only before implementation (Contract phase), proceed autonomously during implementation.**

#### When Questions Are Allowed

**Before Implementation:**
- ✅ **Questions to eliminate Contract ambiguities**
  - Example: "When does this field become `null`?"
  - Example: "Behavior during concurrent execution?"
  - Example: "Is rollback needed on error?"

- ✅ **Security vulnerability indicators**
  - Example: "Does this endpoint require authentication?"
  - Example: "Password hashing?"
  - Example: "Risk of SQL injection"

- ✅ **Local environment impact confirmation**
  - Example: "Using port 8080, is this okay?"
  - Example: "Will pull new Docker image, is this acceptable?"

**During Implementation:**
- ❌ **No questions about implementation details**
  - Example: "Which pattern to use for this process?" → AI decides autonomously
  - Example: "How to handle errors?" → Implement according to Contract
  - Example: "How to structure files?" → Follow best practices

### 3. Contract is Everything

**Contract is the Single Source of Truth for implementation.**

#### What Should Be Included in Contracts

```yaml
service:
  context:
    purpose: |
      Purpose and business intent of this service
    responsibilities:
      - Specific responsibility 1
      - Specific responsibility 2
    constraints:
      - Constraint 1
      - Constraint 2

api:
  endpoints:
    - semantics:
        intent: Intent of this operation
        behavior:
          success: Behavior on success
          edge_cases:
            - case: Edge case description
              response: Expected response
              reason: Why it should be so
```

#### Handling Ambiguous Contracts

**Ask questions before implementation:**
- "What does `null` mean for this field?"
- "Behavior during concurrent execution?"
- "Conditions for returning this status code?"

**Don't guess during implementation:**
- Don't implement what's not written in Contract
- Avoid excessive generalization or abstraction
- Keep implementation minimal

## 🏗️ Development Flow

### Phase 1: Contract Definition (Human's Job)

1. **Define service boundaries**
   - Identify domains
   - Define responsibilities
   - Set constraints

2. **Write Contract**
   - API schema
   - Business context
   - Semantic information
   - Test scenarios

3. **Review Contract**
   - Eliminate ambiguities
   - Verify completeness
   - Check security considerations

### Phase 2: Implementation (AI's Job)

1. **Read Contract**
   - Understand purpose
   - Grasp constraints
   - Identify dependencies

2. **Design Structure**
   - File organization
   - Module division
   - Data flow

3. **Implement**
   - Follow Contract strictly
   - Apply best practices
   - Add appropriate tests

4. **Verify**
   - Contract compliance check
   - Run tests
   - Performance verification

### Phase 3: Review & Deploy (Human + AI)

1. **Human Review**
   - Verify Contract compliance
   - Check for security issues
   - Confirm business logic

2. **Integration Testing**
   - Service-to-service communication
   - End-to-end scenarios
   - Performance testing

3. **Deploy**
   - Docker Compose orchestration
   - Health checks
   - Monitoring

## 🐳 Docker First Detailed Rules

### File Structure

Every service must have:
```
service-name/
├── Dockerfile          # Service container definition
├── docker-compose.yml  # Orchestration configuration
├── main.go            # Entry point
├── go.mod             # Dependency management
└── test.sh            # Test execution script
```

### Dockerfile Best Practices

```dockerfile
# Multi-stage build
FROM golang:1.22 AS builder
WORKDIR /app
COPY go.* ./
RUN go mod download
COPY . .
RUN go build -o service

# Minimal runtime image
FROM gcr.io/distroless/base-debian12
COPY --from=builder /app/service /service
EXPOSE 8080
CMD ["/service"]
```

### Docker Compose Rules

```yaml
version: '3.8'
services:
  service-name:
    build: .
    ports:
      - "8080:8080"
    environment:
      - ENV=development
    networks:
      - tsubo-network

networks:
  tsubo-network:
    driver: bridge
```

## 📋 Contract-Driven Development

### Contract as Prompt Context

Contracts serve as:
1. **For Humans**: Service specification
2. **For AI**: Implementation instructions
3. **For Tests**: Validation criteria

### Contract Compliance Verification

```bash
# Automatic Contract compliance check
potter verify

# Check specific service
potter verify --service user-service
```

### Test-Driven by Contract

```yaml
# Contract includes test scenarios
tests:
  - name: Create user success
    given: Valid user data
    when: POST /users
    then: 201 Created

  - name: Create user duplicate email
    given: Existing email
    when: POST /users
    then: 409 Conflict
```

## 🚫 What NOT to Do

### Don't Ask During Implementation

❌ **Bad:**
- "Should I use repository pattern?"
- "Which error handling approach?"
- "Add logging here?"

✅ **Good:**
- Implement according to Contract
- Follow Go best practices
- Add tests as defined in Contract

### Don't Modify Local Environment

❌ **Bad:**
- Install dependencies on host
- Use host's database
- Modify host's network settings

✅ **Good:**
- Everything in Docker
- Use container's database
- Network isolation with Docker networks

### Don't Over-Engineer

❌ **Bad:**
- Add features not in Contract
- Create complex abstractions
- Anticipate future requirements

✅ **Good:**
- Implement exactly what Contract defines
- Keep it simple
- YAGNI (You Aren't Gonna Need It)

## 📊 Success Metrics

### Code Quality
- ✅ 100% Contract compliance
- ✅ All tests passing
- ✅ No security vulnerabilities

### Development Speed
- ✅ Average implementation time: 2-4 hours per service
- ✅ Parallel implementation: 3-5 services simultaneously
- ✅ First-time success rate: >80%

### Maintainability
- ✅ Code consistency across services
- ✅ Easy to understand and modify
- ✅ Self-documenting through Contracts

---

> "Follow principles strictly, automate ruthlessly, ship quickly."
