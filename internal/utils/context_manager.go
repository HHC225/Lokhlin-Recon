// plugins-------------------------------------
// @file      : context_manager.go
// @time      : 2025/12/20 20:15
// -------------------------------------------

package utils

import (
	"context"
	"sync"
)

type contextInfo struct {
	ctx    context.Context
	cancel context.CancelFunc
}

var (
	pluginContexts = make(map[string]*contextInfo)

	contextMutex sync.RWMutex
)

func GetContext(id string) context.Context {
	contextMutex.Lock()
	defer contextMutex.Unlock()

	if info, exists := pluginContexts[id]; exists && info != nil {
		return info.ctx
	}

	ctx, cancel := context.WithCancel(context.Background())
	pluginContexts[id] = &contextInfo{
		ctx:    ctx,
		cancel: cancel,
	}

	return ctx
}

func CancelContext(id string) {
	contextMutex.Lock()
	defer contextMutex.Unlock()

	if info, exists := pluginContexts[id]; exists && info != nil {
		info.cancel()
		delete(pluginContexts, id)
	}
}

func HasContext(id string) bool {
	contextMutex.RLock()
	defer contextMutex.RUnlock()

	_, exists := pluginContexts[id]
	return exists
}
