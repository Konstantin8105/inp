package inp

// *Heading
//  /tmp/example972865916/tmpfile.inp
// *NODE
// *ELEMENT, type=T3D2, ELSET=Line1
// *ELEMENT, type=CPS3, ELSET=Surface1
// *ELSET,ELSET=fix_1
// *ELSET,ELSET=shell3

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/Konstantin8105/errors"
)

// Format - summary inp format
type Format struct {
	Heading  string
	Nodes    []Node
	Elements []Element
	Nsets    []Set
	Elsets   []Set
	Material struct {
		Name      string
		Density   float64
		Expansion float64
		Elastic   struct {
			E float64
			v float64
		}
	}
	ShellSections []ShellSection
	Boundaries    []Boundary
	Step          struct {
		IsStatic bool
		Static   struct {
			TimeInc    float64
			TimePeriod float64
		}
		Plastic struct {
			Hardening string
			Data      [10]struct {
				StressVonMises float64
				PlasticStrain  float64
				Temperature    float64
			}
		}

		Nlgeom bool // genuine nonlinear geometric calculation
		Inc    int  // The maximum number of increments in the step (for automatic
		// incrementation) can be specified by using
		// the parameter INC (default is 100)

		Buckle     int
		NodeFiles  []Print
		ElFiles    []Print
		NodePrints []Print
		ElPrints   []Print
		Loads      []Load
		TimePoint  struct {
			Name      string
			Generate  bool
			TimeStart float64
			TimeEnd   float64
			TimeInc   float64
		}
	}
}

func (f Format) String() string {
	var buf bytes.Buffer

	if f.Heading != "" {
		fmt.Fprintf(&buf, "*Heading\n")
		fmt.Fprintf(&buf, " %s\n", f.Heading)
	}
	if len(f.Nodes) > 0 {
		addHeader := true
		for pos, node := range f.Nodes {
			if addHeader {
				fmt.Fprintf(&buf, "*NODE")
				if node.Nodeset != "" {
					fmt.Fprintf(&buf, ",NSET=%s", node.Nodeset)
				}
				fmt.Fprintf(&buf, "\n")
				addHeader = false
			}
			fmt.Fprintf(&buf, "%5d, %+.12e, %+.12e, %+.12e\n",
				node.Index, node.Coord[0], node.Coord[1], node.Coord[2])
			if pos != len(f.Nodes)-1 {
				if f.Nodes[pos].Nodeset != f.Nodes[pos+1].Nodeset {
					addHeader = true
				}
			}
		}
	}
	if len(f.Elements) > 0 {
		addHeader := true
		for pos, el := range f.Elements {
			if addHeader {
				fmt.Fprintf(&buf, "*ELEMENT")
				if el.Type != "" {
					fmt.Fprintf(&buf, ", type=%s", el.Type)
				}
				if el.Elset != "" {
					fmt.Fprintf(&buf, ", ELSET=%s", el.Elset)
				}
				fmt.Fprintf(&buf, "\n")
				addHeader = false
			}
			fmt.Fprintf(&buf, "%5d,", el.Index)
			for pos, v := range el.Nodes {
				fmt.Fprintf(&buf, " %5d", v)
				if pos != len(el.Nodes)-1 {
					fmt.Fprintf(&buf, ",")
				} else {
					fmt.Fprintf(&buf, "\n")
				}
			}
			if pos != len(f.Elements)-1 {
				if f.Elements[pos].Type != f.Elements[pos+1].Type {
					addHeader = true
				}
				if f.Elements[pos].Elset != f.Elements[pos+1].Elset {
					addHeader = true
				}
			}
		}
	}
	if len(f.Nsets) > 0 {
		addHeader := true
		for pos, el := range f.Nsets {
			if addHeader {
				fmt.Fprintf(&buf, "*NSET")
				if el.Name != "" {
					fmt.Fprintf(&buf, ", NSET=%s", el.Name)
				}
				if el.Generate {
					fmt.Fprintf(&buf, ", GENERATE")
				}
				fmt.Fprintf(&buf, "\n")
				addHeader = false
			}
			for _, ind := range el.Indexes {
				fmt.Fprintf(&buf, "%5d,\n", ind)
			}
			if pos != len(f.Nsets)-1 {
				if f.Nsets[pos].Name != f.Nsets[pos+1].Name {
					addHeader = true
				}
				if f.Nsets[pos].Generate != f.Nsets[pos+1].Generate {
					addHeader = true
				}
			}
		}
	}
	if len(f.Elsets) > 0 {
		addHeader := true
		for pos, el := range f.Elsets {
			if addHeader {
				fmt.Fprintf(&buf, "*ELSET")
				if el.Name != "" {
					fmt.Fprintf(&buf, ", ELSET=%s", el.Name)
				}
				if el.Generate {
					fmt.Fprintf(&buf, ", GENERATE")
				}
				fmt.Fprintf(&buf, "\n")
				addHeader = false
			}
			for _, ind := range el.Indexes {
				fmt.Fprintf(&buf, "%5d,\n", ind)
			}
			if pos != len(f.Elsets)-1 {
				if f.Elsets[pos].Name != f.Elsets[pos+1].Name {
					addHeader = true
				}
				if f.Elsets[pos].Generate != f.Elsets[pos+1].Generate {
					addHeader = true
				}
			}
		}
	}
	if f.Material.Name != "" {
		fmt.Fprintf(&buf, "*MATERIAL, NAME=%s\n", f.Material.Name)
	}
	if f.Material.Elastic.E != 0.0 {
		fmt.Fprintf(&buf, "*ELASTIC\n%.8e, %.8e\n",
			f.Material.Elastic.E,
			f.Material.Elastic.v,
		)
	}
	fmt.Fprintf(&buf, "*EXPANSION\n%.8e\n", f.Material.Expansion)
	fmt.Fprintf(&buf, "*DENSITY\n%.8e\n", f.Material.Density)

	for _, ss := range f.ShellSections {
		if ss.Elements == "" {
			continue
		}
		fmt.Fprintf(&buf, "*SHELL SECTION")
		fmt.Fprintf(&buf, ", ELSET=%s", ss.Elements)
		fmt.Fprintf(&buf, ", OFFSET=%f", ss.Offset)
		if ss.Composite {
			fmt.Fprintf(&buf, ", COMPOSITE")
			fmt.Fprintf(&buf, "\n")
			for _, row := range ss.Property {
				fmt.Fprintf(&buf, "%.8e,, %s\n", row.Thickness, row.Material)
			}
		} else {
			fmt.Fprintf(&buf, ", MATERIAL=%s", ss.Property[0].Material)
			fmt.Fprintf(&buf, "\n")
			fmt.Fprintf(&buf, "%.8f\n", ss.Property[0].Thickness)
		}
	}

	fmt.Fprintf(&buf, "\n*STEP")
	if f.Step.Nlgeom {
		fmt.Fprintf(&buf, ", NLGEOM")
	}
	if f.Step.Inc != 0 {
		fmt.Fprintf(&buf, ", INC=%d", f.Step.Inc)
	}
	fmt.Fprintf(&buf, "\n")

	if f.Step.IsStatic {
		fmt.Fprintf(&buf, "*STATIC\n")
		if f.Step.Static.TimeInc != 0.0 || f.Step.Static.TimePeriod != 0.0 {
			fmt.Fprintf(&buf, "%.8e, %.8e\n",
				f.Step.Static.TimeInc, f.Step.Static.TimePeriod)
		}
	}

	if f.Step.Plastic.Hardening != "" {
		fmt.Fprintf(&buf, "*PLASTIC, HARDENING=%s\n", f.Step.Plastic.Hardening)
		for _, d := range f.Step.Plastic.Data {
			if d.StressVonMises != 0.0 {
				fmt.Fprintf(&buf, "%.8e, %.8e, %.8e\n",
					d.StressVonMises, d.PlasticStrain, d.Temperature)
			}
		}
	}

	if f.Step.Buckle != 0 {
		fmt.Fprintf(&buf, "*BUCKLE\n%d\n", f.Step.Buckle)
	}

	for _, boun := range f.Boundaries {
		fmt.Fprintf(&buf, "*BOUNDARY\n%s,%d,%d,%.8e\n",
			boun.LoadLocation, boun.Start, boun.Finish, boun.Factor,
		)
	}

	fmt.Fprintf(&buf, "\n*END STEP\n")

	return buf.String()
}

// isHeader return true for example:
// if prefix = "*NODE", but not "*NODE PRINT"
//
// Example of line:
// *NODE
// *NODE, somethinks
// *NODE PRINT
// *NODE PRINT, simethink
func isHeader(line, prefix string) bool {
	index := strings.Index(line, ",")
	if index > 0 {
		line = line[:index]
	}
	line = strings.TrimSpace(line)
	prefix = strings.TrimSpace(prefix)
	return prefix == line
}

func (f *Format) parseHeading(block []string) (ok bool, err error) {
	if !isHeader(block[0], "*HEADING") {
		return false, nil
	}
	if len(block) == 1 {
		f.Heading = block[1]
	}
	return true, nil
}

// Node - coordinate in inp format
type Node struct {
	Nodeset string // NSET
	Index   int
	Coord   [3]float64
}

// parseNode
//
// Examples:
//		*NODE
//		1, 0, 0, 0
//
//		*NODE, NSET=Nall
//		1, 0, 0, 0
// First line:
//   *NODE
//   Enter the optional parameter, if desired.
// Following line:
//   node number.
//   Value of first coordinate.
//   Value of second coordinate.
//   Value of third coordinate.
func (f *Format) parseNode(block []string) (ok bool, err error) {
	if !isHeader(block[0], "*NODE") {
		return false, nil
	}

	// NSET
	var nodeset string
	{
		block[0] = strings.Replace(block[0], ",", " ", -1)
		fields := strings.Fields(block[0])
		if len(fields) > 1 {
			for _, f := range fields {
				if strings.HasPrefix(f, "NSET=") {
					nodeset = f[5:]
				}
			}
		}
	}

	for _, line := range block[1:] {
		line = strings.Replace(line, ",", " ", -1)
		fields := strings.Fields(line)
		if len(fields) != 4 {
			err = fmt.Errorf("not valid fields: %s", line)
			return
		}

		var index int64
		var coord [3]float64

		index, err = strconv.ParseInt(fields[0], 10, 64)
		if err != nil {
			return
		}
		for i := 0; i < 3; i++ {
			coord[i], err = strconv.ParseFloat(fields[i+1], 64)
			if err != nil {
				return
			}
		}
		f.Nodes = append(f.Nodes, Node{
			Nodeset: nodeset,
			Index:   int(index),
			Coord:   coord,
		})
	}
	return true, nil
}

// Element - indexes in inp format
type Element struct {
	Type  string
	Elset string
	Index int
	Nodes []int
}

// parseElement - parser for ELEMENT
//
// First line:
// 	*ELEMENT
// 	Enter any needed parameters and their values.
// Following line:
// 	Element number.
// 	Node numbers forming the element. The order of nodes around the element is
//	given in section 2.1. Use continuation lines for elements having more
//	than 15 nodes (maximum 16 entries per line).
func (f *Format) parseElement(block []string) (ok bool, err error) {
	if !isHeader(block[0], "*ELEMENT") {
		return false, nil
	}

	var Type, Elset string
	{
		block[0] = strings.Replace(block[0], ",", " ", -1)
		fields := strings.Fields(block[0])
		for _, f := range fields {
			if strings.HasPrefix(f, "TYPE=") {
				Type = f[5:]
			}
			if strings.HasPrefix(f, "ELSET=") {
				Elset = f[6:]
			}
		}
	}

	for _, line := range block[1:] {
		line = strings.Replace(line, ",", " ", -1)
		fields := strings.Fields(line)
		var ints []int
		for _, f := range fields {
			var i64 int64
			i64, err = strconv.ParseInt(f, 10, 64)
			if err != nil {
				return
			}
			ints = append(ints, int(i64))
		}
		f.Elements = append(f.Elements, Element{
			Type:  Type,
			Elset: Elset,
			Index: ints[0],
			Nodes: ints[1:],
		})
	}
	return true, nil
}

type Set struct {
	Name     string
	Generate bool
	Indexes  []int
}

func (f *Format) parseSet(s *[]Set, prefix string, block []string) (ok bool, err error) {
	if !isHeader(block[0], "*"+prefix) {
		return false, nil
	}

	var set Set
	for _, s := range strings.Split(block[0], ",")[1:] {
		s = strings.TrimSpace(s)
		switch {
		case strings.HasPrefix(s, "ELSET="):
			set.Name = s[6:]
		case strings.HasPrefix(s, "NSET="):
			set.Name = s[5:]
		case s == "GENERATE":
			set.Generate = true
		default:
			panic(block[0])
		}
	}
	for _, line := range block[1:] {
		line = strings.Replace(line, ",", " ", -1)
		fields := strings.Fields(line)
		for _, f := range fields {
			var i64 int64
			i64, err = strconv.ParseInt(f, 10, 64)
			if err != nil {
				return
			}
			set.Indexes = append(set.Indexes, int(i64))
		}
	}
	(*s) = append((*s), set)

	return true, nil
}

func (f *Format) parseDensity(block []string) (ok bool, err error) {
	if !isHeader(block[0], "*DENSITY") {
		return false, nil
	}
	f.Material.Density, err = strconv.ParseFloat(block[1], 64)
	if err != nil {
		return
	}
	return true, nil
}

func (f *Format) parseExpansion(block []string) (ok bool, err error) {
	if !isHeader(block[0], "*EXPANSION") {
		return false, nil
	}
	f.Material.Expansion, err = strconv.ParseFloat(block[1], 64)
	if err != nil {
		return
	}
	return true, nil
}

func (f *Format) parseMaterial(block []string) (ok bool, err error) {
	if !isHeader(block[0], "*MATERIAL") {
		return false, nil
	}
	line := block[0]
	line = strings.Replace(line, ",", " ", -1)
	fields := strings.Fields(line)
	if len(fields) > 1 {
		for _, field := range fields[1:] {
			switch {
			case strings.HasPrefix(field, "NAME="):
				f.Material.Name = field[5:]
			default:
				panic(fmt.Errorf("`%s` : `%s`", line, field))
			}
		}
	}
	return true, nil
}

func (f *Format) parseElastic(block []string) (ok bool, err error) {
	if !isHeader(block[0], "*ELASTIC") {
		return false, nil
	}
	line := strings.Replace(block[1], ",", " ", -1)
	fields := strings.Fields(line)
	f.Material.Elastic.E, err = strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return
	}
	f.Material.Elastic.v, err = strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return
	}

	return true, nil
}

// for structures:
// – 1: translation in the local x-direction
// – 2: translation in the local y-direction
// – 3: translation in the local z-direction
// – 4: rotation about the local x-axis (only for nodes belonging to beams or shells)
// – 5: rotation about the local y-axis (only for nodes belonging to beams or shells)
// – 6: rotation about the local z-axis (only for nodes belonging to beams or shells)
// – 11: temperature
//
// First line:
// 		*BOUNDARY
// 		Enter any needed parameters and their value.
// Following line:
// 		Node number or node set label
// 		First degree of freedom constrained
// 		Last degree of freedom constrained. This field may be left blank if only one degree of freedom is constrained.
type Boundary struct {
	LoadLocation string
	Start        int
	Finish       int
	Factor       float64
}

func (f *Format) parseBoundary(block []string) (ok bool, err error) {
	if !isHeader(block[0], "*BOUNDARY") {
		return false, nil
	}
	for _, line := range block[1:] {
		line = strings.Replace(line, ",", " ", -1)
		fields := strings.Fields(line)
		var b Boundary
		b.LoadLocation = fields[0]

		var i64 int64

		if len(fields) > 1 {
			i64, err = strconv.ParseInt(fields[1], 10, 64)
			if err != nil {
				return
			}
			b.Start = int(i64)
		}

		if len(fields) > 2 {
			i64, err = strconv.ParseInt(fields[2], 10, 64)
			if err != nil {
				return
			}
			b.Finish = int(i64)
		}

		if len(fields) > 3 {
			b.Factor, err = strconv.ParseFloat(fields[3], 64)
			if err != nil {
				return
			}
		}

		f.Boundaries = append(f.Boundaries, b)
	}

	return true, nil
}

type ShellSection struct {
	Elements  string
	Offset    float64
	Composite bool
	Property  [12]struct {
		Thickness float64
		Material  string
	}
}

// *SHELL SECTION,MATERIAL=steel,ELSET=Eall,,OFFSET=0
// 6.2500E-02
func (f *Format) parseShellSection(block []string) (ok bool, err error) {
	if !isHeader(block[0], "*SHELL SECTION") {
		return false, nil
	}
	var ss ShellSection
	split := strings.Split(block[0], ",")[1:]
	for _, s := range split {
		s = strings.TrimSpace(s)
		switch {
		case strings.HasPrefix(s, "MATERIAL"):
			index := strings.Index(s, "=")
			s = strings.TrimSpace(s[index+1:])
			ss.Property[0].Material = s
		case strings.HasPrefix(s, "ELSET"):
			index := strings.Index(s, "=")
			s = strings.TrimSpace(s[index+1:])
			ss.Elements = s
		case strings.HasPrefix(s, "OFFSET"):
			index := strings.Index(s, "=")
			s = strings.TrimSpace(s[index+1:])
			ss.Offset, err = strconv.ParseFloat(s, 64)
			if err != nil {
				return
			}
		case s == "":
			// do nothing
		case s == "COMPOSITE":
			ss.Composite = true
		default:
			panic(fmt.Errorf("%s", strings.Join(split, "|")))
		}
	}
	if ss.Composite {
		for pos, line := range block[1:] {
			line = strings.Replace(line, ",", " ", -1)
			fields := strings.Fields(line)
			ss.Property[pos].Thickness, err = strconv.ParseFloat(fields[0], 64)
			if err != nil {
				err = fmt.Errorf("%v : %v", block, err)
				return
			}
			ss.Property[pos].Material = fields[1]
		}
	} else {
		line := strings.TrimSpace(block[1])
		ss.Property[0].Thickness, err = strconv.ParseFloat(line, 64)
		if err != nil {
			err = fmt.Errorf("%v : %v", block, err)
			return
		}
	}

	f.ShellSections = append(f.ShellSections, ss)

	return true, nil
}

// Examples:
//
// *STEP
// *STEP,NLGEOM
// *STEP,INC=100,NLGEOM
func (f *Format) parseStep(block []string) (ok bool, err error) {
	if !isHeader(block[0], "*STEP") {
		return false, nil
	}
	if len(block) > 1 {
		panic("*STEP parameter")
	}
	for _, part := range strings.Split(block[0], ",")[1:] {
		part = strings.TrimSpace(part)
		switch {
		case part == "NLGEOM":
			f.Step.Nlgeom = true
		case strings.HasPrefix(part, "INC="):
			part = part[4:]
			var i64 int64
			i64, err = strconv.ParseInt(part, 10, 64)
			if err != nil {
				return
			}
			f.Step.Inc = int(i64)
		default:
			panic(part)
		}
	}

	return true, nil
}

// First line:
// 		*BUCKLE
// Second line:
// 		Number of buckling factors desired (usually 1).
// 		Accuracy desired (default: 0.01).
// 		# Lanczos vectors calculated in each iteration (default: 4 * #eigenvalues).
// 		Maximum # of iterations (default: 1000).
func (f *Format) parseBuckle(block []string) (ok bool, err error) {
	if !isHeader(block[0], "*BUCKLE") {
		return false, nil
	}
	if len(block) > 2 {
		panic("*STEP parameter")
	}
	part := strings.TrimSpace(block[1])
	var i64 int64
	i64, err = strconv.ParseInt(part, 10, 64)
	if err != nil {
		return
	}
	f.Step.Buckle = int(i64)

	return true, nil
}

type File struct {
	Options []string
}

type Print struct {
	SetName        string
	Frequency      int
	Output         string
	TimePoints     string
	ContactElement bool
	Global         bool
	Options        []string
}

// Example:
//
// [*NODE FILE ,TIME POINTS=T1 U,]
// [*EL FILE,TIME POINTS=T1 S,PEEQ,]
//
// MAXU [MDISP]: Maximum displacements orthogonal to a given vector
// at all times for *FREQUENCY calculations with cyclic symmetry. The
// components of the vector are the coordinates of a node stored in a node
// set with the name RAY. This node and node set must have been defined
// by the user.
//
// PU [PDISP]: Displacements: magnitude and phase (only for *STEADY STATE DYNAMICS
// calculations and *FREQUENCY calculations with cyclic symmetry).
//
// RF [FORC(real), FORCI(imaginary)]: External forces (only static forces;
// dynamic forces, such as those caused by dashpots, are not included)
//
// U [DISP(real), DISPI(imaginary)]: Displacements.
//
//
// [*NODE PRINT,NSET=N1 RF]
// [*NODE PRINT,NSET=NALL U,RF]
// [*NODE PRINT,NSET=NALL U]
// [*NODE PRINT,NSET=FIX,TIME POINTS=T1 RF]
// [*NODE PRINT,NSET=LOAD,TIME POINTS=T1 RF]
//
// Displacements (key=U)
//
// External forces (key=RF) (only static forces; dynamic forces, such as those
// caused by dashpots, are not included)
//
// Structural temperatures and total temperatures in networks (key=NT or
// TS; both are equivalent)
//
// Example:
//		*NODE FILE,TIME POINTS=T1
//		RF,NT
// requests the storage of reaction forces and temperatures in the .frd file for
// all time points defined by the T1 time points sequence
func (f *Format) parsePrint(block []string, prefix string, pr *[]Print) (ok bool, err error) {
	if !isHeader(block[0], prefix) {
		return false, nil
	}
	var np Print
	for _, s := range strings.Split(block[0], ",")[1:] {
		s = strings.TrimSpace(s)
		switch {
		case strings.HasPrefix(s, "NSET="):
			s = s[5:]
			np.SetName = s
		case strings.HasPrefix(s, "ELSET="):
			s = s[6:]
			np.SetName = s
		case strings.HasPrefix(s, "GLOBAL="):
			s = s[7:]
			np.Global = s == "YES"
		case strings.HasPrefix(s, "TIME POINTS="):
			s = s[12:]
			np.TimePoints = s
		case strings.HasPrefix(s, "FREQUENCY="):
			s = s[10:]
			var i64 int64
			i64, err = strconv.ParseInt(s, 10, 64)
			if err != nil {
				return
			}
			np.Frequency = int(i64)
		default:
			panic(s)
		}
	}
	np.Options = strings.Fields(strings.Replace(block[1], ",", " ", -1))

	(*pr) = append((*pr), np)

	return true, nil
}

// [*STATIC 0.01,1]
//
// First line:
// • *STATIC
// • Enter any needed parameters and their values.
//
// Second line (only relevant for nonlinear analyses; for linear analyses, the step
// length is always 1)
// • Initial time increment. This value will be modified due to automatic in-
//   crementation, unless the parameter DIRECT was specified (default 1.).
// • Time period of the step (default 1.).
// • Minimum time increment allowed. Only active if DIRECT is not specified.
// Default is the initial time increment or 1.e-5 times the time period of the
// step, whichever is smaller.
// • Maximum time increment allowed. Only active if DIRECT is not specified.
//   Default is 1.e+30
// • Initial time increment for CFD applications (default 1.e-2)
func (f *Format) parseStatic(block []string) (ok bool, err error) {
	if !isHeader(block[0], "*STATIC") {
		return false, nil
	}
	f.Step.IsStatic = true
	if len(block) == 1 {
		return true, nil
	}
	if len(block) != 2 {
		panic(block)
	}
	fields := strings.Fields(strings.Replace(block[1], ",", " ", -1))
	if len(fields) != 2 {
		panic(block)
	}

	f.Step.Static.TimeInc, err = strconv.ParseFloat(fields[0], 64)
	if err != nil {
		err = fmt.Errorf("%v : %v", block, err)
		return
	}

	f.Step.Static.TimePeriod, err = strconv.ParseFloat(fields[1], 64)
	if err != nil {
		err = fmt.Errorf("%v : %v", block, err)
		return
	}

	return true, nil
}

type Load struct {
	Position  string
	Direction int
	Value     float64
}

// [*CLOAD 5, 1, 5000.0]
// [*CLOAD 2,3,0.0025]
// [*CLOAD LOAD,3,-3.3112583E+00]
func (f *Format) parseCload(block []string) (ok bool, err error) {
	if !isHeader(block[0], "*CLOAD") {
		return false, nil
	}
	for _, line := range block[1:] {
		line = strings.Replace(line, ",", " ", -1)
		fields := strings.Fields(line)
		if len(fields) != 3 {
			panic(line)
		}
		var l Load
		l.Position = fields[0]

		var i64 int64
		i64, err = strconv.ParseInt(fields[1], 10, 64)
		if err != nil {
			return
		}
		l.Direction = int(i64)

		l.Value, err = strconv.ParseFloat(fields[2], 64)
		if err != nil {
			return
		}

		f.Step.Loads = append(f.Step.Loads, l)
	}

	return true, nil
}

func (f *Format) parseTimePoint(block []string) (ok bool, err error) {
	if !isHeader(block[0], "*TIME POINTS") {
		return false, nil
	}

	for _, s := range strings.Split(block[0], ",")[1:] {
		s = strings.TrimSpace(s)
		switch {
		case strings.HasPrefix(s, "NAME="):
			s = s[5:]
			f.Step.TimePoint.Name = s
		case strings.HasPrefix(s, "GENERATE"):
			f.Step.TimePoint.Generate = true
		default:
			panic(s)
		}
	}

	line := strings.Replace(block[1], ",", " ", -1)
	fields := strings.Fields(line)

	f.Step.TimePoint.TimeStart, err = strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return
	}
	f.Step.TimePoint.TimeEnd, err = strconv.ParseFloat(fields[1], 64)
	if err != nil {
		return
	}
	f.Step.TimePoint.TimeInc, err = strconv.ParseFloat(fields[2], 64)
	if err != nil {
		return
	}

	return true, nil
}

func (f *Format) parsePlastic(block []string) (ok bool, err error) {
	if !isHeader(block[0], "*PLASTIC") {
		return false, nil
	}

	for _, s := range strings.Split(block[0], ",")[1:] {
		s = strings.TrimSpace(s)
		prefixH := "HARDENING="
		switch {
		case strings.HasPrefix(s, prefixH):
			s = s[len(prefixH):]
			f.Step.Plastic.Hardening = s
		default:
			panic(s)
		}
	}

	for pos, line := range block[1:] {
		line = strings.Replace(line, ",", " ", -1)
		fields := strings.Fields(line)

		f.Step.Plastic.Data[pos].StressVonMises, err = strconv.ParseFloat(fields[0], 64)
		if err != nil {
			return
		}
		f.Step.Plastic.Data[pos].PlasticStrain, err = strconv.ParseFloat(fields[1], 64)
		if err != nil {
			return
		}
		if len(fields) == 2 {
			continue
		}
		f.Step.Plastic.Data[pos].Temperature, err = strconv.ParseFloat(fields[2], 64)
		if err != nil {
			return
		}
	}

	return true, nil
}

func ignore(prefix string) func(block []string) (ok bool, err error) {
	return func(block []string) (ok bool, err error) {
		if !isHeader(block[0], prefix) {
			return false, nil
		}
		return true, nil
	}
}

func Parse(content []byte) (f *Format, err error) {
	// split into lines
	lines := strings.Split(string(content), "\n")

	// split into block
	var blocks [][]string
	for _, s := range lines {
		s = strings.TrimSpace(s)
		s = strings.ToUpper(s)
		if s == "" {
			continue
		}
		if strings.HasPrefix(s, "**") || strings.HasPrefix(s, ">**") {
			continue
		}
		if strings.HasPrefix(s, "*") {
			blocks = append(blocks, []string{})
		}
		blocks[len(blocks)-1] = append(blocks[len(blocks)-1], s)
	}

	{
		counter := 0
		for i := range blocks {
			if len(blocks) == 0 {
				continue
			}
			if isHeader(blocks[i][0], "*STEP") {
				counter++
			}
		}
		if 1 < counter {
			panic("counter *STEP is not support")
		}
	}

	// 	for _, bs := range blocks {
	// 		fmt.Println(	"------------")
	// 		for _,s := range bs {
	// 			fmt.Println(	s)
	// 		}
	// 	}

	// parsing
	f = new(Format)

	et := errors.New("parsing errors")

	for _, block := range blocks {
		var found bool
		for parserPos, parser := range []func(block []string) (ok bool, err error){
			f.parseNode,
			f.parseHeading,
			f.parseElement,
			func(block []string) (ok bool, err error) {
				return f.parseSet(&(f.Nsets), "NSET", block)
			},
			func(block []string) (ok bool, err error) {
				return f.parseSet(&(f.Elsets), "ELSET", block)
			},
			f.parseDensity,
			f.parseExpansion,
			f.parseElastic,
			f.parseBoundary,
			f.parseMaterial,
			ignore("*SURFACE"),
			f.parseShellSection,
			f.parseStep,
			f.parseBuckle,
			f.parseStatic,
			f.parsePlastic,
			func(block []string) (ok bool, err error) {
				return f.parsePrint(block, "*NODE FILE", &(f.Step.NodeFiles))
			},
			func(block []string) (ok bool, err error) {
				return f.parsePrint(block, "*EL FILE", &(f.Step.ElFiles))
			},
			func(block []string) (ok bool, err error) {
				return f.parsePrint(block, "*NODE PRINT", &(f.Step.NodePrints))
			},
			func(block []string) (ok bool, err error) {
				return f.parsePrint(block, "*EL PRINT", &(f.Step.ElPrints))
			},
			f.parseCload,
			f.parseTimePoint,
			ignore("*END STEP"),
			// ignore("*HEAT TRANSFER"),
			// ignore("*CONDUCTIVITY"),
			// ignore("*FLUID"),
			// ignore("*SPECIFIC GAS CONSTANT"),
			// ignore("*SPECIFIC HEAT"),
			// ignore("*PHYSICAL CONSTANTS"),
		} {
			if len(block) == 0 {
				continue
			}
			var ok bool
			ok, err = parser(block)
			if err != nil {
				et.Add(fmt.Errorf("№ %d: %v", parserPos, err))
				continue
			}
			found = found || ok
		}
		if !found {
			if len(block) > 3 {
				block = block[:3]
			}
			err = fmt.Errorf("Not found block : %v", strings.Join(block, "\n"))
			et.Add(err)
		}
	}

	if et.IsError() {
		err = et
		return
	}

	return f, nil
}

// ParseBucklingFactor in file for example `shell2.dat` and return
// slice of buckling factors.
//
//      B U C K L I N G   F A C T O R   O U T P U T
//
//  MODE NO       BUCKLING
//                 FACTOR
//
//       1   0.4185108E+03
//       2   0.4196190E+03
//       3   0.4200342E+03
//       4   0.4212441E+03
func ParseBucklingFactor(content []byte) (factors []float64, err error) {
	for _, line := range strings.Split(string(content), "\n")[5:] {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 2 {
			panic(line)
		}
		var factor float64
		factor, err = strconv.ParseFloat(fields[0], 64)
		if err != nil {
			return
		}
		factors = append(factors, factor)
	}
	return
}

// lineGroup - group of points
//  *------*------*
//  A      C      B
// points A,B - exist index of points
// point C - new point
// type lineGroup struct {
// 	indexA, indexB int
// 	nodeC          Node
// }
//
// // for sorting by indexA
// type byIndexA []lineGroup

// func (l byIndexA) Len() int {
// 	return len(l)
// }
// func (l byIndexA) Swap(i, j int) {
// 	l[i], l[j] = l[j], l[i]
// }
// func (l byIndexA) Less(i, j int) bool {
// 	return l[i].indexA < l[j].indexA
// }

// ChangeTypeFiniteElement - change type finite element for example
// from S4 to S8
// func (f *Format) ChangeTypeFiniteElement(from *FiniteElement, to *FiniteElement) (err error) {
// 	if from == to {
// 		return nil
// 	}
//
// 	if from.Shape == to.Shape && from.AmountNodes == to.AmountNodes {
// 		// modify finite element with middle point
// 		for elemenentI := range f.Elements {
// 			if f.Elements[elemenentI].FE.Name != from.Name {
// 				continue
// 			}
// 			f.Elements[elemenentI].FE = to
// 		}
// 		return nil
// 	}
//
// 	if from.Shape != to.Shape {
// 		if from.AmountNodes == 4 && to.AmountNodes == 3 {
// 			s3, _ := GetFiniteElementByName("S3")
// 			f.changeFEfromQuadraticToTriangle(from, s3)
// 			return nil
// 		}
// 		if from.AmountNodes == 4 && to.AmountNodes == 6 {
// 			s3, _ := GetFiniteElementByName("S3")
// 			err = f.ChangeTypeFiniteElement(from, s3)
// 			if err != nil {
// 				return err
// 			}
// 			err = f.ChangeTypeFiniteElement(s3, to)
// 			if err != nil {
// 				return err
// 			}
// 			return nil
// 		}
// 	}
//
// 	if from.Shape == to.Shape && from.AmountNodes*2 == to.AmountNodes {
//
// 		// divide middle point inside exist
// 		group, err := f.createMiddlePoint(from)
// 		if err != nil {
// 			return fmt.Errorf("Wrong in createMiddlePoint: %v", err)
// 		}
//
// 		// add points in format
// 		for _, node := range group {
// 			f.Nodes = append(f.Nodes, node.nodeC)
// 		}
//
// 		// modify finite element with middle point
// 		for elemenentI := range f.Elements {
// 			if f.Elements[elemenentI].FE.Name != from.Name {
// 				continue
// 			}
// 			f.Elements[elemenentI].FE = to
// 			for iData := range f.Elements[elemenentI].Data {
// 				iPoints := f.Elements[elemenentI].Data[iData].IPoint
// 				// modification
// 				var newPoints []int
// 				for index := range iPoints {
// 					var pointIndex1 int
// 					if index == 0 {
// 						pointIndex1 = iPoints[len(iPoints)-1]
// 					} else {
// 						pointIndex1 = iPoints[index-1]
// 					}
// 					pointIndex2 := iPoints[index]
// 					var newPoint int
// 					if pointIndex1 > pointIndex2 {
// 						newPoint, err = f.foundPointCIndexInLineGroup(pointIndex2, pointIndex1, &group)
// 					} else {
// 						newPoint, err = f.foundPointCIndexInLineGroup(pointIndex1, pointIndex2, &group)
// 					}
// 					if err != nil {
// 						return fmt.Errorf("Cannot found point in lineGroup : %v", err)
// 					}
// 					newPoints = append(newPoints, newPoint)
// 				}
// 				// end of modification
// 				for i := range newPoints {
// 					if i == len(newPoints)-1 {
// 						f.Elements[elemenentI].Data[iData].IPoint = append(f.Elements[elemenentI].Data[iData].IPoint, newPoints[0])
// 					} else {
// 						f.Elements[elemenentI].Data[iData].IPoint = append(f.Elements[elemenentI].Data[iData].IPoint, newPoints[i+1])
// 					}
// 				}
// 			}
// 		}
//
// 		// NodeNames changes
// 		if len(f.NodesWithName) != 0 {
// 			return fmt.Errorf("Cannot work with Named nodes")
// 		}
//
// 		return nil
// 	}
//
// 	return fmt.Errorf("Cannot change FE from %v to %v", from, to)
// }

// func (f *Format) foundPointCIndexInLineGroup(p1, p2 int, group *[]lineGroup) (middlePoint int, err error) {
// 	if p1 > p2 {
// 		return -1, fmt.Errorf("Case p1 < p2 is not correct")
// 	}
// 	for _, g := range *group {
// 		if g.indexA == p1 && g.indexB == p2 {
// 			return g.nodeC.Index, nil
// 		}
// 	}
// 	return -1, fmt.Errorf("Cannot found in group with point %v,%v\nGroup = %v", p1, p2, *group)
// }

// func (f *Format) createMiddlePoint(fe *FiniteElement) (group []lineGroup, err error) {
// 	// check slice of nodes inp format - index must by from less to more
// 	// if it is true, then we can use binary sort for fast found the point
// 	for index := range f.Nodes {
// 		if index == 0 {
// 			continue
// 		}
// 		if f.Nodes[index-1].Index >= f.Nodes[index].Index {
// 			return nil, fmt.Errorf("Please sort the nodes in inp format")
// 		}
// 	}
//
// 	// create slice of linegroup
// 	for _, element := range f.Elements {
// 		if element.FE.Name != fe.Name {
// 			continue
// 		}
//
// 		for _, data := range element.Data {
// 			for index := range data.IPoint {
// 				var pointIndex1 int
// 				if index == 0 {
// 					pointIndex1 = data.IPoint[len(data.IPoint)-1]
// 				} else {
// 					pointIndex1 = data.IPoint[index-1]
// 				}
// 				pointIndex2 := data.IPoint[index]
// 				var g lineGroup
// 				if pointIndex1 > pointIndex2 {
// 					g = lineGroup{indexA: pointIndex2, indexB: pointIndex1}
// 				} else {
// 					g = lineGroup{indexA: pointIndex1, indexB: pointIndex2}
// 				}
// 				group = append(group, g)
// 			}
// 		}
// 	}
//
// 	// sorting linegroup
// 	sort.Sort(byIndexA(group))
// 	for {
// 		var isChange bool
// 		for i := range group {
// 			if i == 0 {
// 				continue
// 			}
// 			if group[i-1].indexA != group[i].indexA {
// 				continue
// 			}
// 			if group[i-1].indexB > group[i].indexB {
// 				// swap
// 				group[i-1].indexB, group[i].indexB = group[i].indexB, group[i-1].indexB
// 				isChange = true
// 			}
// 		}
// 		if !isChange {
// 			break
// 		}
// 	}
//
// 	// create unique slice : true - if unique
// 	unique := make([]bool, len(group), len(group))
// 	for index := range group {
// 		if index == 0 {
// 			unique[0] = true
// 			continue
// 		}
// 		unique[index] = !(group[index-1].indexA == group[index].indexA && group[index-1].indexB == group[index].indexB)
// 	}
//
// 	amount := 0
// 	for _, u := range unique {
// 		if u {
// 			amount++
// 		}
// 	}
//
// 	// create unique linegroup
// 	var buffer []lineGroup
// 	for i, u := range unique {
// 		if u {
// 			buffer = append(buffer, group[i])
// 		}
// 	}
// 	group = buffer
//
// 	// 2-step for calculate middle point
// 	for index := range group {
//
// 		// step 1: loop - add to nodeC coordinate of NodeA
// 		group[index].nodeC.Coord, err = f.foundByIndex(group[index].indexA)
// 		if err != nil {
// 			return nil, fmt.Errorf("Cannot found point with index : %v", group[index].indexA)
// 		}
// 		// step 2: loop - calculate nodeC = (nodeC+nodeB)/2.
// 		coord, err := f.foundByIndex(group[index].indexB)
// 		if err != nil {
// 			return nil, fmt.Errorf("Cannot found point with index : %v", group[index].indexB)
// 		}
// 		// calculate middle
// 		for i := 0; i < 3; i++ {
// 			group[index].nodeC.Coord[i] += coord[i]
// 			group[index].nodeC.Coord[i] /= 2.0
// 		}
// 	}
//
// 	// find maximal index of point
// 	maximalIndex := f.Nodes[0].Index
// 	for index := range f.Nodes {
// 		if maximalIndex < f.Nodes[index].Index {
// 			maximalIndex = f.Nodes[index].Index
// 		}
// 	}
// 	maximalIndex++
//
// 	// add index to indexC
// 	for index := range group {
// 		group[index].nodeC.Index = maximalIndex
// 		maximalIndex++
// 	}
//
// 	return group, nil
// }

// func (f *Format) foundByIndex(index int) (node [3]float64, err error) {
// 	i := sort.Search(len(f.Nodes), func(a int) bool { return f.Nodes[a].Index >= index })
// 	if i < len(f.Nodes) && f.Nodes[i].Index == index {
// 		// index is present at nodes
// 		return f.Nodes[i].Coord, nil
// 	}
// 	// index is not present in nodes,
// 	// but i is the index where it would be inserted.
// 	return node, fmt.Errorf("Cannot found in sort.Search : %v, but i = %v", index, i)
// }

// func (f *Format) changeFEfromQuadraticToTriangle(from *FiniteElement, to *FiniteElement) {
// 	var maximalIndex int
// 	for _, element := range f.Elements {
// 		for _, data := range element.Data {
// 			if maximalIndex < data.Index {
// 				maximalIndex = data.Index
// 			}
// 		}
// 	}
// 	maximalIndex++
//
// 	// add new elements
// 	for elemenentI := range f.Elements {
// 		if f.Elements[elemenentI].FE.Name != from.Name {
// 			continue
// 		}
// 		var newElement Element
// 		newElement.Name = f.Elements[elemenentI].Name
// 		newElement.FE = to
// 		for iData := range f.Elements[elemenentI].Data {
// 			// add random dividing for avoid anisotrop finite element model
// 			//if rand.Float64() > 0.5 {
// 			newElement.Data = append(newElement.Data, ElementData{Index: maximalIndex, IPoint: []int{
// 				f.Elements[elemenentI].Data[iData].IPoint[0],
// 				f.Elements[elemenentI].Data[iData].IPoint[1],
// 				f.Elements[elemenentI].Data[iData].IPoint[2],
// 			}})
// 			maximalIndex++
// 			newElement.Data = append(newElement.Data, ElementData{Index: maximalIndex, IPoint: []int{
// 				f.Elements[elemenentI].Data[iData].IPoint[2],
// 				f.Elements[elemenentI].Data[iData].IPoint[3],
// 				f.Elements[elemenentI].Data[iData].IPoint[0],
// 			}})
// 			maximalIndex++
// 			/*} else {
// 				newElement.Data = append(newElement.Data, ElementData{Index: maximalIndex, IPoint: []int{
// 					f.Elements[elemenentI].Data[iData].IPoint[1],
// 					f.Elements[elemenentI].Data[iData].IPoint[2],
// 					f.Elements[elemenentI].Data[iData].IPoint[3],
// 				}})
// 				maximalIndex++
// 				newElement.Data = append(newElement.Data, ElementData{Index: maximalIndex, IPoint: []int{
// 					f.Elements[elemenentI].Data[iData].IPoint[3],
// 					f.Elements[elemenentI].Data[iData].IPoint[0],
// 					f.Elements[elemenentI].Data[iData].IPoint[1],
// 				}})
// 				maximalIndex++
// 			}*/
// 		}
// 		f.Elements = append(f.Elements, newElement)
// 	}
// 	// remove old FE
// AGAIN:
// 	for elemenentI := range f.Elements {
// 		if f.Elements[elemenentI].FE.Name != from.Name {
// 			continue
// 		}
// 		f.Elements = append(f.Elements[:elemenentI], f.Elements[(elemenentI+1):]...)
// 		goto AGAIN
// 	}
//
// 	return
// }
//
// // FiniteElementShape - const of finite element shape
// type FiniteElementShape int
//
// // Shapes of finite element
// const (
// 	Triangle FiniteElementShape = iota
// 	Quadratic
// 	Beam
// )
//
// // FiniteElement - information about finite element
// type FiniteElement struct {
// 	Shape       FiniteElementShape
// 	AmountNodes int
// 	Name        string
// 	Description string
// }
//
// // FiniteElementDatabase - information about all allowable finite elements
// var FiniteElementDatabase []FiniteElement
//
// func init() {
// 	FiniteElementDatabase = []FiniteElement{
// 		FiniteElement{
// 			Shape:       Triangle,
// 			AmountNodes: 3,
// 			Name:        "CPS3",
// 			Description: "Three-node plane stress element",
// 		},
// 		FiniteElement{
// 			Shape:       Beam,
// 			AmountNodes: 2,
// 			Name:        "T3D2",
// 			Description: "Two-node truss element",
// 		},
// 		FiniteElement{
// 			Shape:       Triangle,
// 			AmountNodes: 3,
// 			Name:        "S3",
// 			Description: "Three-node shell element",
// 		},
// 		FiniteElement{
// 			Shape:       Quadratic,
// 			AmountNodes: 4,
// 			Name:        "S4",
// 			Description: "Four-node shell element",
// 		},
// 		FiniteElement{
// 			Shape:       Quadratic,
// 			AmountNodes: 4,
// 			Name:        "S4R",
// 			Description: "Four-node shell element",
// 		},
// 		FiniteElement{
// 			Shape:       Triangle,
// 			AmountNodes: 6,
// 			Name:        "S6",
// 			Description: "Six-node shell element",
// 		},
// 		FiniteElement{
// 			Shape:       Quadratic,
// 			AmountNodes: 8,
// 			Name:        "S8",
// 			Description: "Eight-node shell element",
// 		},
// 		FiniteElement{
// 			Shape:       Quadratic,
// 			AmountNodes: 8,
// 			Name:        "S8R",
// 			Description: "Eight-node shell element",
// 		},
// 	}
// }
//
// GetFiniteElementByName - get finite element by name
// func GetFiniteElementByName(name string) (fe *FiniteElement, err error) {
// 	for i := range FiniteElementDatabase {
// 		if name == FiniteElementDatabase[i].Name {
// 			return &FiniteElementDatabase[i], nil
// 		}
// 	}
// 	return nil, fmt.Errorf("Cannot found finite element by name - %v", name)
// }

//------------------------------------------
// INP file format
// *Heading
//  cone.inp
// *NODE
// 1, 0, 0, 0
// ******* E L E M E N T S *************
// *ELEMENT, type=T3D2, ELSET=Line1
// 7, 1, 7
// *ELEMENT, type=CPS3, ELSET=Surface17
// 1906, 39, 234, 247
//------------------------------------------

// /*
// *MATERIAL, NAME=stell
// *ELASTIC
// 2.9E+07,0.28
//
// *DENSITY
// 7.35E-4
// *EXPANSION
// 7.228E-6
//
// *SHELL SECTION, MATERIAL=stell,ELSET=shell,OFFSET=0
// 0.005
//
// *BOUNDARY
// Bottom,1,1,0
//
// *STEP
// *BUCKLE
// 5
// *CLOAD
// Top,2,-100
//
// *NODE FILE
// U
// *NODE PRINT
// U,NT,RF
// *NODE FILE
// U,NT,RF
// *EL FILE
// U,S
// *EL PRINT
// S
// *END STEP
// */

// ElementData - inp elements
// type ElementData struct {
// 	Index  int
// 	IPoint []int
// }
//
// // Element - inp element
// type Element struct {
// 	Name string
// 	FE   *FiniteElement
// 	Data []ElementData
// }
//
// // NamedNode - list of nodes with specific name
// type NamedNode struct {
// 	Name  string
// 	Nodes []int
// }
//
// // ShellSection - add thickness for shell elements
// type ShellSection struct {
// 	ElementName string
// 	Thickness   float64
// }

// BoundaryProperty - fixed point
// For structures:
// – 1: translation in the local x-direction
// – 2: translation in the local y-direction
// – 3: translation in the local z-direction
// – 4: rotation about the local x-axis (only for nodes belonging to beams or shells)
// – 5: rotation about the local y-axis (only for nodes belonging to beams or shells)
// – 6: rotation about the local z-axis (only for nodes belonging to beams or shells)
// type BoundaryProperty struct {
// 	NodesByName   string
// 	StartFreedom  int
// 	FinishFreedom int
// 	Value         float64
// }
//
// //StepProperty - property of load case
// type StepProperty struct {
// 	AmountBucklingShapes int
// 	Loads                []Load
// }
//
// // Load - load
// type Load struct {
// 	NodesByName string
// 	Direction   int
// 	LoadValue   float64
// }
//
// var materialProperty string
// var stepProperty string
//
// func init() {
// 	materialProperty = `
// *MATERIAL,NAME=steel
// *ELASTIC
// 2.1e11,0.3
//
// *DENSITY
// 7.35E-4
// *EXPANSION
// 7.228E-6
// `
// 	stepProperty = `
// *NODE FILE
// U
// ***NODE PRINT
// **U,NT,RF
// ***NODE FILE
// **U,NT,RF
// ***EL FILE
// **U,S
// ***EL PRINT
// **S
// `
// }
//
// AddUniqueIndexToElements - add unique index for element with Index == -1
// func (f *Format) AddUniqueIndexToElements() {
// 	var maxIndexElement int
// 	for _, element := range f.Elements {
// 		for _, data := range element.Data {
// 			if data.Index > maxIndexElement {
// 				maxIndexElement = data.Index
// 			}
// 		}
// 	}
// 	if maxIndexElement <= 0 {
// 		maxIndexElement = 1
// 	}
// 	// add unique index only if "Index == -1"
// 	for iE, element := range f.Elements {
// 		for iD, data := range element.Data {
// 			if data.Index == -1 {
// 				maxIndexElement++
// 				f.Elements[iE].Data[iD].Index = maxIndexElement
// 			}
// 		}
// 	}
// }

// AddNamedNodesOnLevel - add named nodes on specific elevation with name
// func (f *Format) AddNamedNodesOnLevel(level float64, name string) int {
// 	eps := 1e-8
// 	var n NamedNode
// 	n.Name = name
// 	for _, node := range f.Nodes {
// 		y := node.Coord[1]
// 		if math.Abs(y-level) <= eps {
// 			n.Nodes = append(n.Nodes, node.Index)
// 		}
// 	}
// 	if len(n.Nodes) > 0 {
// 		f.NodesWithName = append(f.NodesWithName, n)
// 		return len(n.Nodes)
// 	}
// 	return -1
// }

// Open - open file in inp format
// func (inp *Format) Open(file string) (err error) {
// 	inFile, err := os.Open(file)
// 	if err != nil {
// 		return
// 	}
// 	defer func() {
// 		errFile := inFile.Close()
// 		if errFile != nil {
// 			if err != nil {
// 				err = fmt.Errorf("%v ; %v", err, errFile)
// 			} else {
// 				err = errFile
// 			}
// 		}
// 	}()
// 	scanner := bufio.NewScanner(inFile)
// 	scanner.Split(bufio.ScanLines)
//
// 	type stageReading uint
//
// 	const (
// 		stageHeading stageReading = iota
// 		stageNode
// 		stageElement
// 		stageNamedNode
// 	)
//
// 	var stage stageReading
// 	var element Element
// 	var namedNode NamedNode
//
// 	for scanner.Scan() {
// 		line := scanner.Text()
// 		line = strings.TrimSpace(line)
//
// 		// empty line
// 		if len(line) == 0 {
// 			continue
// 		}
//
// 		// comments
// 		if len(line) >= 2 && line[0] == '*' && line[1] == '*' {
// 			continue
// 		}
//
// 		// change stage
// 		if line[0] == '*' {
// 			s := strings.ToUpper(line)
// 			switch {
// 			case strings.Contains(s, "HEADING"):
// 				stage = stageHeading
// 			case strings.Contains(s, "NODE"):
// 				stage = stageNode
// 			case strings.Contains(s, "ELEMENT"):
// 				saveElement(element, inp)
// 				element, err = convertElement(line)
// 				if err != nil {
// 					return err
// 				}
// 				stage = stageElement
// 			case strings.Contains(s, "NSET"):
// 				saveNamedNode(namedNode, inp)
// 				namedNode, err = convertNamedNode(line)
// 				if err != nil {
// 					return err
// 				}
// 				stage = stageNamedNode
// 			default:
// 				return fmt.Errorf("Cannot found type for that line : %v", line)
// 			}
// 			continue
// 		}
//
// 		switch stage {
// 		case stageHeading:
// 			inp.Name = line
// 			continue
// 		case stageNode:
// 			node, err := convertStringToNode(line)
// 			if err != nil {
// 				return err
// 			}
// 			inp.Nodes = append(inp.Nodes, node)
// 		case stageElement:
// 			el, err := convertStringToElement(element, line)
// 			if err != nil {
// 				return err
// 			}
// 			element.Data = append(element.Data, el)
// 		case stageNamedNode:
// 			index, err := convertStringToNameIndex(line)
// 			if err != nil {
// 				return err
// 			}
// 			namedNode.Nodes = append(namedNode.Nodes, index)
// 		}
// 	}
// 	saveElement(element, inp)
// 	saveNamedNode(namedNode, inp)
//
// 	return nil
// }

// func saveNamedNode(namedNode NamedNode, inp *Format) {
// 	if len(namedNode.Nodes) == 0 {
// 		return
// 	}
// 	inp.NodesWithName = append(inp.NodesWithName, namedNode)
// }
//
// func saveElement(element Element, inp *Format) {
// 	if len(element.Data) == 0 {
// 		return
// 	}
// 	inp.Elements = append(inp.Elements, element)
// }

// convert named nodes
// *NSET, NSET = name
// func convertNamedNode(line string) (namedNode NamedNode, err error) {
// 	s := strings.Split(line, ",")
// 	for i := range s {
// 		s[i] = strings.TrimSpace(s[i])
// 	}
// 	{
// 		r := strings.Split(s[1], "=")
// 		for i := range r {
// 			r[i] = strings.TrimSpace(r[i])
// 		}
// 		if len(r) != 2 {
// 			return namedNode, fmt.Errorf("Wrong in second NSET - %v", line)
// 		}
// 		namedNode.Name = strings.TrimSpace(r[1])
// 		if len(namedNode.Name) == 0 {
// 			return namedNode, fmt.Errorf("Name is empty and this is not acceptable - %v", line)
// 		}
// 	}
// 	return namedNode, nil
// }

// convert element
// *ELEMENT, type=CPS3, ELSET=shell
// func convertElement(line string) (el Element, err error) {
// 	s := strings.Split(line, ",")
// 	for i := range s {
// 		s[i] = strings.TrimSpace(s[i])
// 	}
// 	// found the type
// 	{
// 		r := strings.Split(s[1], "=")
// 		for i := range r {
// 			r[i] = strings.ToUpper(strings.TrimSpace(r[i]))
// 		}
// 		if len(r) != 2 {
// 			return el, fmt.Errorf("Wrong in second element - %v", line)
// 		}
// 		var found bool
// 		for i, f := range FiniteElementDatabase {
// 			if r[1] == f.Name {
// 				el.FE = &FiniteElementDatabase[i]
// 				found = true
// 				break
// 			}
// 		}
// 		if !found {
// 			return el, fmt.Errorf("Cannot convert to finite element - %v", line)
// 		}
// 	}
// 	{
// 		r := strings.Split(s[2], "=")
// 		for i := range r {
// 			r[i] = strings.TrimSpace(r[i])
// 		}
// 		if len(r) != 2 {
// 			return el, fmt.Errorf("Wrong in 3 element - %v", line)
// 		}
// 		el.Name = r[1]
// 		if len(el.Name) == 0 {
// 			return el, fmt.Errorf("Name is empty and this is not acceptable - %v", line)
// 		}
// 	}
// 	return el, nil
// }

// separate by , and trim
// func separate(line string) (s []string) {
// 	s = strings.Split(line, ",")
// 	for i := range s {
// 		s[i] = strings.TrimSpace(s[i])
// 	}
// 	return s
// }

// convert index of node in string to int
// 1,
// // 5921,
// func convertStringToNameIndex(line string) (index int, err error) {
// 	s := separate(line)
// 	i, err := strconv.ParseInt(s[0], 10, 64)
// 	if err != nil {
// 		return
// 	}
// 	return int(i), nil
// }
//
//
// // *ELEMENT, type=T3D2, ELSET=Line1
// // 7, 1, 7
// // *ELEMENT, type=CPS3, ELSET=Surface17
// // 1906, 39, 234, 247
// func convertStringToElement(el Element, line string) (c ElementData, err error) {
// 	s := separate(line)
// 	if el.FE == nil {
// 		return c, fmt.Errorf("Error in convertStringToElement: element is nil")
// 	}
// 	if len(s) != el.FE.AmountNodes+1 {
// 		return c, fmt.Errorf("Wrong string line for convert coordinate of point.\nLine = %v.\nSlice=%v", line, s)
// 	}
// 	var array []int
// 	for i := 0; i < el.FE.AmountNodes+1; i++ {
// 		result, err := strconv.ParseInt(s[0], 10, 64)
// 		if err != nil {
// 			return c, fmt.Errorf("Cannot convert to int - %v on line - %v", s[i], line)
// 		}
// 		array = append(array, int(result))
// 	}
//
// 	c.Index = array[0]
// 	c.IPoint = array[1:]
//
// 	return c, err
// }
//
// //------------------------------------------
// // INP file format
// // *Heading
// //  cone.inp
// // *NODE
// // 1, 0, 0, 0
// // ******* E L E M E N T S *************
// // *ELEMENT, type=T3D2, ELSET=Line1
// // 7, 1, 7
// // *ELEMENT, type=CPS3, ELSET=Surface17
// // 1906, 39, 234, 247
// //------------------------------------------
//
// type pp []Node
//
// // func (a pp) Len() int           { return len(a) }
// // func (a pp) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
// // func (a pp) Less(i, j int) bool { return a[i].Index < a[j].Index }
//
// // Save - convertor
// func (f Format) Save(filename string) (err error) {
// 	if len(f.Name) == 0 {
// 		f.Name = filename
// 	}
// 	panic(" 	err = utils.CreateNewFile(filename, f.SaveINPtoLines())")
// 	return err
//
// }
//
// // SaveINPtoLines - converting
// func (f Format) SaveINPtoLines() (lines []string) {
//
// 	lines = make([]string, 0, len(f.Elements)+len(f.Nodes)+10)
//
// 	lines = append(lines, "*HEADING")
// 	f.Name = strings.TrimSpace(f.Name)
// 	if len(f.Name) == 0 {
// 		f.Name = "Convertor"
// 	}
// 	lines = append(lines, f.Name)
//
// 	// sort points by index
// 	sort.Sort(pp(f.Nodes))
//
// 	lines = append(lines, "*NODE")
// 	for _, node := range f.Nodes {
// 		lines = append(lines, fmt.Sprintf("%v, %.10e, %.10e, %.10e", node.Index, node.Coord[0], node.Coord[1], node.Coord[2]))
// 	}
//
// 	lines = append(lines, "**** ELEMENTS ****")
// 	for _, element := range f.Elements {
// 		element.Name = strings.TrimSpace(element.Name)
// 		if len(element.Name) == 0 {
// 			element.Name = "Convertor"
// 		}
// 		lines = append(lines, fmt.Sprintf("*ELEMENT, type=%v, ELSET=%v", element.FE.Name, element.Name))
// 		for _, data := range element.Data {
// 			s := fmt.Sprintf("%v", data.Index)
// 			for _, point := range data.IPoint {
// 				s += fmt.Sprintf(",%v", point)
// 			}
// 			lines = append(lines, s)
// 		}
// 	}
//
// 	lines = append(lines, "**** Named nodes ****")
// 	for _, n := range f.NodesWithName {
// 		lines = append(lines, fmt.Sprintf("*NSET,NSET=%v", n.Name))
// 		for _, i := range n.Nodes {
// 			lines = append(lines, fmt.Sprintf("%v,", i))
// 		}
// 	}
//
// 	lines = append(lines, "**** Property of material ****")
// 	lines = append(lines, materialProperty)
//
// 	lines = append(lines, "**** Shell property ****")
// 	for _, s := range f.ShellSections {
// 		lines = append(lines, fmt.Sprintf("*SHELL SECTION,MATERIAL=steel,ELSET=%v", s.ElementName)) //Remove: ,,OFFSET=0
// 		lines = append(lines, fmt.Sprintf("%.10e", s.Thickness))
// 	}
//
// 	lines = append(lines, "**** Boundary property ****")
// 	for _, b := range f.Boundary {
// 		lines = append(lines, "*BOUNDARY")
// 		lines = append(lines, fmt.Sprintf("%v,%v,%v,%v", b.NodesByName, b.StartFreedom, b.FinishFreedom, b.Value))
// 	}
//
// 	lines = append(lines, "**** STEP PROPERTY ****")
// 	if f.Step.AmountBucklingShapes > 0 || len(f.Step.Loads) > 0 {
// 		lines = append(lines, "*STEP")
// 		if f.Step.AmountBucklingShapes > 0 {
// 			lines = append(lines, "*BUCKLE")
// 			lines = append(lines, fmt.Sprintf("%v", f.Step.AmountBucklingShapes))
// 		}
// 		if len(f.Step.Loads) > 0 {
// 			for _, l := range f.Step.Loads {
// 				lines = append(lines, "*CLOAD")
// 				lines = append(lines, fmt.Sprintf("%v,%v,%.10e", l.NodesByName, l.Direction, l.LoadValue))
// 			}
// 		}
// 		lines = append(lines, stepProperty)
// 		lines = append(lines, "*END STEP")
// 	}
//
// 	return lines
// }
//
// // SupportForce - struct for saving information in support force
// type SupportForce struct {
// 	Time     float64
// 	NodeName string
// 	Forces   []Force
// }
//
// // Force - force
// type Force struct {
// 	NodeIndex int
// 	Load      [3]float64
// }
//
// // SupportForces - return forces on support
// // Examples in dat file:
// // forces (fx,fy,fz) for set FIX and time  0.4000000E-01
// // forces (fx,fy,fz) for set LOAD and time  0.2000000E-01
// // 204  3.485854E+00  1.025290E+01  3.092803E+01
// func SupportForces(datLines []string) (supportForces []SupportForce, err error) {
// 	headerPrefix := "forces (fx,fy,fz) for set"
// 	headerMiddle := "and time"
//
// 	type stage int
// 	const (
// 		undefined stage = iota
// 		header
// 		load
// 	)
//
// 	present := undefined
// 	for _, line := range datLines {
// 		line = strings.TrimSpace(line)
// 		if len(line) == 0 {
// 			if present == load {
// 				present = undefined
// 			}
// 			continue
// 		}
// 		if present != header && strings.HasPrefix(line, headerPrefix) {
// 			present = header
// 			var support SupportForce
// 			line = line[len(headerPrefix):]
// 			s := strings.Split(line, headerMiddle)
// 			support.NodeName = strings.TrimSpace(s[0])
// 			time, err := strconv.ParseFloat(strings.TrimSpace(s[1]), 64)
// 			if err != nil {
// 				return supportForces, fmt.Errorf("line = %v\nerr=%v", line, err)
// 			}
// 			support.Time = time
// 			supportForces = append(supportForces, support)
// 			continue
// 		}
// 		if present == header || present == load {
// 			present = load
// 			f, err := parseForce(line)
// 			if err != nil {
// 				return supportForces, err
// 			}
// 			supportForces[len(supportForces)-1].Forces = append(supportForces[len(supportForces)-1].Forces, f)
// 		}
// 	}
// 	return supportForces, nil
// }
//
// func parseForce(line string) (force Force, err error) {
// 	s := strings.Split(line, " ")
// 	for i := range s {
// 		s[i] = strings.TrimSpace(s[i])
// 	}
//
// 	var index int
//
// 	for index = 0; index < len(s); index++ {
// 		if len(s[index]) == 0 {
// 			continue
// 		}
// 		i, err := strconv.ParseInt(s[index], 10, 64)
// 		if err != nil {
// 			return force, fmt.Errorf("Error: string parts - %v, error - %v, in line - %v", s, err, line)
// 		}
// 		force.NodeIndex = int(i)
// 		break
// 	}
//
// 	foundPositions := 0
// 	for position := 0; position < 3; position++ {
// 		for index++; index < len(s); index++ {
// 			if len(s[index]) == 0 {
// 				continue
// 			}
// 			factor, err := strconv.ParseFloat(s[index], 64)
// 			if err != nil {
// 				return force, fmt.Errorf("Error: string parts - %v, error - %v, in line - %v", s, err, line)
// 			}
// 			force.Load[position] = factor
// 			foundPositions++
// 			break
// 		}
// 	}
// 	if foundPositions != 3 {
// 		return force, fmt.Errorf("Cannot found enought values. line = %v\ns = %v\nforce = %v", line, s, force)
// 	}
//
// 	return force, nil
// }
//
// // SupportForceSummary - type of summary force
// type SupportForceSummary struct {
// 	Time     float64
// 	NodeName string
// 	Load     [3]float64
// }
//
// // SupportForcesSummary - return summary force on support
// func SupportForcesSummary(datLines []string) (summaryForce []SupportForceSummary, err error) {
// 	s, err := SupportForces(datLines)
// 	if err != nil {
// 		return nil, err
// 	}
// 	for _, force := range s {
// 		var summ SupportForceSummary
// 		summ.Time = force.Time
// 		summ.NodeName = force.NodeName
// 		for _, f := range force.Forces {
// 			for i := 0; i < 3; i++ {
// 				summ.Load[i] += f.Load[i]
// 			}
// 		}
// 		summaryForce = append(summaryForce, summ)
// 	}
// 	return summaryForce, nil
// }
