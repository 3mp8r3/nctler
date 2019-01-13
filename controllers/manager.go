package controllers

import (
	"fmt"
	"sync"

	"node/common"

	"github.com/Sirupsen/logrus"
)

// ControllerManager manage and start controllers
type ControllerManager struct {
	mutex       sync.Mutex
	controllers map[string]common.Controller
}

// Register register controllers to ControllerManager
func (cm *ControllerManager) Register(controllers []common.Controller) []error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	ret := make([]error, 0, len(controllers))
	if cm.controllers == nil {
		cm.controllers = map[string]common.Controller{}
	}

	for _, controller := range controllers {
		name := controller.GetControllerName()

		if _, found := cm.controllers[name]; found {
			ret = append(ret, fmt.Errorf("controller %s: registered more than once", name))
			continue
		}
		go controller.Start()
		// go func() {
		// 	time.Sleep(10 * time.Second)
		// 	cm.StopAll()
		// }()

		cm.controllers[name] = controller
		logrus.Infof("success loaded controller %s", name)
	}
	return ret
}

// StopAll stop all controllers
func (cm *ControllerManager) StopAll() {
	for _, c := range cm.controllers {
		c.Stop()
	}
}
