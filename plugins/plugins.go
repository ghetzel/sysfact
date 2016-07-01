package plugins

import (
	"github.com/op/go-logging"
)

var log = logging.MustGetLogger(`plugins`)

type Observation struct {
	Name  string
	Value interface{}
}

type Plugin interface {
	Collect() ([]Observation, error)
}
