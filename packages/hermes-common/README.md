# hermes-common

Shared utilities and packages for all Hermes services.

## Packages

- `pkg/logger` - Structured logging with slog
- (Future: `pkg/errors`, `pkg/middleware`, `pkg/metrics`)

## Usage

```go
import "github.com/eulerbutcooler/hermes-common/pkg/logger"

log := logger.New("my-service", "development", "INFO")
log.Info("service started")
```
