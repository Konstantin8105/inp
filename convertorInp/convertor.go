package convertorInp

import (
	"io/ioutil"

	"github.com/Konstantin8105/inp"
	"github.com/Konstantin8105/inp/staad"
)

// Convert - convert test of inp file format to std format
func Convert(inpFilename, stdFilename string) (err error) {
	content, err := ioutil.ReadFile(inpFilename)
	if err != nil {
		return
	}
	f, err := inp.Parse(content)
	if err != nil {
		return err
	}
	var std staad.Format
	// convert
	{
		for _, node := range f.Nodes {
			coord := staad.Coordinate{Index: node.Index, Coord: node.Coord}
			std.Points = append(std.Points, coord)
		}
		for _, data := range f.Elements {
			switch len(data.Nodes) {
			case 2:
				var b staad.Beam
				b.Index = data.Index
				b.IPoint[0] = data.Nodes[0]
				b.IPoint[1] = data.Nodes[1]
				std.Members = append(std.Members, b)
			case 3,6:
				var s staad.Shell
				s.Index = data.Index
				s.IPoint = make([]int, 3)
				for i := 0; i < 3; i++ {
					s.IPoint[i] = data.Nodes[i]
				}
				std.Shells = append(std.Shells, s)
			case 4, 8:
				var s staad.Shell
				s.Index = data.Index
				s.IPoint = make([]int, 4)
				for i := 0; i < 4; i++ {
					s.IPoint[i] = data.Nodes[i]
				}
				std.Shells = append(std.Shells, s)
			default:
				// TODO realize more flexibility
			}
		}
	}

	return ioutil.WriteFile(stdFilename, []byte(std.String()), 0666)
}
