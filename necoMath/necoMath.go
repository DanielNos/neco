package necoMath

func PowerInt64(base, exponent int64) int64 {
	var result int64 = 1

	for exponent > 0 {
		if exponent%2 == 1 {
			result *= base
		}
		base *= base
		exponent /= 2
	}

	return result
}

func FactorialInt64(base int64) int64 {
	var result int64 = 1
	for base > 1 {
		result *= base
		base--
	}
	return result
}
