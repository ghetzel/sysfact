package plugins

type Observation struct {
	Name  string
	Value interface{}
}

type Plugin interface {
	Collect() ([]Observation, error)
}
