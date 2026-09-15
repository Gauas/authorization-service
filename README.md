# Authorization Service

Lightweight Traefik ForwardAuth service. It verifies access tokens, checks session revocation in Redis, and returns trusted identity headers. It does not issue tokens, manage accounts, or proxy requests to backend services.

## Request path

```text
Traefik -> GET /v1/authorization/forward-auth
             -> verify EdDSA JWT from cached JWKS
             -> check revoked:sid:<session-id> in Redis
             -> return 204 and X-Gauas-* identity headers
Traefik -> destination backend
```

RabbitMQ and Kafka are selected from the `QUEUE_URL` scheme (`amqp`, `amqps`, or `kafka`). Revocation events are consumed from the durable `auth.session.revoked` destination and written to Redis with the remaining token lifetime as TTL.

## Configuration

Required:

- `JWKS_URL` (the account/authentication service `/.well-known/jwks.json` endpoint)
- `JWT_ISSUER`
- `JWT_AUDIENCE`
- `REDIS_URL`
- `QUEUE_URL`

Only EdDSA tokens with a `kid` are accepted. Shared-secret JWT verification is not supported.

## Run

```bash
go test ./...
go run ./cmd/forwardauth
```

Kubernetes resources and the `gauas-protected` middleware are owned by the Argo CD GitOps repository. Attach that middleware only to protected routes; public authentication routes must not use it.
