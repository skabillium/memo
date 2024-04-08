package db

type Set struct {
	Size  int
	items map[string]bool
}

func NewSet() *Set {
	return &Set{
		Size:  0,
		items: map[string]bool{},
	}
}

func (s *Set) Items() []string {
	keys := make([]string, len(s.items))
	var i int
	for k := range s.items {
		keys[i] = k
		i++
	}

	return keys
}

func (s *Set) Add(item string) {
	_, found := s.items[item]
	if !found {
		s.Size++
	}

	s.items[item] = true
}

func (s *Set) Has(key string) bool {
	return s.items[key]
}

func (s *Set) Delete(key string) bool {
	_, found := s.items[key]
	if !found {
		return false
	}

	s.Size--
	delete(s.items, key)
	return true
}
