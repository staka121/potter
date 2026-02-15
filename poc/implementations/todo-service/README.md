# Todo Service

A simple TODO management service implemented following the Tsubo philosophy.

## Overview

This service provides CRUD operations for TODO items, with user authentication via the `X-User-ID` header. Each TODO is associated with a user, and users can only access their own TODOs.

## Implementation Details

### Technology Stack
- **Language**: Go 1.22
- **Storage**: In-memory (thread-safe with mutex)
- **Dependencies**: user-service for user validation

### Architecture

```
todo-service/
├── main.go          # Entry point, HTTP server setup
├── handlers.go      # HTTP request handlers
├── models.go        # Data models (Todo, requests, responses)
├── storage.go       # In-memory storage with thread-safe operations
├── user_client.go   # HTTP client for user-service communication
├── go.mod           # Go module definition
├── Dockerfile       # Multi-stage Docker build
└── docker-compose.yml  # Docker Compose configuration
```

### Key Features

1. **User Authentication**: Every request requires an `X-User-ID` header
2. **User Validation**: User existence is validated via user-service before creating TODOs
3. **Data Isolation**: Users can only see and modify their own TODOs
4. **Thread-Safe**: In-memory storage uses RWMutex for concurrent access
5. **Contract-Compliant**: Implements exactly what the contract specifies

### API Endpoints

All endpoints require the `X-User-ID` header.

- `POST /api/v1/todos` - Create a new TODO
- `GET /api/v1/todos` - List all TODOs (filterable by status)
- `GET /api/v1/todos/{id}` - Get a specific TODO
- `PATCH /api/v1/todos/{id}` - Update TODO status
- `DELETE /api/v1/todos/{id}` - Delete a TODO
- `GET /health` - Health check endpoint

### Service Communication

The service communicates with user-service via HTTP:
- **Endpoint**: `POST /api/v1/users/validate`
- **Purpose**: Validate user existence before creating TODOs
- **URL**: Configurable via `USER_SERVICE_URL` environment variable
- **Default**: `http://user-service:8080` (Docker network)

## Running the Service

### Prerequisites

1. Build and run user-service first:
```bash
cd ../user-service
docker build -t tsubo-user-service:latest .
```

2. Build and run todo-service:
```bash
cd ../todo-service
docker-compose up --build
```

The service will be available at:
- todo-service: http://localhost:8081
- user-service: http://localhost:8080

### Environment Variables

- `PORT` - Server port (default: 8080)
- `USER_SERVICE_URL` - User service URL (default: http://user-service:8080)

## Testing

### Create a User First

```bash
# Create a user in user-service
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com"
  }'

# Response will include the user ID
```

### Create a TODO

```bash
curl -X POST http://localhost:8081/api/v1/todos \
  -H "Content-Type: application/json" \
  -H "X-User-ID: <user-id-from-above>" \
  -d '{
    "title": "Buy groceries",
    "description": "Milk, bread, eggs"
  }'
```

### List TODOs

```bash
# All TODOs
curl -X GET http://localhost:8081/api/v1/todos \
  -H "X-User-ID: <user-id>"

# Filter by status
curl -X GET "http://localhost:8081/api/v1/todos?status=pending" \
  -H "X-User-ID: <user-id>"
```

### Update TODO Status

```bash
curl -X PATCH http://localhost:8081/api/v1/todos/<todo-id> \
  -H "Content-Type: application/json" \
  -H "X-User-ID: <user-id>" \
  -d '{
    "status": "completed"
  }'
```

### Delete a TODO

```bash
curl -X DELETE http://localhost:8081/api/v1/todos/<todo-id> \
  -H "X-User-ID: <user-id>"
```

## Design Decisions

### Following Tsubo Principles

1. **Docker First**: Everything runs in Docker containers
2. **Contract is Everything**: Implementation follows the contract exactly
3. **Go Language**: Minimal hallucination, explicit error handling
4. **No Over-Engineering**: Simple, direct implementation without unnecessary abstractions
5. **User Isolation**: Each user can only access their own TODOs

### Storage Implementation

- In-memory map with RWMutex for thread safety
- Physical deletion (not soft delete) as specified in contract
- UUIDv4 for ID generation
- Sorted by creation time (newest first)

### Error Handling

All error cases from the contract are implemented:
- Empty/whitespace-only titles
- Title/description length validation
- Invalid UUID format
- Invalid status values
- TODO not found
- User not found/invalid

### Service Dependency

- todo-service depends on user-service for user validation
- Uses HTTP client with 5-second timeout
- Validates user before creating TODOs
- Returns 401 Unauthorized for invalid users

## Contract Compliance

This implementation is 100% compliant with the contract specification at:
`/Users/shun/Workspace/tsubo/poc/contracts/todo-service.object.yaml`

All endpoints, error cases, validations, and behaviors are implemented exactly as specified.
