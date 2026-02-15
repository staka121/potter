# User Service - Quick Start Guide

## Start the Service

```bash
cd /Users/shun/Workspace/tsubo/poc/implementations/user-service
docker-compose up -d
```

Wait a few seconds for the service to start, then verify:

```bash
curl http://localhost:8082/health
# Expected: OK
```

## API Examples

### Create a User

```bash
curl -X POST http://localhost:8082/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Alice Smith",
    "email": "alice@example.com"
  }'
```

**Response (201 Created):**
```json
{
  "id": "f088f42b-d74b-4624-b16e-053d3f85f398",
  "name": "Alice Smith",
  "email": "alice@example.com",
  "created_at": "2026-02-15T11:24:18.692402467Z"
}
```

### Get User by ID

```bash
# Replace {id} with actual user ID
curl http://localhost:8082/api/v1/users/f088f42b-d74b-4624-b16e-053d3f85f398
```

**Response (200 OK):**
```json
{
  "id": "f088f42b-d74b-4624-b16e-053d3f85f398",
  "name": "Alice Smith",
  "email": "alice@example.com",
  "created_at": "2026-02-15T11:24:18.692402467Z"
}
```

### List All Users

```bash
curl http://localhost:8082/api/v1/users
```

**Response (200 OK):**
```json
{
  "users": [
    {
      "id": "f088f42b-d74b-4624-b16e-053d3f85f398",
      "name": "Alice Smith",
      "email": "alice@example.com",
      "created_at": "2026-02-15T11:24:18.692402467Z"
    }
  ]
}
```

### Validate a User

```bash
curl -X POST http://localhost:8082/api/v1/users/validate \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "f088f42b-d74b-4624-b16e-053d3f85f398"
  }'
```

**Response (200 OK) - User exists:**
```json
{
  "valid": true,
  "user": {
    "id": "f088f42b-d74b-4624-b16e-053d3f85f398",
    "name": "Alice Smith",
    "email": "alice@example.com",
    "created_at": "2026-02-15T11:24:18.692402467Z"
  }
}
```

**Response (404 Not Found) - User doesn't exist:**
```json
{
  "valid": false
}
```

## Error Examples

### Duplicate Email

```bash
# Try to create a user with an existing email
curl -X POST http://localhost:8082/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Another Alice",
    "email": "alice@example.com"
  }'
```

**Response (409 Conflict):**
```json
{
  "error": "email already exists"
}
```

### Invalid Email Format

```bash
curl -X POST http://localhost:8082/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Bob",
    "email": "not-an-email"
  }'
```

**Response (400 Bad Request):**
```json
{
  "error": "invalid email format"
}
```

### Empty Name

```bash
curl -X POST http://localhost:8082/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "name": "",
    "email": "test@example.com"
  }'
```

**Response (400 Bad Request):**
```json
{
  "error": "name is required"
}
```

### User Not Found

```bash
curl http://localhost:8082/api/v1/users/00000000-0000-0000-0000-000000000000
```

**Response (404 Not Found):**
```json
{
  "error": "user not found"
}
```

## Email Normalization

The service automatically normalizes email addresses to lowercase:

```bash
curl -X POST http://localhost:8082/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Jane Doe",
    "email": "JANE@EXAMPLE.COM"
  }'
```

**Response:**
```json
{
  "id": "ca7c5c03-daa0-433c-90b2-2c23fd8a40e0",
  "name": "Jane Doe",
  "email": "jane@example.com",  // Note: lowercase
  "created_at": "2026-02-15T11:24:18.704680467Z"
}
```

## Run Contract Tests

```bash
./test-contract.sh
```

This will run all contract compliance tests and verify that the service behaves according to the specification.

## Stop the Service

```bash
docker-compose down
```

This will stop and remove the container, network, and clean up all resources.

## Troubleshooting

### Port Already in Use

If port 8082 is already in use, edit `docker-compose.yml` and change the port mapping:

```yaml
ports:
  - "8083:8080"  # Change 8082 to 8083 or any available port
```

### View Logs

```bash
docker-compose logs -f
```

### Restart Service

```bash
docker-compose restart
```

### Rebuild Image

```bash
docker-compose build --no-cache
docker-compose up -d
```
