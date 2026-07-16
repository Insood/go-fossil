package main

import "fmt"

func Must[T any](value T, ok bool) T {
	if !ok {
		panic(fmt.Errorf("must: missing value"))
	}
	return value
}
