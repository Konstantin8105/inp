package main

import (
	"fmt"

	"github.com/Konstantin8105/Convert-INP-to-STD-format/convertorInp"
)

func main() {
	err := convertorInp.Convert("cone.inp", "s.std")
	if err != nil {
		fmt.Println(err)
	}
}
