package inp_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/Konstantin8105/inp"
)

func Test(t *testing.T) {
	var files []string

	for _, pattern := range []string{
		"beamlin",
		"beammix",
		"beammpc",
		"beammr",
		"beammnh",
		"beammld",
		// "shell",
		// "concrete",
		"solidshell",
	} {
		fs, err := filepath.Glob(".test/" + pattern + "*.inp")
		if err != nil {
			t.Fatal(err)
		}
		files = append(files, fs...)
	}

	for _, f := range files {
		t.Run(f, func(t *testing.T) {
			content, err := ioutil.ReadFile(f)
			if err != nil {
				t.Fatal(err)
			}
			_, err = inp.Parse(content)
			if err != nil {
				t.Fatal(err)
			}
		})
	}

}

func Example() {
	for _, file := range []string{
		"./convertorInp/Example/shell.inp",
		"./convertorInp/Example/shell2.inp",
		"./convertorInp/Example/cone.inp",
	} {
		fmt.Fprintf(os.Stdout, "\nfilename: %s\n", file)
		content, err := ioutil.ReadFile(file)
		if err != nil {
			fmt.Fprintf(os.Stdout, "%v\n", err)
			return
		}
		f, err := inp.Parse(content)
		if err != nil {
			fmt.Fprintf(os.Stdout, "%v\n", err)
			return
		}
		fmt.Fprintf(os.Stdout, "%s", f.String())
	}
	// Output:
}
