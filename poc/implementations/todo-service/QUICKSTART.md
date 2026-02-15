# Quick Start Guide

## Prerequisites

- Docker and Docker Compose installed
- user-service already built and tagged as `tsubo-user-service:latest`

## Step 1: Build user-service (if not already done)

```bash
cd /Users/shun/Workspace/tsubo/poc/implementations/user-service
docker build -t tsubo-user-service:latest .
```

## Step 2: Start todo-service

```bash
cd /Users/shun/Workspace/tsubo/poc/implementations/todo-service
docker-compose up --build
```

This will:
- Build the todo-service Docker image
- Start both user-service and todo-service
- Expose user-service on port 8080
- Expose todo-service on port 8081

## Step 3: Test the Service

### Option A: Use the test script

```bash
./test.sh
```

This runs a comprehensive test suite covering all endpoints and edge cases.

### Option B: Manual testing

#### 1. Create a user first

```bash
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Alice",
    "email": "alice@example.com"
  }'
```

Save the returned user ID.

#### 2. Create a TODO

```bash
curl -X POST http://localhost:8081/api/v1/todos \
  -H "Content-Type: application/json" \
  -H "X-User-ID: <user-id-from-step-1>" \
  -d '{
    "title": "Buy groceries",
    "description": "Milk, bread, eggs"
  }'
```

#### 3. List TODOs

```bash
curl -X GET http://localhost:8081/api/v1/todos \
  -H "X-User-ID: <user-id>"
```

#### 4. Update TODO status

```bash
curl -X PATCH http://localhost:8081/api/v1/todos/<todo-id> \
  -H "Content-Type: application/json" \
  -H "X-User-ID: <user-id>" \
  -d '{
    "status": "completed"
  }'
```

#### 5. Delete a TODO

```bash
curl -X DELETE http://localhost:8081/api/v1/todos/<todo-id> \
  -H "X-User-ID: <user-id>"
```

## Step 4: Stop the Services

```bash
docker-compose down
```

## Troubleshooting

### user-service not found

If you get an error about user-service image not found:

```bash
cd /Users/shun/Workspace/tsubo/poc/implementations/user-service
docker build -t tsubo-user-service:latest .
```

### Port already in use

If ports 8080 or 8081 are already in use, you can change them in `docker-compose.yml`:

```yaml
ports:
  - "8081:8080"  # Change 8081 to another port
```

### Services not starting

Check logs:

```bash
docker-compose logs todo-service
docker-compose logs user-service
```

### Health check failing

Wait a few seconds for services to start. Health checks run every 10 seconds.

## Verification

Check that services are healthy:

```bash
# Check user-service
curl http://localhost:8080/health

# Check todo-service
curl http://localhost:8081/health
```

Both should return `OK`.

## Next Steps

- See [README.md](./README.md) for detailed API documentation
- See [IMPLEMENTATION_NOTES.md](./IMPLEMENTATION_NOTES.md) for implementation details
- Review the contract at `/Users/shun/Workspace/tsubo/poc/contracts/todo-service.object.yaml`
