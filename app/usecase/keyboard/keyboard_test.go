package keyboard

import "testing"

func Test_QWERTY(t *testing.T) {
	t.Parallel()

	kbd, err := NewQWERTY()
	if err != nil {
		t.Error(err)
	}

	testData := []struct {
		A, B     byte
		Expected int
	}{
		{'s', 'a', 1},
		{'s', 'd', 1},
		{'s', 'w', 1},
		{'s', 'x', 1},
		{'a', 'l', 8},
		{'q', 'm', 8},
		{'t', 'b', 2},
	}

	for _, testCase := range testData {
		dist, err := kbd.GetDistance(testCase.A, testCase.B)
		if err != nil {
			t.Error(err)
		}

		if dist != testCase.Expected {
			t.Errorf("Expected distance between '%s' and '%s': %d; got: %d",
				string(testCase.A),
				string(testCase.B),
				testCase.Expected,
				dist)
		}
	}
}
