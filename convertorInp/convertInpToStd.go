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
		switch element.FE.Shape {
		case inp.Beam:
			switch element.FE.AmountNodes {
			case 2:
				for _, data := range element.Data {
					var b staad.Beam
					b.Index = data.Index
					for i := 0; i < 2; i++ {
						b.IPoint[i] = data.IPoint[i]
					}
					std.Members = append(std.Members, b)
				}
			default:
				// TODO realize more flexibility
			}
		case inp.Triangle:
			switch element.FE.AmountNodes {
			case 3:
				for _, data := range element.Data {
					var s staad.Shell
					s.Index = data.Index
					s.IPoint = data.IPoint
					std.Shells = append(std.Shells, s)
				}
			default:
				// TODO realize more flexibility
			}
		case inp.Quadratic:
			switch element.FE.AmountNodes {
			case 4:
				for _, data := range element.Data {
					var s staad.Shell
					s.Index = data.Index
					s.IPoint = data.IPoint
					std.Shells = append(std.Shells, s)
				}
			default:
				// TODO realize more flexibility
			}
		}
	}
	return
}
