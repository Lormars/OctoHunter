package controller

import (
	"context"
	"fmt"
	"sync"

	"github.com/lormars/octohunter/common"
)

type Module struct {
	Name   string
	Ctx    context.Context
	Cancel context.CancelFunc
	Wg     *sync.WaitGroup
}

func NewModule(name string) *Module {
	ctx, cancel := context.WithCancel(context.Background())
	return &Module{
		Name:   name,
		Ctx:    ctx,
		Cancel: cancel,
		Wg:     &sync.WaitGroup{},
	}
}

type ModuleManager struct {
	Modules map[string]*Module
	mu      sync.Mutex
}

func NewModuleManager() *ModuleManager {
	return &ModuleManager{
		Modules: make(map[string]*Module),
	}
}

func (m *ModuleManager) StartModule(name string, startFunc func(ctx context.Context, wg *sync.WaitGroup, opts *common.Opts), opts *common.Opts) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.Modules[name]; exists {
		return
	}

	module := NewModule(name)
	m.Modules[name] = module

	module.Wg.Add(1)
	go func() {
		startFunc(module.Ctx, module.Wg, opts)
		m.mu.Lock()
		defer m.mu.Unlock()
		delete(m.Modules, name)
	}()
}

func (m *ModuleManager) StopModule(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if module, exists := m.Modules[name]; exists {
		fmt.Println("Stopping module ", name)
		module.Cancel()
		fmt.Println("Waiting for module ", name, " to stop")
		module.Wg.Wait()
		delete(m.Modules, name)
		fmt.Println("Module ", name, " stopped")
	} else {
		fmt.Println("Module ", name, " not found")
	}

}
