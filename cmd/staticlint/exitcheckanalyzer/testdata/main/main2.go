package main

import (
	"fmt"
	"os"
)

func test() {
	fmt.Printf("Test analyzer with the os.Exit not in main function")
	os.Exit(1)
}
