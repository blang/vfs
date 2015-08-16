package examples

import (
	"fmt"
	"os"
)

func fatal(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args)
	os.Exit(1)
}
