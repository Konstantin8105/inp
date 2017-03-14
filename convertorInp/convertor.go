package convertorInp

// Convert - convert test of inp file format to std format
func Convert(inpFilename, stdFilename string) (err error) {
	inp, err := readInp(inpFilename)
	if err != nil {
		return err
	}

	std := convertInpToStd(inp)

	strings := std.getLines()

	err = createNewFile(stdFilename, strings)
	return err
}
