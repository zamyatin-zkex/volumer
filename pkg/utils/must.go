package utils

func Must[T any](in T, err error) T {
	if err != nil {
		panic(err)
	}
	return in
}
