package utils

import "golang.org/x/exp/constraints"

func Sum[A constraints.Ordered](values ...A) A { //nolint:ireturn // linter bug
	var sum A

	for i := 0; i < len(values); i++ {
		sum += values[i]
	}

	return sum
}

func Insert[A any](slice []A, value A, position int) []A {
	if position < 0 || position > len(slice) {
		panic("wrong index")
	}

	var a A
	slice = append(slice, a)
	copy(slice[position+1:], slice[position:])
	slice[position] = value

	return slice
}

func MakeRange(min, max int) []int {
	a := make([]int, max-min+1)
	for i := range a {
		a[i] = min + i
	}

	return a
}
