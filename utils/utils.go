package utils

func InsertAt[T any](sliceA []T, sliceB []T, index int) []T {
	if index < 0 || index > len(sliceA) {
		panic("index out of range")
	}

	return append(sliceA[:index], append(sliceB, sliceA[index:]...)...)
}
