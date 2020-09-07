package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/Konstantin8105/inp/convertorInp"
)

// Convert all INP files to STD files
func main() {
	for _, inpFilename := range os.Args[1:] {
		stdFilename := inpFilename
		index := strings.LastIndex(stdFilename, ".")
		stdFilename = stdFilename[:index] + ".std"
		if err := convertorInp.Convert(inpFilename, stdFilename); err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing: %v", err)
			return
		}
	}
}
