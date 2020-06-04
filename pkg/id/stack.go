package id

// stack is a simple implementation of the stack pattern
type stack struct {
	items []Table
}

func newStack() stack {
	return stack{items: []Table{}}
}

// push an element on top of the stack
func (s *stack) push(v Table) {
	s.items = append(s.items, v)
}

// peek the element on top of the stack (return nil is stack is empty)
/* func (s *stack) peek() *Table {
	var v *Table
	if len(s.vertices) > 0 {
		v = s.vertices[len(s.vertices)-1]
	}
	return v
} */

// pop the top element (return nil is stack is empty)
func (s *stack) pop() Table {
	var v Table
	if len(s.items) > 0 {
		v = s.items[len(s.items)-1]
		s.items = s.items[:len(s.items)-1]
	}
	return v
}

// empty is true if the stack is empty
/* func (s *stack) empty() bool {
	return len(s.vertices) == 0
} */

func (s *stack) asSlice() []Table {
	result := []Table{}
	result = append(result, s.items...)
	return result
}
