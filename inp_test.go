package inp_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Konstantin8105/inp"
)

const data = "testdata"

func TestParse(t *testing.T) {
	var files []string

	// pattern := ""
	// _ = pattern
	for _, pattern := range []string{
		"beam8b",
		"beamb",
		"beamlin",
		// "beammix",
		// "beammpc",
		// "beammr",
		// "concrete",
		"platestress",
		"shellbeam",
		"solidshell",
		// "simplebeam",
		// "shell3",
	} {
		fs, err := filepath.Glob(filepath.Join(data, pattern+"*.inp"))
		if err != nil {
			t.Fatal(err)
		}
		files = append(files, fs...)
	}

	files = append(files, []string{
		filepath.Join("convertorInp", "Example", "shell.inp"),
		filepath.Join("convertorInp", "Example", "shell2.inp"),
		filepath.Join("convertorInp", "Example", "cone.inp"),
		filepath.Join(data, "/beampiso.inp"),
	}...)

	for _, f := range files {
		t.Run(f, func(t *testing.T) {
			content, err := os.ReadFile(f)
			if err != nil {
				t.Fatal(err)
			}
			format, err := inp.Parse(content)
			if err != nil {
				t.Fatal(err)
			}
			f1 := format.String()
			_ = os.WriteFile("f1.out", []byte(f1), 0644)
			t.Logf("step1 ... ok")

			format2, err := inp.Parse([]byte(f1))
			if err != nil {
				t.Fatal(err)
			}
			f2 := format2.String()
			_ = os.WriteFile("f2.out", []byte(f2), 0644)
			t.Logf("step2 ... ok")

			format3, err := inp.Parse([]byte(f2))
			if err != nil {
				t.Fatal(err)
			}
			f3 := format3.String()
			_ = os.WriteFile("f3.out", []byte(f3), 0644)
			t.Logf("step3 ... ok")

			if f1 != f3 {
				t.Fatalf("not same")
			}
			if testing.Verbose() {
				t.Logf("%s", f3)
			}
			// fmt.Fprintf(os.Stdout, "%s", format.String())
		})
	}
}

func TestDat(t *testing.T) {
	var files []string

	for _, pattern := range []string{
		"beam8b",
		"beamb",
		"beamlin",
		"beammix",
		"beammpc",
		// "beammr",
		"concrete",
		"platestress",
		"shellbeam",
		"solidshell",
		// "simplebeam",
		"shell3",
	} {
		fs, err := filepath.Glob(filepath.Join(data, pattern+"*.dat*"))
		if err != nil {
			t.Fatal(err)
		}
		files = append(files, fs...)
	}

	files = append(files,
		filepath.Join(data, "plastic.dat"),
		filepath.Join(data, "tmpfile.dat"),
		filepath.Join(data, "pe.dat"),
		filepath.Join(data, "eigen.dat"),
		filepath.Join(data, "temperature.dat"),
	)

	for _, f := range files {
		t.Run(f, func(t *testing.T) {
			content, err := os.ReadFile(f)
			if err != nil {
				t.Fatal(err)
			}
			d, err := inp.ParseDat(content)
			if err != nil {
				t.Fatal(err)
			}
			_ = d
		})
	}
}

func Benchmark(b *testing.B) {
	content, err := os.ReadFile(filepath.Join(data, "plastic.dat"))
	if err != nil {
		panic(err)
	}
	for n := 0; n < b.N; n++ {
		_, err := inp.ParseDat(content)
		if err != nil {
			panic(err)
		}
	}
}
