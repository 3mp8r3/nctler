package controllers

import (
	"sync"

	"github.com/maoqide/nctler/common"
)

var defaultControllerManager *ControllerManager
var cMutex sync.Mutex

func init() {
	defaultControllerManager = NewControllerManager()
}

// DefaultControllerManager return defaultControllerManager
func DefaultControllerManager() *ControllerManager {
	return defaultControllerManager
}

// RegisterDefault register to default controller manager
func RegisterDefault(c common.Controller) {
	cMutex.Lock()
	defer cMutex.Unlock()
	defaultControllerManager.register(c)
}
