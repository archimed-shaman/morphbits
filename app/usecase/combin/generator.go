package combin

type Generator[A any] struct {
	dict      []A
	length    int
	predicate func([]A) bool

	previous, next []A
	idx            []int // Just generate next values digit-by-digit to avoid complicated math
	done           bool
}

func NewGenerator[A any](dict []A, length int, predicate func([]A) bool) *Generator[A] {
	if len(dict) == 0 {
		panic("empty dictionary")
	}

	if length <= 0 {
		panic("non-positive length")
	}

	return &Generator[A]{
		dict:      dict,
		length:    length,
		predicate: predicate,

		previous: nil,
		idx:      make([]int, length),
		done:     false,
	}
}

func (g *Generator[A]) Next() bool {
	// Generate the first possible value if any for the initial iteration
	for !g.done && (g.next == nil || !compare(g.next, g.predicate)) {
		g.next, g.done = mkNext(g.next, g.dict, g.idx)
	}

	hasValue := g.next != nil
	if hasValue {
		g.previous = g.next
		g.next = nil
	}

	// Generate the next possible value
	predicate := false
	for !g.done && !predicate {
		g.next, g.done = mkNext(g.next, g.dict, g.idx)
		predicate = compare(g.next, g.predicate)
	}

	if !predicate {
		g.next = nil
	}

	return hasValue
}

func (g *Generator[A]) Combination(dst []A) []A {
	if dst == nil {
		dst = make([]A, len(g.previous))
	} else if len(dst) != len(g.previous) {
		panic("wrong input slice length")
	}

	copy(dst, g.previous)

	return dst
}

func compare[A any](value []A, predicate func([]A) bool) bool {
	if predicate == nil {
		return true
	}

	return predicate(value)
}

func mkNext[A any](dst, dict []A, idx []int) ([]A, bool) {
	done := false

	if dst == nil {
		dst = make([]A, len(idx))
	}

	for i := 0; i < len(idx); i++ {
		dst[i] = dict[idx[i]]
	}

	var shift bool

	for i := len(idx) - 1; i >= 0; i-- {
		idx[i]++

		shift = idx[i] >= len(dict)
		if shift {
			idx[i] = 0
		} else {
			break
		}
	}

	if shift {
		done = true
	}

	return dst, done
}
