package simulated

import "log"

type LinearParameters struct {
	values []float64
	keys   []string
}

type Parameters interface {
	Get(key string) float64
	GetInt(key string) int
}

func NewParameters(values []float64, keys []string) Parameters {
	if values == nil || keys == nil {
		log.Fatalln("parameters cannot be nil")
	}
	if len(values) < len(keys) {
		log.Fatalf("expected %d parameters but only got %d\n", len(keys), len(values))
	}
	return &LinearParameters{
		values: values,
		keys:   keys,
	}
}

func (p *LinearParameters) Get(key string) float64 {
	index := -1
	for i, k := range p.keys {
		if k == key {
			index = i
			break
		}
	}
	if index == -1 {
		log.Fatalf("parameter \"%s\" does not exist\n", key)
	}
	return p.values[index]
}

func (p *LinearParameters) GetInt(key string) int {
	return int(p.Get(key))
}
