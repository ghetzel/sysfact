package plugins

import (
	"fmt"

	"github.com/ghetzel/sysfact/data"
)

type EmbeddedPlugin struct {
	Plugin
}

func (self EmbeddedPlugin) Collect() ([]Observation, error) {
	observations := make([]Observation, 0)

	for name, value := range data.Collect() {
		if value != nil && fmt.Sprintf("%v", value) != `` {
			observations = append(observations, Observation{
				Name:  name,
				Value: value,
			})
		}
	}

	return observations, nil
}
