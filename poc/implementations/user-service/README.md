# User Service

User management microservice for the Tsubo TODO application.

## Overview

This service implements the User domain, providing user CRUD operations and simple user validation for other services.

## Implementation Details

### Technology Stack
- **Language**: Go 1.22
- **Storage**: In-memory (no external dependencies)
- **HTTP Framework**: Standard library `net/http`

### Contract Compliance

This implementation follows the contract defined in `/Users/shun/Workspace/tsubo/poc/contracts/user-service.object.yaml`.

### Key Features

1. **User Creation**: Creates users with UUIDv4 IDs and normalized email addresses
2. **User Retrieval**: Gets user by ID or lists all users
3. **User Validation**: Simple authentication endpoint for other services
4. **Email Uniqueness**: Ensures email addresses are unique (case-insensitive)
5. **Email Normalization**: Converts all emails to lowercase

### API Endpoints

All endpoints are prefixed with `/api/v1`:

- `POST /users` - Create a new user
- `GET /users/{id}` - Get user by ID
- `GET /users` - List all users
- `POST /users/validate` - Validate user ID

Additional:
- `GET /health` - Health check endpoint

### File Structure

```
user-service/
├── main.go              # Entry point and routing
├── handlers.go          # HTTP handlers
├── models.go            # Data models
├── storage.go           # In-memory storage implementation
├── go.mod               # Go module definition
├── Dockerfile           # Multi-stage Docker build
├── docker-compose.yml   # Docker Compose configuration
├── .dockerignore        # Docker build exclusions
└── README.md            # This file
```

## Running the Service

### Using Docker Compose (Recommended)

```bash
# Start the service
docker-compose up -d

# View logs
docker-compose logs -f

# Stop the service
docker-compose down
```

### Using Docker

```bash
# Build the image
docker build -t user-service .

# Run the container
docker run -p 8080:8080 user-service
```

### Local Development (Go installed)

```bash
# Install dependencies
go mod download

# Run the service
go run .
```

## Testing

### Create a User

```bash
curl -X POST http://localhost:8082/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com"
  }'
```

### Get User by ID

```bash
curl http://localhost:8082/api/v1/users/{user-id}
```

### List All Users

```bash
curl http://localhost:8082/api/v1/users
```

### Validate User

```bash
curl -X POST http://localhost:8082/api/v1/users/validate \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "{user-id}"
  }'
```

### Health Check

```bash
curl http://localhost:8082/health
```

### Run Contract Tests

```bash
./test-contract.sh
```

## Edge Cases Handled

1. **Duplicate Email**: Returns 409 Conflict
2. **Invalid Email Format**: Returns 400 Bad Request
3. **Empty Name**: Returns 400 Bad Request
4. **User Not Found**: Returns 404 Not Found
5. **Email Case Insensitivity**: Normalized to lowercase

## Tsubo Philosophy Alignment

This implementation follows Tsubo's core principles:

- **Docker First**: Everything runs in Docker
- **Contract-Driven**: Implements exactly what the contract specifies
- **No Questions During Implementation**: The contract is complete and unambiguous
- **Go Language**: Minimizes hallucination with simple, consistent patterns
- **Domain Boundary**: Only handles User domain, independent of TODO domain
