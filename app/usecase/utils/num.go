package utils

func MaxInt() int {
	return int(^uint(0) >> 1)
}
