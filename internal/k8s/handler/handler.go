package handler

import (
	"k8sync/internal/config"
	"k8sync/pkg/logger"
)

// Handler is implemented by any handler.
// The Handle method is used to process event
type Handler interface {
	Init(c *config.Config) error
	Handle(e *Event)
	Clean()
}

// Map maps each event handler function to a name for easily lookup
var Map = map[string]Handler{
	"default": &Default{},
}

// Default handler implements Handler interface,
// print each event with JSON format
type Default struct {
}

// Init initializes handler configuration
// Do nothing for default handler
func (d *Default) Init(c *config.Config) error {
	return nil
}

// Handle handles an event.
func (d *Default) Handle(e *Event) {
	logger.Infof("%v", e)
}

func (d *Default) Clean() {
}
