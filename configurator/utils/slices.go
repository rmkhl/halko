package utils

func Filter[T any](ss []T, include func(item T) bool) []T {
	filtered := []T{}
	for _, s := range ss {
		if include(s) {
			filtered = append(filtered, s)
		}
	}
	return filtered
}

func Map[T, U any](ss []T, mapper func(item T) U) []U {
	mapped := make([]U, 0, len(ss))
	for _, s := range ss {
		mapped = append(mapped, mapper(s))
	}
	return mapped
}
