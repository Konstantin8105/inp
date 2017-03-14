package convertorInp

import "fmt"

//------------------------------------------
// STD format
// STAAD SPACE
// START JOB INFORMATION
// ENGINEER DATE 07-Mar-17
// END JOB INFORMATION
// INPUT WIDTH 79
// UNIT METER NEWTON
// JOINT COORDINATES
// 1 0 3 0;
// 2 1.75 2.3 0;
// MEMBER INCIDENCES
// 1 2 1609;
// 2 3 4;
// ELEMENT INCIDENCES SHELL
// 15 4 14 13 3;
// 19 14 16 15 13;
// 4310 762 1604 763;
//------------------------------------------
type stdCoordinate struct {
	index uint64
	coord [3]float64
}

type stdBeam struct {
	index  uint64
	iPoint [2]uint64
}

type stdShell struct {
	index  uint64
	iPoint []uint64
}

type stdFormat struct {
	points  []stdCoordinate
	members []stdBeam
	shells  []stdShell
}

func (std stdFormat) getLines() (lines []string) {
	// STD preambula
	lines = append(lines, "STAAD SPACE")
	lines = append(lines, "START JOB INFORMATION")
	lines = append(lines, "ENGINEER DATE 07-Mar-17")
	lines = append(lines, "END JOB INFORMATION")
	lines = append(lines, "INPUT WIDTH 79")
	lines = append(lines, "UNIT METER NEWTON")

	lines = append(lines, "JOINT COORDINATES")
	for _, p := range std.points {
		lines = append(lines, fmt.Sprintf("%v %.10e %.10e %.10e;", p.index, p.coord[0], p.coord[1], p.coord[2]))
	}

	lines = append(lines, "MEMBER INCIDENCES")
	for _, p := range std.members {
		lines = append(lines, fmt.Sprintf("%v %v %v;", p.index, p.iPoint[0], p.iPoint[1]))
	}

	lines = append(lines, "ELEMENT INCIDENCES SHELL")
	for _, p := range std.shells {
		s := fmt.Sprintf("%v ", p.index)
		for _, e := range p.iPoint {
			s = fmt.Sprintf("%s %v", s, e)
		}
		s = fmt.Sprintf("%s ;", s)
		lines = append(lines, s)
	}

	lines = append(lines, "FINISH")

	return lines
}
