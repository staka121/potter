# Todo Service Implementation Notes

## Implementation Summary

The todo-service has been implemented following the Tsubo philosophy and the contract specification at `/Users/shun/Workspace/tsubo/poc/contracts/todo-service.object.yaml`.

## Key Implementation Decisions

### 1. User Association (Not in Contract, Added per Instructions)

**Issue**: The contract does not specify a `user_id` field in the Todo type, but the instructions require:
- "Associate each TODO with a user ID"
- "Filter TODOs by user ID (users only see their own TODOs)"

**Solution**:
- Added `UserID` field to internal Todo struct with JSON tag `json:"-"` to exclude it from API responses
- This maintains contract compliance (API responses match contract exactly) while enabling user isolation
- User authentication via `X-User-ID` header (as per instructions)

### 2. User Validation Flow

1. Client sends request with `X-User-ID` header
2. Service validates header is present (returns 401 if missing)
3. For CREATE operations: Service calls user-service `/api/v1/users/validate` endpoint
4. Service validates user exists (returns 401 if invalid)
5. Service associates TODO with validated user ID
6. For READ/UPDATE/DELETE: Service checks TODO belongs to requesting user

### 3. Thread Safety

- In-memory storage uses `sync.RWMutex` for concurrent access
- Read operations use `RLock()` for parallel reads
- Write operations use `Lock()` for exclusive access
- All map access is protected

### 4. Service Communication

**User Service Integration:**
- HTTP client with 5-second timeout
- Endpoint: `POST /api/v1/users/validate`
- Request body: `{"user_id": "<user-id>"}`
- Response: `{"valid": true/false, "user": {...}}`
- Configurable via `USER_SERVICE_URL` environment variable

### 5. Error Handling

All error cases from contract are implemented:

| Error Case | HTTP Status | Error Message |
|-----------|-------------|---------------|
| Empty/whitespace title | 400 | "title cannot be empty" |
| Title > 200 chars | 400 | "title too long (max 200 characters)" |
| Description > 2000 chars | 400 | "description too long (max 2000 characters)" |
| Invalid status value | 400 | "invalid status value" |
| Invalid UUID format | 400 | "invalid id format" |
| TODO not found | 404 | "todo not found" |
| Missing X-User-ID | 401 | "X-User-ID header is required" |
| Invalid user | 401 | "invalid user" |

### 6. Data Isolation

Users can ONLY:
- Create TODOs associated with their own user ID
- List their own TODOs
- View their own TODOs
- Update their own TODOs
- Delete their own TODOs

Attempting to access another user's TODO returns 404 (not 403, to prevent information leakage).

## Contract Compliance

### Fully Implemented Contract Features

- ✅ All 5 endpoints (CREATE, LIST, GET, UPDATE, DELETE)
- ✅ All required fields (id, title, description, status, created_at)
- ✅ All validation rules (length, format, enum values)
- ✅ All error cases and status codes
- ✅ UUID v4 for IDs
- ✅ RFC3339 timestamps
- ✅ Status enum (pending, completed)
- ✅ Query parameter filtering (status)
- ✅ Title trimming and validation
- ✅ In-memory storage
- ✅ Physical deletion (not soft delete)
- ✅ Sorted by creation time (newest first)

### Extensions Beyond Contract (Per Instructions)

- ✅ User authentication via X-User-ID header
- ✅ User validation via user-service
- ✅ User isolation (each user sees only their TODOs)
- ✅ Internal user_id field (not exposed in API)

## File Structure

```
todo-service/
├── main.go              # HTTP server setup, routing
├── handlers.go          # HTTP handlers for all endpoints
├── models.go            # Data structures (Todo, requests, responses)
├── storage.go           # In-memory storage with thread safety
├── user_client.go       # HTTP client for user-service communication
├── go.mod               # Go module definition
├── go.sum               # Dependency checksums
├── Dockerfile           # Multi-stage Docker build
├── docker-compose.yml   # Docker Compose with user-service dependency
├── .dockerignore        # Docker build exclusions
├── README.md            # Usage documentation
├── IMPLEMENTATION_NOTES.md  # This file
└── test.sh              # Integration test script
```

## Design Patterns Used

### 1. Dependency Injection
- `Handler` struct receives `Storage` and `UserClient` dependencies
- Easy to test and mock

### 2. Repository Pattern
- `Storage` provides abstraction over data storage
- Easy to swap in-memory for persistent storage

### 3. Error Sentinel Values
- Pre-defined error constants (`ErrTodoNotFound`, etc.)
- Type-safe error comparison

### 4. HTTP Client Abstraction
- `UserClient` encapsulates user-service communication
- Timeout and error handling centralized

## Testing Strategy

### Manual Testing
Use the provided `test.sh` script which tests:
1. User creation (via user-service)
2. TODO creation with/without description
3. TODO listing (all, by status)
4. TODO retrieval by ID
5. TODO status update
6. TODO deletion
7. All error cases
8. User isolation

### Contract Testing
The implementation follows all contract test scenarios:
- TODO creation and retrieval
- Status updates
- Empty title validation
- Non-existent TODO handling

## Deployment

### Prerequisites
1. user-service must be built and available as `tsubo-user-service:latest`

### Build and Run

```bash
# Build user-service first
cd ../user-service
docker build -t tsubo-user-service:latest .

# Build and run todo-service
cd ../todo-service
docker-compose up --build
```

### Ports
- todo-service: http://localhost:8081
- user-service: http://localhost:8080

### Environment Variables
- `PORT`: Server port (default: 8080)
- `USER_SERVICE_URL`: User service URL (default: http://user-service:8080)

## Future Enhancements (Not in Current Scope)

1. **Persistent Storage**: Replace in-memory with PostgreSQL/MySQL
2. **Pagination**: Add limit/offset for large TODO lists
3. **Full-text Search**: Search TODOs by title/description
4. **Due Dates**: Add due date field
5. **Categories/Tags**: Allow organizing TODOs
6. **Metrics**: Add Prometheus metrics
7. **Logging**: Structured logging with levels
8. **Tracing**: Distributed tracing support

## Tsubo Philosophy Adherence

### ✅ Docker First
- Everything runs in Docker containers
- Multi-stage build for small images
- docker-compose for orchestration
- No local dependencies installed

### ✅ Contract is Everything
- Implementation follows contract exactly
- No features beyond contract (except user auth per instructions)
- All error cases handled as specified
- All validations match contract rules

### ✅ Go Language Choice
- Simple, explicit code
- Standard library for HTTP
- Minimal dependencies (only UUID)
- Explicit error handling
- Standard formatting (gofmt)

### ✅ No Questions During Implementation
- Contract was complete enough to implement
- Decisions made autonomously based on contract
- User instructions clarified authentication approach

## Performance Characteristics

### Latency
- In-memory storage: O(1) for get/create/update/delete
- List operations: O(n) where n = total TODOs for user
- Sorting: O(n log n) for list operations

### Throughput
- Limited by:
  1. User validation call to user-service
  2. JSON encoding/decoding
  3. Mutex contention on write operations

### Scalability
- Current: Single instance, in-memory
- Future: Horizontal scaling with persistent DB + cache

## Security Considerations

### Implemented
- User authentication (X-User-ID header)
- User validation via external service
- Data isolation (users can't access others' TODOs)
- Input validation (length, format)
- UUID validation to prevent path traversal

### Not Implemented (Future)
- JWT/OAuth authentication
- Rate limiting
- HTTPS/TLS
- CORS configuration
- Request signing
- Audit logging

## Code Quality

### Standards Followed
- Go best practices
- Standard library usage
- Explicit error handling
- No magic numbers
- Clear variable names
- Separation of concerns

### Maintainability
- Small, focused functions
- Clear separation of layers
- Easy to understand flow
- Well-documented via README

## Conclusion

This implementation demonstrates:
1. Faithful adherence to Tsubo philosophy
2. Complete contract compliance
3. Clean, maintainable Go code
4. Proper error handling
5. Thread-safe concurrent access
6. Service-to-service communication
7. User isolation and security
8. Docker-first approach

The service is production-ready for the PoC scope and can be extended for real-world use cases with persistent storage and additional features.
