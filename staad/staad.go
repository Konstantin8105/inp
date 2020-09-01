package staad

import (
	"fmt"
	"sort"
	"strings"
)

// Coordinate - coordinate
type Coordinate struct {
	Index int
	Coord [3]float64
}

// Beam - staad beam
type Beam struct {
	Index  int
	IPoint [2]int
}

// Shell - staad triangle or quadroelement
type Shell struct {
	Index  int
	IPoint []int
}

// Format - summary format of staad data
type Format struct {
	Points  []Coordinate
	Members []Beam
	Shells  []Shell
}

// GetLines - convert staad format to strings lines
func (std Format) String() string {
	var lines []string

	// STD preambula
	lines = append(lines, "STAAD SPACE")
	lines = append(lines, "START JOB INFORMATION")
	lines = append(lines, "ENGINEER DATE 07-Mar-17")
	lines = append(lines, "END JOB INFORMATION")
	lines = append(lines, "INPUT WIDTH 79")
	lines = append(lines, "UNIT METER NEWTON")

	// points
	lines = append(lines, "JOINT COORDINATES")
	sort.SliceStable(std.Points, func(i, j int) bool {
		return std.Points[i].Index < std.Points[j].Index
	})
	for _, p := range std.Points {
		lines = append(lines,
			fmt.Sprintf("%v %.10e %.10e %.10e;",
				p.Index, p.Coord[0], p.Coord[1], p.Coord[2]))
	}

	// members
	if len(std.Members) != 0 {
		sort.SliceStable(std.Members, func(i, j int) bool {
			return std.Members[i].Index < std.Members[j].Index
		})
		lines = append(lines, "MEMBER INCIDENCES")
		for _, p := range std.Members {
			lines = append(lines,
				fmt.Sprintf("%v %v %v;",
					p.Index, p.IPoint[0], p.IPoint[1]))
		}
	}

	// shells
	if len(std.Shells) != 0 {
		sort.SliceStable(std.Shells, func(i, j int) bool {
			return std.Shells[i].Index < std.Shells[j].Index
		})
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

	// finish
	lines = append(lines, "FINISH")

	return strings.Join(lines, "\n")
}
