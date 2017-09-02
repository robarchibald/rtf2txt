package rtf2txt

type stack struct {
	top  *element
	size int
}

type element struct {
	value string
	next  *element
}

func (s *stack) Len() int {
	return s.size
}

func (s *stack) Push(value string) {
	s.top = &element{value, s.top}
	s.size++
}

func (s *stack) Pop() (value string) {
	if s.size > 0 {
		value, s.top = s.top.value, s.top.next
		s.size--
		return
	}
	return ""
}
