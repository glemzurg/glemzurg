package requirements

// Test only to simplify tests.
func T_Must[T any](obj T, err error) T {
	if err != nil {
		panic(err)
	}
	return obj
}
