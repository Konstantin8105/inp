package staad

import (
	"fmt"
	"sort"
)

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

// Coordinate - coordinate
type Coordinate struct {
	Index int
	Coord [3]float64
}

type pCoordinate []Coordinate

func (a pCoordinate) Len() int           { return len(a) }
func (a pCoordinate) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a pCoordinate) Less(i, j int) bool { return a[i].Index < a[j].Index }

// Beam - staad beam
type Beam struct {
	Index  int
	IPoint [2]int
}

type pBeam []Beam

func (a pBeam) Len() int           { return len(a) }
func (a pBeam) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a pBeam) Less(i, j int) bool { return a[i].Index < a[j].Index }

// Shell - staad triangle or quadroelement
type Shell struct {
	Index  int
	IPoint []int
}

type pShell []Shell

func (a pShell) Len() int           { return len(a) }
func (a pShell) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a pShell) Less(i, j int) bool { return a[i].Index < a[j].Index }

// Format - summary format of staad data
type Format struct {
	Points  []Coordinate
	Members []Beam
	Shells  []Shell
}

// GetLines - convert staad format to strings lines
func (std Format) GetLines() (lines []string) {
	// STD preambula
	lines = append(lines, "STAAD SPACE")
	lines = append(lines, "START JOB INFORMATION")
	lines = append(lines, "ENGINEER DATE 07-Mar-17")
	lines = append(lines, "END JOB INFORMATION")
	lines = append(lines, "INPUT WIDTH 79")
	lines = append(lines, "UNIT METER NEWTON")

	lines = append(lines, "JOINT COORDINATES")
	// sorting
	sort.Sort(pCoordinate(std.Points))
	for _, p := range std.Points {
		lines = append(lines, fmt.Sprintf("%v %.10e %.10e %.10e;", p.Index, p.Coord[0], p.Coord[1], p.Coord[2]))
	}

	if len(std.Members) != 0 {
		// sorting
		sort.Sort(pBeam(std.Members))
		lines = append(lines, "MEMBER INCIDENCES")
		for _, p := range std.Members {
			lines = append(lines, fmt.Sprintf("%v %v %v;", p.Index, p.IPoint[0], p.IPoint[1]))
		}
	}

	if len(std.Shells) != 0 {
		// sorting
		sort.Sort(pShell(std.Shells))
		lines = append(lines, "ELEMENT INCIDENCES SHELL")
		for _, p := range std.Shells {
			s := fmt.Sprintf("%v ", p.Index)
			for _, e := range p.IPoint {
				s = fmt.Sprintf("%s %v", s, e)
			}
			s = fmt.Sprintf("%s ;", s)
			lines = append(lines, s)
		}
	}

	lines = append(lines, "FINISH")

	return lines
}