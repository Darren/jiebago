package summary

type set map[string]struct{}

func newSet() set {
	return make(map[string]struct{}, 0)
}

func (s set) has(text string) bool {
	_, ok := s[text]
	return ok
}

func (s set) add(text string) {
	if _, ok := s[text]; !ok {
		s[text] = struct{}{}
	}
}

func (s set) elems() []string {
	els := make([]string, 0, len(s))
	for k := range s {
		els = append(els, k)
	}
	return els
}

func (s set) intersect(o set) (r set) {
	r = newSet()

	for k := range s {
		if o.has(k) {
			r.add(k)
		}
	}
	return
}

func (s set) len() int {
	return len(s)
}
