package combin

import "testing"

func Test_Generator1(t *testing.T) {
	t.Parallel()

	g := NewGenerator([]int{1, 2}, 3, nil)

	got := make([][]int, 0)

	for g.Next() {
		buff := make([]int, 3)
		buff = g.Combination(buff)

		got = append(got, buff)
	}

	expected := [][]int{
		{1, 1, 1},
		{1, 1, 2},
		{1, 2, 1},
		{1, 2, 2},
		{2, 1, 1},
		{2, 1, 2},
		{2, 2, 1},
		{2, 2, 2},
	}

	for i := 0; i < len(expected); i++ {
		for j := 0; j < len(expected[i]); j++ {
			if got[i][j] != expected[i][j] {
				t.Fatalf("wrong sequence: %v", got)
			}
		}
	}
}

func Test_Generator2(t *testing.T) {
	t.Parallel()

	g := NewGenerator([]int{1, 2, 3}, 3, nil)

	got := make([][]int, 0)

	for g.Next() {
		buff := make([]int, 3)
		buff = g.Combination(buff)

		got = append(got, buff)
	}

	expected := [][]int{
		{1, 1, 1},
		{1, 1, 2},
		{1, 1, 3},
		{1, 2, 1},
		{1, 2, 2},
		{1, 2, 3},
		{1, 3, 1},
		{1, 3, 2},
		{1, 3, 3},

		{2, 1, 1},
		{2, 1, 2},
		{2, 1, 3},
		{2, 2, 1},
		{2, 2, 2},
		{2, 2, 3},
		{2, 3, 1},
		{2, 3, 2},
		{2, 3, 3},

		{3, 1, 1},
		{3, 1, 2},
		{3, 1, 3},
		{3, 2, 1},
		{3, 2, 2},
		{3, 2, 3},
		{3, 3, 1},
		{3, 3, 2},
		{3, 3, 3},
	}

	for i := 0; i < len(expected); i++ {
		for j := 0; j < len(expected[i]); j++ {
			if got[i][j] != expected[i][j] {
				t.Fatalf("wrong sequence: %v", got)
			}
		}
	}
}

func Test_GeneratorStrings(t *testing.T) {
	t.Parallel()

	g := NewGenerator([]string{"first", "second"}, 2, nil)

	got := make([][]string, 0)

	for g.Next() {
		buff := make([]string, 2)
		buff = g.Combination(buff)

		got = append(got, buff)
	}

	expected := [][]string{
		{"first", "first"},
		{"first", "second"},
		{"second", "first"},
		{"second", "second"},
	}

	for i := 0; i < len(expected); i++ {
		for j := 0; j < len(expected[i]); j++ {
			if got[i][j] != expected[i][j] {
				t.Fatalf("wrong sequence: %v", got)
			}
		}
	}
}

func Test_GeneratorFilter(t *testing.T) {
	t.Parallel()

	g := NewGenerator([]int{1, 2, 3}, 3, func(v []int) bool {
		return v[0] == 2 && v[1] == 2 && v[2] == 2
	})

	got := make([]int, 3)

	for g.Next() {
		g.Combination(got)
	}

	if got[0] != 2 || got[1] != 2 || got[2] != 2 {
		t.Errorf("wrong sequence: %v", got)
	}
}

func Test_GeneratorEmptyDict(t *testing.T) {
	t.Parallel()

	assertPanic(t, func() {
		g := NewGenerator([]int{}, 3, nil)

		got := make([]int, 3)

		for g.Next() {
			g.Combination(got)
		}
	})
}

func Test_GeneratorZeroLength(t *testing.T) {
	t.Parallel()

	assertPanic(t, func() {
		g := NewGenerator([]int{1, 2, 3}, 0, nil)

		got := make([]int, 0)

		for g.Next() {
			g.Combination(got)
		}
	})
}

func assertPanic(t *testing.T, f func()) {
	t.Helper()

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	f()
}
