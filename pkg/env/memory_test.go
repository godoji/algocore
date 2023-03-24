package env

import (
	"fmt"
	"testing"
)

func TestFiLoStack_At(t *testing.T) {

	stack := NewFiLoStack(5)
	stack.Push(1)
	stack.Push(2)
	stack.Push(3)
	stack.Push(4)
	stack.Push(5)
	stack.Push(6)
	stack.Push(7)

	if stack.At(0).(int) != 3 {
		fmt.Printf("expected %d but got %d\n", 3, stack.At(0).(int))
		t.Fail()
	}
	if stack.At(1).(int) != 4 {
		fmt.Printf("expected %d but got %d\n", 4, stack.At(1).(int))
		t.Fail()
	}
	if stack.At(2).(int) != 5 {
		fmt.Printf("expected %d but got %d\n", 5, stack.At(2).(int))
		t.Fail()
	}
	if stack.At(3).(int) != 6 {
		fmt.Printf("expected %d but got %d\n", 6, stack.At(3).(int))
		t.Fail()
	}
	if stack.At(4).(int) != 7 {
		fmt.Printf("expected %d but got %d\n", 7, stack.At(4).(int))
		t.Fail()
	}
}
