package convertorInp

import (
	"github.com/Konstantin8105/Convert-INP-to-STD-format/inp"
	"github.com/Konstantin8105/Convert-INP-to-STD-format/utils"
)

// Convert - convert test of inp file format to std format
func Convert(inpFilename, stdFilename string) (err error) {
	var f inp.Format
	err = f.Open(inpFilename)
	if err != nil {
		return err
	}

	std := convertInpToStd(f)

	strings := std.GetLines()

	err = utils.CreateNewFile(stdFilename, strings)
	return err
}
