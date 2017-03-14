package convertorInp

func convertInpToStd(inp inpFormat) (std stdFormat) {
	for _, node := range inp.nodes {
		coord := stdCoordinate{index: node.index, coord: node.coord}
		std.points = append(std.points, coord)
	}

	for _, element := range inp.elements {
		switch element.elType {
		case inpTypeT3D2:
			// member
			var b stdBeam
			b.index = element.index
			for i := 0; i < 2; i++ {
				b.iPoint[i] = element.iPoint[i]
			}
			std.members = append(std.members, b)
		case inpTypeCPS3:
			var s stdShell
			s.index = element.index
			s.iPoint = element.iPoint
			std.shells = append(std.shells, s)
		}
	}
	return
}
