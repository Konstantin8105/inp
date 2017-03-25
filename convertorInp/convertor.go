package convertorInp

import "github.com/Konstantin8105/Convert-INP-to-STD-format/inp"

// Convert - convert test of inp file format to std format
func Convert(inpFilename, stdFilename string) (err error) {
	var f inp.Format
	err = f.ReadInp(inpFilename)
	if err != nil {
		return err
	}

	std := convertInpToStd(f)

	strings := std.GetLines()

	err = CreateNewFile(stdFilename, strings)
	return err
}
