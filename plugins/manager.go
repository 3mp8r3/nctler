package plugins

import (
	"fmt"
	"sync"

	"node/common"

	"github.com/Sirupsen/logrus"
)

type PluginManager struct {
	mutex   sync.Mutex
	plugins map[string]common.BasePlugin
}

func (pm *PluginManager) RegPlugins(plugins []common.BasePlugin) []error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	ret := make([]error, 0, len(plugins))
	if pm.plugins == nil {
		pm.plugins = map[string]common.BasePlugin{}
	}

	for _, plugin := range plugins {
		name := plugin.GetPluginName()

		if _, found := pm.plugins[name]; found {
			ret = append(ret, fmt.Errorf("plugin %s: registered more than once", name))
			continue
		}
		go plugin.Init()
		//		err := plugin.Init()
		//		if err != nil {
		//			ret = append(ret, fmt.Errorf("plugin %s: init failed, %s", name, err.Error()))
		//			continue
		//		}
		pm.plugins[name] = plugin
		logrus.Infof("success loaded plugin %s", name)
	}
	return ret
}
