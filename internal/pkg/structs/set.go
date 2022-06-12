package structs

type StringSet struct {
	m map[string]struct{}
}

func NewStringSet() *StringSet {
	return &StringSet{
		m: map[string]struct{}{},
	}
}

func (s *StringSet) Add(strs ...string) {
	for _, str := range strs {
		s.m[str] = struct{}{}
	}
}

func (s *StringSet) Has(str string) bool {
	_, exists := s.m[str]
	return exists
}
