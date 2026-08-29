package plugins

type Observation struct {
	Name  string
	Value any
}

type Plugin interface {
	Collect() ([]Observation, error)
}
