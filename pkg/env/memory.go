package env

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

type FiLoStack struct {
	stack   []interface{}
	index   int
	size    int
	counter int
}

func NewFiLoStack(size int) *FiLoStack {
	return &FiLoStack{
		stack: make([]interface{}, size),
		index: 0,
		size:  size,
	}
}

func (s *FiLoStack) Size() int {
	return s.size
}

func (s *FiLoStack) IsFull() bool {
	return s.counter >= s.size
}

func (s *FiLoStack) Clear() {
	for i := range s.stack {
		s.stack[i] = nil
	}
	s.index = 0
	s.counter = 0
}

func (s *FiLoStack) Push(item interface{}) {
	s.stack[s.index] = item
	s.index--
	if s.index < 0 {
		s.index += s.size
	}
	s.counter++
}

func (s *FiLoStack) At(index int) interface{} {
	i := s.index - index
	if i < 0 {
		i += s.size
	}
	return s.stack[i]
}

func (s *FiLoStack) ToSlice() []interface{} {
	return s.stack
}
