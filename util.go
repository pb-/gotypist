package main

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func must(errArg int) func(...interface{}) {
	return func(retval ...interface{}) {
		if err := retval[errArg]; err != nil {
			panic(err)
		}
	}
}
