package lib

func Filter[T any](arr []T, f func(T) bool) (ret []T) {
	if len(arr) == 0 {
		return
	}
	for _, v := range arr {
		if f(v) {
			ret = append(ret, v)
		}
	}
	return
}

func FilterValue[T comparable](arr []T, val T) (ret []T) {
	if len(arr) == 0 {
		return
	}
	for _, v := range arr {
		if v != val {
			ret = append(ret, v)
		}
	}
	return
}
