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
			var b staad.Beam
			b.Index = element.Index
			for i := 0; i < 2; i++ {
				b.IPoint[i] = element.IPoint[i]
			}
			std.Members = append(std.Members, b)
		case inp.TypeCPS3:
			var s staad.Shell
			s.Index = element.Index
			s.IPoint = element.IPoint
			std.Shells = append(std.Shells, s)
		}
	}
	return
}
