package virtualMachine

type SymbolMap struct {
	values []*Symbol
	size   int
}

func NewSymbolMap(size int) *SymbolMap {
	return &SymbolMap{make([]*Symbol, size), size}
}

func (s *SymbolMap) Insert(key int, value *Symbol) {
	s.values[key%s.size] = value
}

func (s *SymbolMap) Get(key int) *Symbol {
	return s.values[key%s.size]
}

func (s *SymbolMap) Delete(key int) {
	s.values[key%s.size] = nil
}
