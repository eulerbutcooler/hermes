# hermes-hooks (Ingestion Service)

Ingestion layer of the Hermes automation platform.
Takes in the requests and pushes them onto the NATS JetStream queue.

Uses Chi Router, NATS JetStream and Go (ofc!)

```bash
docker compose up
```

```bash
go mod tidy
```

```bash
go run cmd/server/main.go
```

Sample Request - 
```
curl -X POST http://localhost:8080/hooks/zap_123 -H "Content-Type: application/json" -d '{"event_type": "push", "repo": "my-project"}'
```

Expected Response - 
```{"status":"queued"}```

To run test:

```
go test ./internal/api... -v
```
