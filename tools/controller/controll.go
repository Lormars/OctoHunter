package controller

import (
	"context"
	"sync"
	"time"

	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/internal/logger"
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
		logger.Infof("Module %s already running\n", name)
		return
	}

	module := NewModule(name)
	m.Modules[name] = module

	go func() {
		for {
			module.Wg.Add(1)
			logger.Infof("Starting module %s\n", name)
			startFunc(module.Ctx, module.Wg, opts)
			logger.Infof("Module %s stopped\n", name)
			m.mu.Lock()
			delete(m.Modules, name)
			m.mu.Unlock()
			logger.Infof("Module %s removed from manager\n", name)
			time.Sleep(15 * time.Minute)
		}
	}()
}

func (m *ModuleManager) StopModule(name string) {

	if module, exists := m.Modules[name]; exists {
		logger.Infoln("Stopping module ", name)
		module.Cancel()
		logger.Infoln("Waiting for module ", name, " to stop")
	} else {
		logger.Infoln("Module ", name, " not found")
	}

}
