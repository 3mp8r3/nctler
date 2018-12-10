package common

type BasePlugin interface {
	Init() error
	GetPluginName() string
}
