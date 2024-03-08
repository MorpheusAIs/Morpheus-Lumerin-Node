package lib

func CopySlice[T any](s []T) []T {
	return append([]T(nil), s...)
}
