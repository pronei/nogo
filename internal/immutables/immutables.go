package immutables

type Map[K comparable, V any] struct {
	internalMap map[K]V
}

func NewMap[K comparable, V any](m map[K]V) *Map[K, V] {
	n := make(map[K]V, len(m))
	for k, v := range m {
		n[k] = v
	}
	return &Map[K, V]{n}
}

func (m *Map[K, V]) Get(key K) (V, bool) {
	val, ok := m.internalMap[key]
	return val, ok
}

type Set[T comparable] struct {
	internalSet map[T]struct{}
}

func NewSet[T comparable](args ...T) *Set[T] {
	s := make(map[T]struct{}, len(args))
	for _, arg := range args {
		s[arg] = struct{}{}
	}
	return &Set[T]{s}
}

func (s *Set[T]) Has(key T) bool {
	_, ok := s.internalSet[key]
	return ok
}
