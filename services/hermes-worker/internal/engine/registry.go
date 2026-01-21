package engine

import "fmt"

type Registry struct {
	executors map[string]ActionExecutor
}

func NewRegistry() *Registry {
	return &Registry{
		executors: make(map[string]ActionExecutor),
	}
}

func (r *Registry) Register(name string, executor ActionExecutor) {
	r.executors[name] = executor
}

func (r *Registry) Get(name string) (ActionExecutor, error) {
	exec, exists := r.executors[name]
	if !exists {
		return nil, fmt.Errorf("Unknown action type: %s", name)
	}
	return exec, nil
}
