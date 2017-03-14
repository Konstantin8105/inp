package convertorInp

// Convert - convert test of inp file format to std format
func Convert(inpFilename, stdFilename string) (err error) {
	inp, err := readInp(inpFilename)
	if err != nil {
		return err
	}

	std := convertInpToStd(inp)

	strings, err := std.getLines()
	if err != nil {
		return err
	}

	err = createNewFile(strings)
	return err
}
