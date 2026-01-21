package engine

import "context"

type ActionExecutor interface {
	Execute(ctx context.Context, config map[string]interface{}, payload []byte) error
}
