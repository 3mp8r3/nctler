package common

// Controller is a infinite loop
type Controller interface {
	Start() error
	Stop()
	GetControllerName() string
}
