package controllers

import (
	"fmt"
	"sync"

	"github.com/maoqide/nctler/common"

	"github.com/Sirupsen/logrus"
)

// ControllerManager manage and start controllers
type ControllerManager struct {
	mutex       sync.Mutex
	controllers map[string]common.Controller
}

// NewControllerManager create ControllerManager
func NewControllerManager() *ControllerManager {
	return &ControllerManager{controllers: make(map[string]common.Controller)}
}

func (cm *ControllerManager) register(controller common.Controller) error {
	name := controller.GetControllerName()
	if _, found := cm.controllers[name]; found {
		return fmt.Errorf("controller %s: registered more than once", name)
	}
	cm.controllers[name] = controller
	logrus.Infof("success registered controller %s", name)
	return nil
}

// RegisterAll register controllers to ControllerManager
func (cm *ControllerManager) RegisterAll(controllers []common.Controller) []error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	res := make([]error, len(controllers))
	if cm.controllers == nil {
		cm.controllers = make(map[string]common.Controller)
	}

	for _, controller := range controllers {
		cm.register(controller)
		// name := controller.GetControllerName()

		// if _, found := cm.controllers[name]; found {
		// 	res = append(res, fmt.Errorf("controller %s: registered more than once", name))
		// 	continue
		// }
		// cm.controllers[name] = controller
		// logrus.Infof("success registered controller %s", name)
	}
	return res
}

// StartAll start all controllers
func (cm *ControllerManager) StartAll() {
	fmt.Println("asas", cm.controllers)
	for _, c := range cm.controllers {
		go c.Start()
	}
	// go func() {
	// 	time.Sleep(10 * time.Second)
	// 	cm.StopAll()
	// }()
}

// StopAll stop all controllers
func (cm *ControllerManager) StopAll() {
	for _, c := range cm.controllers {
		c.Stop()
	}
}
