package convertorInp

import (
	"github.com/Konstantin8105/Convert-INP-to-STD-format/inp"
	"github.com/Konstantin8105/Convert-INP-to-STD-format/staad"
)

func convertInpToStd(f inp.Format) (std staad.Format) {
	for _, node := range f.Nodes {
		coord := staad.Coordinate{Index: node.Index, Coord: node.Coord}
		std.Points = append(std.Points, coord)
	}

	for _, element := range f.Elements {
		switch element.ElType {
		case inp.TypeT3D2:
			// member
			for _, data := range element.Data {
				var b staad.Beam
				b.Index = data.Index
				for i := 0; i < 2; i++ {
					b.IPoint[i] = data.IPoint[i]
				}
				std.Members = append(std.Members, b)
			}
		case inp.TypeCPS3:
			// triangle
			for _, data := range element.Data {
				var s staad.Shell
				s.Index = data.Index
				s.IPoint = data.IPoint
				std.Shells = append(std.Shells, s)
			}
		}
	}
	return
}
