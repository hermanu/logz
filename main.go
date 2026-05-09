// Command logz parses, filters, and summarizes log files.
//
// See https://github.com/hermanu/logz for usage.
package main

import (
	"os"

	"github.com/hermanu/logz/cmd"
)

func main() {
	os.Exit(int(cmd.Run(os.Args[1:], os.Stdout, os.Stderr, os.Stdin)))
}
