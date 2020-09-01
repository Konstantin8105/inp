package convertorInp

import (
	"io/ioutil"

	"github.com/Konstantin8105/inp"
	"github.com/Konstantin8105/inp/staad"
)

// Convert - convert test of inp file format to std format
func Convert(inpFilename, stdFilename string) (err error) {
	var f inp.Format
	err = f.Open(inpFilename)
	if err != nil {
		return err
	}
	std := convertInpToStd(f)
	return ioutil.WriteFile(stdFilename, []byte(std.String), 0666)
}

func convertInpToStd(f inp.Format) (std staad.Format) {
	for _, node := range f.Nodes {
		coord := staad.Coordinate{Index: node.Index, Coord: node.Coord}
		std.Points = append(std.Points, coord)
	}
	for _, element := range f.Elements {
		switch len(element) {
		case 2:
			var b staad.Beam
			b.Index = data.Index
			b.IPoint = data.Nodes
			std.Members = append(std.Members, b)
		case 3:
			var s staad.Shell
			s.Index = data.Index
			s.IPoint = data.Nodes
			std.Shells = append(std.Shells, s)
		case 4:
			var s staad.Shell
			s.Index = data.Index
			s.IPoint = data.Nodes
			std.Shells = append(std.Shells, s)
		default:
			// TODO realize more flexibility
		}
	}
	return
}
