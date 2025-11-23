package data

// Mappings agrupa los mapeos usados por el servicio.
// - UserOriginalToIndex:  "userId" -> userIndex (string -> int)
// - UserIndexToOriginal:  userIndex -> "userId" (int -> string)
// - MovieOriginalToIndex: "movieId" -> movieIndex (string -> int)
// - MovieIndexToOriginal: movieIndex -> "movieId" (int -> string)
type Mappings struct {
	UserOriginalToIndex  map[string]int
	UserIndexToOriginal  map[int]string
	MovieOriginalToIndex map[string]int
	MovieIndexToOriginal map[int]string
}

// NewMappings crea una estructura Mappings vacía (útil para inicializar).
func NewMappings() *Mappings {
	return &Mappings{
		UserOriginalToIndex:  make(map[string]int),
		UserIndexToOriginal:  make(map[int]string),
		MovieOriginalToIndex: make(map[string]int),
		MovieIndexToOriginal: make(map[int]string),
	}
}

// Convenience helpers

// UserIndex devuelve el índice interno dado un userId (original).
func (m *Mappings) UserIndex(userID string) (int, bool) {
	idx, ok := m.UserOriginalToIndex[userID]
	return idx, ok
}

// UserOriginal devuelve el userId original dado un índice interno.
func (m *Mappings) UserOriginal(index int) (string, bool) {
	orig, ok := m.UserIndexToOriginal[index]
	return orig, ok
}

// MovieIndex devuelve el índice interno dado un movieId (original).
func (m *Mappings) MovieIndex(movieID string) (int, bool) {
	idx, ok := m.MovieOriginalToIndex[movieID]
	return idx, ok
}

// MovieOriginal devuelve el movieId original dado un índice interno.
func (m *Mappings) MovieOriginal(index int) (string, bool) {
	orig, ok := m.MovieIndexToOriginal[index]
	return orig, ok
}
