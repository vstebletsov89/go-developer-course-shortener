package pkg1

import (
	"fmt"
	"os"
)

func main() {
	fmt.Printf("Test analyzer with the os.Exit not in main package")
	os.Exit(1)
}
