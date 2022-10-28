package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Printf("Test analyzer with the os.Exit in main")
	os.Exit(1) // want "os.Exit is not allowed in main package"
}
