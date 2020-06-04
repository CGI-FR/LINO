package id

// set is a simple implementation of the set pattern.
type set map[string]string

func newSet() set {
	return map[string]string{}
}

// add an element in the set
func (s set) add(v string) {
	s[v] = v
}

// remove an element in the set
func (s set) remove(v string) {
	delete(s, v)
}

// contains the element ?
func (s set) contains(v string) bool {
	_, ok := s[v]
	return ok
}
