package simulated

type Memory struct {
	data interface{}
}

func NewMemory() *Memory {
	return &Memory{
		data: nil,
	}
}

func (m *Memory) Store(data interface{}) {
	m.data = data
}

func (m *Memory) Read() interface{} {
	return m.data
}
