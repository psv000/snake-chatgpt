package main

import (
	"fmt"
	"os"
)

var (
	squareVertices = []float32{
		0, 0,
		1, 0,
		1, 1,
		0, 1,
	}

	indices = []uint32{
		0, 1, 2,
		0, 2, 3,
	}
)

func panicOnError(err error) {
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
