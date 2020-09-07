package inp_test

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/Konstantin8105/inp"
)

func Test(t *testing.T) {
	var files []string

	for _, pattern := range []string{
		// "beam8b",
		// "beamb",
		// "beamlin",
		// "beammix",
		// "beammpc",
		// "beammr",
		// "concrete",
		// "platestress",
		// "shellbeam",
		// "solidshell",
	} {
		fs, err := filepath.Glob(".test/" + pattern + "*.inp")
		if err != nil {
			t.Fatal(err)
		}
		files = append(files, fs...)
	}

	files = append(files, []string{
		"./convertorInp/Example/shell.inp",
		"./convertorInp/Example/shell2.inp",
		"./convertorInp/Example/cone.inp",
	}...)

	for _, f := range files {
		t.Run(f, func(t *testing.T) {
			content, err := ioutil.ReadFile(f)
			if err != nil {
				t.Fatal(err)
			}
			format, err := inp.Parse(content)
			if err != nil {
				t.Fatal(err)
			}
			format2, err := inp.Parse([]byte(format.String()))
			if err != nil {
				t.Fatal(err)
			}
			if format.String() != format2.String() {
				t.Fatalf("not same")
			}
			// fmt.Fprintf(os.Stdout, "%s", format.String())
		})
	}
}
