package app

import (
	"testing"

	"github.com/golang/mock/gomock"
	"golang.org/x/exp/constraints"
	mockApp "morphbits.io/app/usecase/app/mock"
)

func Test_getLenCombinations(t *testing.T) {
	t.Parallel()

	dict := []int{1, 3, 5}
	length := 2
	minPassLen := 5
	maxPassLen := 6

	expected := [][]int{
		{1, 5},
		{3, 3},
		{5, 1},
	}

	got := getLenCombinations(dict, length, minPassLen, maxPassLen)

	if !eqSlice2(expected, got) {
		t.Errorf("Bad length combinations.\nExpected: %v\nGot: %v", expected, got)
	}
}

func Test_calcInternalDistance(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	calc := mockApp.NewMockDistanceCalculator(ctrl)

	calc.EXPECT().GetDistance(uint8('w'), uint8('o')).Return(1, nil)
	calc.EXPECT().GetDistance(uint8('o'), uint8('r')).Return(2, nil)
	calc.EXPECT().GetDistance(uint8('r'), uint8('d')).Return(3, nil)

	dist, err := calcInternalDistance("word", calc)
	if err != nil {
		t.Fatal(err)
	}

	if expected := 1 + 2 + 3; dist != expected {
		t.Errorf("Expected: %v, got: %v", expected, dist)
	}
}

func Test_calcInternalDistance_zero(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	calc := mockApp.NewMockDistanceCalculator(ctrl)

	dist, err := calcInternalDistance("a", calc)
	if err != nil {
		t.Fatal(err)
	}

	if expected := 0; dist != expected {
		t.Errorf("Expected: %v, got: %v", expected, dist)
	}
}

func eqSlice2[A constraints.Ordered](a, b [][]A) bool {
	if len(a) != len(b) {
		return false
	}

	for i := 0; i < len(a); i++ {
		if len(a[i]) != len(b[i]) {
			return false
		}

		for j := 0; j < len(a[i]); j++ {
			if a[i][j] != b[i][j] {
				return false
			}
		}
	}

	return true
}
