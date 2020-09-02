package inp_test

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/Konstantin8105/inp"
)

func Example() {
	for _, file := range []string{
		"./convertorInp/Example/shell.inp",
		"./convertorInp/Example/shell2.inp",
		"./convertorInp/Example/cone.inp",
	} {
		fmt.Fprintf(os.Stdout, "\nfilename: %s\n", file)
		content, err := ioutil.ReadFile(file)
		if err != nil {
			panic(err)
		}
		f,err := inp.Parse(content)
		if err != nil {
			panic(err)
		}
		fmt.Fprintf(os.Stdout, "%s", f.String())
	}
	// Output:
}
