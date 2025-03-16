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
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/Konstantin8105/efmt"
	"github.com/Konstantin8105/errors"
	"github.com/Konstantin8105/pow"
)

var CcxApp = flag.String("ccx", "", "Example:\n`ccx` for Linux\n`ccx.exe` for Windows")

func DefaultCcx() {
	if *CcxApp != "" {
		return
	}
	*CcxApp = "ccx"
	if runtime.GOOS == "windows" {
		*CcxApp = "ccx.exe"
	}
}

func CcxCpu(cpu int) {
	amount := runtime.NumCPU()
	if cpu <= 0 || amount < cpu {
		cpu = amount
	}
	os.Setenv("OMP_NUM_THREADS", fmt.Sprintf("%d", cpu))
}

// Model - summary inp format
type Model struct {
	Heading           string
	Nodes             []Node
	Elements          []Element
	Nsets             []Set
	Elsets            []Set
	Surfaces          []Surface
	Materials         []Material
	InitialConditions Condition
	BeamSections      []BeamSection
	SolidSections     []SolidSection
	ShellSections     []ShellSection
	Boundaries        []Boundary
	Springs           []Spring
	Steps             []Step
	TimePoint         struct {
		Name     string
		Generate bool
		Time     []float64
	}
	RigidBodies           []RigidBody
	DistributingCouplings []DistributingCoupling
}

type Property struct {
	E           float64 // Young`s modudlus
	V           float64 // Poisson`s ratio
	Temperature float64 // Temperature
}

type Expansion struct {
	Value       float64 // Expansion
	Temperature float64 // Temperature
}

type Material struct {
	Name       string
	Density    float64
	Expansions []Expansion
	Properties []Property
	Plastic    struct {
		Hardening string
		Data      [10]struct {
			StressVonMises float64
			PlasticStrain  float64
			Temperature    float64
		}
	}
}

func (m Material) String() string {
	var buf bytes.Buffer
	if m.Name != "" {
		fmt.Fprintf(&buf, "*MATERIAL, NAME=%s\n", m.Name)
	}
	if 0 < len(m.Properties) {
		fmt.Fprintf(&buf, "*ELASTIC\n")
		for _, pr := range m.Properties {
			fmt.Fprintf(&buf, "%s, %s, %s\n",
				efmt.Sprint(pr.E),
				efmt.Sprint(pr.V),
				efmt.Sprint(pr.Temperature),
			)
		}
	}
	if m.Plastic.Hardening != "" {
		fmt.Fprintf(&buf, "*PLASTIC, HARDENING=%s\n", m.Plastic.Hardening)
		for _, d := range m.Plastic.Data {
			if d.StressVonMises != 0.0 {
				fmt.Fprintf(&buf, "%s, %s, %s\n",
					efmt.Sprint(d.StressVonMises),
					efmt.Sprint(d.PlasticStrain),
					efmt.Sprint(d.Temperature))
			}
		}
	}
	if len(m.Expansions) == 1 {
		fmt.Fprintf(&buf, "*EXPANSION\n%s\n", efmt.Sprint(m.Expansions[0].Value))
	} else if 1 < len(m.Expansions) {
		fmt.Fprintf(&buf, "*EXPANSION, TYPE=ISO, ZERO=%s\n",
			efmt.Sprint(m.Expansions[0].Temperature))
		for _, e := range m.Expansions {
			fmt.Fprintf(&buf, "%s, %s\n",
				efmt.Sprint(e.Value),
				efmt.Sprint(e.Temperature),
			)
		}
	}
	fmt.Fprintf(&buf, "*DENSITY\n%s\n", efmt.Sprint(m.Density))
	return buf.String()
}

type Surface struct {
	Name          string
	IsElementType bool
	List          [][2]string
}

func (s Surface) String() string {
	if len(s.List) == 0 {
		return ""
	}
	var buf bytes.Buffer

	fmt.Fprintf(&buf, "\n*SURFACE")
	if s.Name != "" {
		fmt.Fprintf(&buf, ", NAME=%s", s.Name)
	}
	if s.IsElementType {
		fmt.Fprintf(&buf, ", TYPE=ELEMENT\n")
		for _, l := range s.List {
			fmt.Fprintf(&buf, "%s, %s\n", l[0], l[1])
		}
	} else {
		fmt.Fprintf(&buf, ", TYPE=NODE\n")
		for i, l := range s.List {
			fmt.Fprintf(&buf, "%s", l[0])
			if i == len(s.List) {
				fmt.Fprintf(&buf, "\n")
			} else {
				fmt.Fprintf(&buf, ",\n")
			}
		}
	}
	fmt.Fprintf(&buf, "\n")
	return buf.String()
}

type Step struct {
	IsStatic bool
	Static   struct {
		TimeInc    float64
		TimePeriod float64
	}

	Nlgeom bool // genuine nonlinear geometric calculation
	Inc    int  // The maximum number of increments in the step (for automatic
	// incrementation) can be specified by using
	// the parameter INC (default is 100)

	Boundaries []Boundary

	Buckle struct {
		Number   int     // Number of buckling factors desired (usually 1)
		Accuracy float64 // Accuracy desired (default: 0.01).
	}
	NodeFiles    []Print
	ElFiles      []Print
	NodePrints   []Print
	ElPrints     []Print
	Cloads       []Cload
	Dloads       []Dload
	Temperatures []Temperature
}

func (s Step) String() string {
	var buf bytes.Buffer

	fmt.Fprintf(&buf, "\n*STEP")
	if s.Nlgeom {
		fmt.Fprintf(&buf, ", NLGEOM")
	} else {
		fmt.Fprintf(&buf, ", NLGEOM=NO")
	}
	if s.Inc != 0 {
		fmt.Fprintf(&buf, ", INC=%d", s.Inc)
	}
	fmt.Fprintf(&buf, "\n")

	if s.IsStatic {
		fmt.Fprintf(&buf, "*STATIC\n")
		if s.Static.TimeInc != 0.0 || s.Static.TimePeriod != 0.0 {
			fmt.Fprintf(&buf, "%.8e, %.8e\n",
				s.Static.TimeInc, s.Static.TimePeriod)
		}
	}

	if 0 < s.Buckle.Number {
		fmt.Fprintf(&buf, "*BUCKLE\n")
		if s.Buckle.Accuracy == 0 {
			fmt.Fprintf(&buf, "%d\n", s.Buckle.Number)
		} else {
			fmt.Fprintf(&buf, "%d,%.12e\n", s.Buckle.Number, s.Buckle.Accuracy)
		}
	}

	for _, load := range s.Cloads {
		fmt.Fprintf(&buf, "%s", load.String())
	}
	for _, load := range s.Dloads {
		fmt.Fprintf(&buf, "%s", load.String())
	}
	for _, load := range s.Temperatures {
		fmt.Fprintf(&buf, "%s", load.String())
	}
	for _, boun := range s.Boundaries {
		if boun.LoadLocation == "" {
			continue
		}
		fmt.Fprintf(&buf, "*BOUNDARY\n%s,%d,%d,%.8e\n",
			boun.LoadLocation, boun.Start, boun.Finish, boun.Factor,
		)
	}

	for _, slice := range []struct {
		prefix     string
		prefixName string
		prints     []Print
	}{
		{prefix: "*NODE FILE", prefixName: "NSET", prints: s.NodeFiles},
		{prefix: "*EL FILE", prefixName: "ELSET", prints: s.ElFiles},
		{prefix: "*NODE PRINT", prefixName: "NSET", prints: s.NodePrints},
		{prefix: "*EL PRINT", prefixName: "ELSET", prints: s.ElPrints},
	} {
		for _, pr := range slice.prints {
			fmt.Fprintf(&buf, "%s", slice.prefix)
			if pr.SetName != "" {
				fmt.Fprintf(&buf, ", %s=%s", slice.prefixName, pr.SetName)
			}
			if pr.Frequency != "" {
				fmt.Fprintf(&buf, ", FREQUENCY=%s", pr.Frequency)
			}
			if pr.Output != "" {
				fmt.Fprintf(&buf, ", OUTPUT=%s", pr.Output)
			}
			if pr.TotalOnly {
				fmt.Fprintf(&buf, ", TOTALS=ONLY")
			}
			if pr.TimePoints != "" {
				fmt.Fprintf(&buf, ", TIME POINTS=%s", pr.TimePoints)
			}
			if pr.ContactElement {
				fmt.Fprintf(&buf, ", CONTACT ELEMENT")
			}
			if pr.Global {
				fmt.Fprintf(&buf, ", GLOBAL=YES")
			}
			fmt.Fprintf(&buf, "\n%s\n", strings.Join(pr.Options, ", "))
		}
	}

	fmt.Fprintf(&buf, "\n*END STEP\n")

	return buf.String()
}

type Condition struct {
	Type            string
	NodeSet         string
	TemperatureNode float64
}

func (c Condition) String() string {
	if c.Type == "" {
		return ""
	}
	var out string
	out += "*INITIAL CONDITIONS"
	if c.Type != "" {
		out += fmt.Sprintf(", TYPE=%s", c.Type)
	}
	out += "\n"
	out += fmt.Sprintf("%s, %.7e", c.NodeSet, c.TemperatureNode)
	out += "\n"
	return out
}

func (f Model) String() string {
	var buf bytes.Buffer

	if f.Heading != "" {
		fmt.Fprintf(&buf, "*Heading\n%s\n", f.Heading)
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
			if len(el.Nodes) == 0 {
				continue
			}
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
	writeSet(&buf, "NSET", f.Nsets)
	writeSet(&buf, "ELSET", f.Elsets)

	for _, s := range f.Surfaces {
		fmt.Fprintf(&buf, "%s", s)
	}
	fmt.Fprintf(&buf, "%s", f.InitialConditions.String())
	for _, s := range f.SolidSections {
		fmt.Fprintf(&buf, "%s", s.String())
	}
	for _, s := range f.ShellSections {
		fmt.Fprintf(&buf, "%s", s.String())
	}
	for _, s := range f.BeamSections {
		fmt.Fprintf(&buf, "%s", s.String())
	}
	for _, s := range f.Springs {
		fmt.Fprintf(&buf, "%s", s.String())
	}

	if f.TimePoint.Name != "" {
		fmt.Fprintf(&buf, "*TIME POINTS, NAME=%s", f.TimePoint.Name)
		if f.TimePoint.Generate {
			fmt.Fprintf(&buf, ", GENERATE")
		}
		fmt.Fprintf(&buf, "\n")
		for pos, t := range f.TimePoint.Time {
			fmt.Fprintf(&buf, "%f ", t)
			if pos != len(f.TimePoint.Time)-1 {
				fmt.Fprintf(&buf, ",")
			}
		}
		fmt.Fprintf(&buf, "\n")
	}

	for _, boun := range f.Boundaries {
		if boun.LoadLocation == "" {
			continue
		}
		fmt.Fprintf(&buf, "*BOUNDARY\n%s,%d,%d,%.8e\n",
			boun.LoadLocation, boun.Start, boun.Finish, boun.Factor,
		)
	}

	for i := range f.RigidBodies {
		fmt.Fprintf(&buf, "%s", f.RigidBodies[i])
	}
	for _, d := range f.DistributingCouplings {
		fmt.Fprintf(&buf, "%s", d)
	}

	for i := range f.Materials {
		fmt.Fprintf(&buf, "%s", f.Materials[i].String())
	}
	for i := range f.Steps {
		fmt.Fprintf(&buf, "%s", f.Steps[i].String())
	}

	return strings.ToUpper(buf.String())
}

type Temperature struct {
	Parameters      []string
	NodeSet         string
	TemperatureNode float64
	Gradient2       float64
	Gradient1       float64
}

func (t Temperature) String() string {
	if t.NodeSet == "" {
		return ""
	}
	var out string
	out += "*TEMPERATURE"
	if 0 < len(t.Parameters) {
		out += "," + strings.Join(t.Parameters, ",")
	}
	out += "\n"
	out += fmt.Sprintf("%s, %.7e", t.NodeSet, t.TemperatureNode)
	if t.Gradient2 != 0 || t.Gradient1 != 0 {
		out += fmt.Sprintf(" , %.7e", t.Gradient2)
	}
	if t.Gradient1 != 0 {
		out += fmt.Sprintf(" , %.7e", t.Gradient1)
	}
	return out + "\n"
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

func (f *Model) parseHeading(block []string) (ok bool, err error) {
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
//
//	*NODE
//	1, 0, 0, 0
//
//	*NODE, NSET=Nall
//	1, 0, 0, 0
//
// First line:
//
//	*NODE
//	Enter the optional parameter, if desired.
//
// Following line:
//
//	node number.
//	Value of first coordinate.
//	Value of second coordinate.
//	Value of third coordinate.
func (f *Model) parseNode(block []string) (ok bool, err error) {
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
			coord[i], err = parseFloat(fields[i+1])
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
//
//	*ELEMENT
//	Enter any needed parameters and their values.
//
// Following line:
//
//	Element number.
//	Node numbers forming the element. The order of nodes around the element is
//	given in section 2.1. Use continuation lines for elements having more
//	than 15 nodes (maximum 16 entries per line).
func (f *Model) parseElement(block []string) (ok bool, err error) {
	if !isHeader(block[0], "*ELEMENT") {
		return false, nil
	}

	var Type, Elset string
	{
		fs := fields(block[0])
		for _, f := range fs {
			if strings.HasPrefix(f, "TYPE=") {
				Type = f[5:]
			}
			if strings.HasPrefix(f, "ELSET=") {
				Elset = f[6:]
			}
		}
	}

	for _, line := range block[1:] {
		fs := fields(line)
		var ints []int
		for _, f := range fs {
			if f == "" {
				continue
			}
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
	Addition []string
	Indexes  []int
	Names    []string
}

func (s Set) String(name string) string {
	if len(s.Indexes) == 0 && len(s.Names) == 0 {
		return "\n"
	}
	var buf bytes.Buffer
	// first line
	fmt.Fprintf(&buf, "*%s", name)
	if s.Name != "" {
		switch name {
		case "ELSET":
			fmt.Fprintf(&buf, ", ELSET=%s", s.Name)
		case "NSET":
			fmt.Fprintf(&buf, ", NSET=%s", s.Name)
		default:
			panic(fmt.Errorf("not implemented: %s", name))
		}
	}
	if s.Generate {
		fmt.Fprintf(&buf, ", GENERATE")
	}
	fmt.Fprintf(&buf, "%s\n", strings.Join(s.Addition, ","))

	// list
	var list []string
	for _, ind := range s.Indexes {
		list = append(list, fmt.Sprintf(" %5d ", ind))
	}
	for _, ind := range s.Names {
		list = append(list, fmt.Sprintf(" %s  ", ind))
	}

	// combine
	for i := range list {
		fmt.Fprintf(&buf, "%s ", list[i])
		if i != len(list)-1 {
			fmt.Fprintf(&buf, " , ")
		}
		if (i+1)%9 == 0 && i != len(list)-1 {
			fmt.Fprintf(&buf, "\n")
		}
	}
	return buf.String()
}

func writeSet(out io.Writer, name string, sets []Set) {
	if len(sets) == 0 {
		return
	}
	for _, s := range sets {
		if len(s.Indexes) == 0 && len(s.Names) == 0 {
			continue
		}
		fmt.Fprintf(out, "%s\n", s.String(name))
	}
}

func (f *Model) parseSet(s *[]Set, prefix string, block []string) (ok bool, err error) {
	if !isHeader(block[0], "*"+prefix) {
		return false, nil
	}
	defer func() {
		if err != nil {
			err = fmt.Errorf("parseSet: %v", err)
		}
	}()

	var set Set
	fs := fields(block[0])[1:]
	for _, s := range fs {
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
				err = nil
				set.Names = append(set.Names, f)
			} else {
				set.Indexes = append(set.Indexes, int(i64))
			}
		}
	}
	(*s) = append((*s), set)

	return true, nil
}

func (f *Model) parseDensity(block []string) (ok bool, err error) {
	if !isHeader(block[0], "*DENSITY") {
		return false, nil
	}
	var ro float64
	ro, err = parseFloat(block[1])
	if err != nil {
		return
	}
	if len(f.Materials) == 0 {
		f.Materials = make([]Material, 1)
	}
	f.Materials[0].Density = ro
	return true, nil
}

func (f *Model) parseExpansion(block []string) (ok bool, err error) {
	if !isHeader(block[0], "*EXPANSION") {
		return false, nil
	}
	block = block[1:] // TODO for ZERO, TYPE
	if len(f.Materials) == 0 {
		f.Materials = make([]Material, 1)
	}
	for i := range block {
		fs := strings.Fields(block[i])
		var e Expansion
		switch len(fs) {
		case 2:
			var t float64
			t, err = parseFloat(block[1])
			if err != nil {
				return
			}
			e.Temperature = t
		case 1:
			var v float64
			v, err = parseFloat(block[0])
			if err != nil {
				return
			}
			e.Value = v
		default:
			err = fmt.Errorf("Expansion: %v", fs)
			return
		}
		f.Materials[0].Expansions = append(f.Materials[0].Expansions, e)
	}
	return true, nil
}

func (f *Model) parseMaterial(block []string) (ok bool, err error) {
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
				if len(f.Materials) == 0 {
					f.Materials = make([]Material, 1)
				}
				f.Materials[0].Name = field[5:]
			default:
				panic(fmt.Errorf("`%s` : `%s`", line, field))
			}
		}
	}
	return true, nil
}

func (f *Model) parseElastic(block []string) (ok bool, err error) {
	if !isHeader(block[0], "*ELASTIC") {
		return false, nil
	}
	line := strings.Replace(block[1], ",", " ", -1)
	fields := strings.Fields(line)
	if len(f.Materials) == 0 {
		f.Materials = make([]Material, 1)
	}
	for pos := 1; pos < len(block); pos++ {
		var pr Property
		switch len(fields) {
		case 3:
			pr.Temperature, err = parseFloat(fields[2])
			if err != nil {
				return
			}
			fallthrough
		case 2:
			pr.E, err = parseFloat(fields[0])
			if err != nil {
				return
			}
			pr.V, err = parseFloat(fields[1])
			if err != nil {
				return
			}
		}
		f.Materials[0].Properties = append(f.Materials[0].Properties, pr)
	}
	return true, nil
}

// Spring
//
// First line:
// *SPRING
// Enter the parameter ELSET and its value and any optional parameter, if needed.
//
// Second line for SPRINGA type elements: enter a blank line
// Second line for SPRING1 or SPRING2 type elements:
// • first degree of freedom (integer, for SPRING1 and SPRING2 elements)
// • second degree of freedom (integer, only for SPRING2 elements)
// Following line if the parameter NONLINEAR is not used:
// • Spring constant (real number).
type Spring struct {
	ElsetName      string
	Freedom        [2]int
	SpringConstant float64
}

func (s Spring) String() string {
	var out string
	out += fmt.Sprintf("*SPRING,ELSET=%s\n", s.ElsetName)
	if 0 < s.Freedom[0] && 0 < s.Freedom[1] {
		out += fmt.Sprintf("%d, %d\n", s.Freedom[0], s.Freedom[1])
	} else if 0 < s.Freedom[0] {
		out += fmt.Sprintf("%d\n", s.Freedom[0])
	} else {
		out += "\n"
	}
	out += fmt.Sprintf("%.7e\n", s.SpringConstant)
	return out
}

// First and only line:
//
//	*RIGID BODY
//	Enter any needed parameters and their values
type RigidBody struct {
	Nset    string
	RefNode int
	RotNode int
}

func (r RigidBody) String() string {
	var out string
	out += fmt.Sprintf("*RIGID BODY, NSET=%s", r.Nset)
	if 0 < r.RefNode {
		out += fmt.Sprintf(",REF NODE=%d", r.RefNode)
	}
	if 0 < r.RotNode {
		out += fmt.Sprintf(",ROT NODE=%d", r.RotNode)
	}
	out += "\n"
	return out
}

// First line:
//
//	*DISTRIBUTING COUPLING
//	Enter the ELSET parameter and its value
//
// Following line:
//
//	Node number or node set
//	Weight
//
// Repeat this line if needed.
type DistributingCoupling struct {
	ElsetName     string
	ElsetNode     int
	NodeIndexes   []int // with weight 1.0
	NodeNames     []string
	ReferenceNode int
}

func (d DistributingCoupling) String() string {
	if d.ElsetName == "" {
		return "\n"
	}
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "*DISTRIBUTING COUPLING,ELSET=%s\n", d.ElsetName)
	for _, n := range d.NodeIndexes {
		fmt.Fprintf(&buf, "%d,1.\n", n)
	}
	for _, n := range d.NodeNames {
		fmt.Fprintf(&buf, "%s,1.\n", n)
	}
	fmt.Fprintf(&buf, "*ELSET,ELSET=%s\n", d.ElsetName)
	fmt.Fprintf(&buf, "%d\n", d.ElsetNode)
	fmt.Fprintf(&buf, "*ELEMENT,TYPE=DCOUP3D\n")
	fmt.Fprintf(&buf, "%d, %d\n", d.ElsetNode, d.ReferenceNode)
	return buf.String()
}

// Boundary for structures:
// – 1: translation in the local x-direction
// – 2: translation in the local y-direction
// – 3: translation in the local z-direction
// – 4: rotation about the local x-axis (only for nodes belonging to beams or shells)
// – 5: rotation about the local y-axis (only for nodes belonging to beams or shells)
// – 6: rotation about the local z-axis (only for nodes belonging to beams or shells)
// – 11: temperature
//
// First line:
//
//	*BOUNDARY
//	Enter any needed parameters and their value.
//
// Following line:
//
//	Node number or node set label
//	First degree of freedom constrained
//	Last degree of freedom constrained. This field may be left blank if only one degree of freedom is constrained.
type Boundary struct {
	LoadLocation string
	Start        int
	Finish       int
	Factor       float64
}

func parseBoundary(bs *[]Boundary) func(block []string) (ok bool, err error) {

	return func(block []string) (ok bool, err error) {
		if !isHeader(block[0], "*BOUNDARY") {
			return false, nil
		}
		var b Boundary
		for _, line := range block[1:] {
			line = strings.Replace(line, ",", " ", -1)
			fields := strings.Fields(line)
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
				b.Factor, err = parseFloat(fields[3])
				if err != nil {
					return
				}
			}

			*bs = append(*bs, b)
		}

		return true, nil
	}
}

// *SOLID SECTION,ELSET=EALL,MATERIAL=HY
type SolidSection struct {
	Elset    string
	Material string
}

func (ss SolidSection) String() string {
	if ss.Elset == "" {
		return ""
	}
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "*SOLID SECTION")
	fmt.Fprintf(&buf, ", ELSET=%s", ss.Elset)
	fmt.Fprintf(&buf, ", MATERIAL=%s", ss.Material)
	fmt.Fprintf(&buf, "\n")
	return buf.String()
}

func (f *Model) parseSolidSection(block []string) (ok bool, err error) {
	if !isHeader(block[0], "*SOLID SECTION") {
		return false, nil
	}
	var ss SolidSection
	split := fields(block[0])[1:]
	for _, s := range split {
		s = strings.TrimSpace(s)
		switch {
		case strings.HasPrefix(s, "MATERIAL"):
			index := strings.Index(s, "=")
			s = strings.TrimSpace(s[index+1:])
			ss.Material = s
		case strings.HasPrefix(s, "ELSET"):
			index := strings.Index(s, "=")
			s = strings.TrimSpace(s[index+1:])
			ss.Elset = s
		case s == "":
			// do nothing
		default:
			panic(fmt.Errorf("%s", strings.Join(split, "|")))
		}
	}
	block = block[1:]
	if 0 < len(block) {
		err = fmt.Errorf("other lines: %s", strings.Join(block, "\n"))
		return
	}

	f.SolidSections = append(f.SolidSections, ss)

	return true, nil
}

type BeamSection struct {
	Section  string
	Elset    string
	Material string

	Offset1, Offset2 float64

	Thks   [2]float64
	Vector [3]float64
}

func (b BeamSection) String() string {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "*BEAM SECTION")
	fmt.Fprintf(&buf, ", SECTION=%s", b.Section)
	fmt.Fprintf(&buf, ", ELSET=%s", b.Elset)
	fmt.Fprintf(&buf, ", MATERIAL=%s", b.Material)
	if 1e-5 < math.Abs(b.Offset1) {
		fmt.Fprintf(&buf, ", OFFSET1=%.12e", b.Offset1)
	}
	if 1e-5 < math.Abs(b.Offset2) {
		fmt.Fprintf(&buf, ", OFFSET2=%.12e", b.Offset2)
	}
	fmt.Fprintf(&buf, "\n")
	for iv, v := range b.Thks {
		fmt.Fprintf(&buf, "%.7e", v)
		if iv != len(b.Thks)-1 {
			fmt.Fprintf(&buf, ",")
		}
	}
	fmt.Fprintf(&buf, "\n")
	for iv, v := range b.Vector {
		fmt.Fprintf(&buf, "%.7e", v)
		if iv != len(b.Vector)-1 {
			fmt.Fprintf(&buf, ",")
		}
	}
	fmt.Fprintf(&buf, "\n")
	return buf.String()
}

// [*BEAM SECTION, SECTION=RECT, ELSET=LINKS, MATERIAL=STEEL 10.0, 10.0 0.0, 1.0, 0.0]
// [*BEAM SECTION, SECTION=RECT, ELSET=RECHTS, MATERIAL=STEEL 5.0, 5.0 0.0, 1.0, 0.0]
// [*BEAM SECTION,ELSET=SET1,MATERIAL=EL,SECTION=RECT 0.05, 0.08 0.D0,1.D0,0.D0]
// [*BEAM SECTION,ELSET=SET2,MATERIAL=EL,SECTION=CIRC,OFFSET1=0.5,OFFSET2=.5 0.05, 0.08 0.D0,0.7071D0,0.7071D0]
// [*BEAM SECTION,ELSET=EBEAM,MATERIAL=EL,SECTION=RECT 0.05,0.10 0.,0.,1.]
// [*BEAM SECTION,ELSET=EBEAM,MATERIAL=EL,SECTION=RECT 0.05,0.10 0.,0.,1.]
func (f *Model) parseBeamSection(block []string) (ok bool, err error) {
	if !isHeader(block[0], "*BEAM SECTION") {
		return false, nil
	}
	if len(block) != 3 {
		return false, fmt.Errorf("not valid *BEAM SECTION")
	}
	var b BeamSection
	split := fields(block[0])[1:]
	for _, s := range split {
		s = strings.TrimSpace(s)
		switch {
		case strings.HasPrefix(s, "MATERIAL"):
			index := strings.Index(s, "=")
			s = strings.TrimSpace(s[index+1:])
			b.Material = s
		case strings.HasPrefix(s, "SECTION"):
			index := strings.Index(s, "=")
			s = strings.TrimSpace(s[index+1:])
			b.Section = s
		case strings.HasPrefix(s, "ELSET"):
			index := strings.Index(s, "=")
			s = strings.TrimSpace(s[index+1:])
			b.Elset = s
		case s == "":
			// do nothing
		default:
			panic(fmt.Errorf("%s", strings.Join(split, "|")))
		}
	}
	for i, s := range fields(block[1]) {
		var v float64
		v, err = parseFloat(s)
		if err != nil {
			return
		}
		b.Thks[i] = v
	}
	for i, s := range fields(block[2]) {
		var v float64
		v, err = parseFloat(s)
		if err != nil {
			return
		}
		b.Vector[i] = v
	}

	f.BeamSections = append(f.BeamSections, b)
	return true, nil
}

type ShellSection struct {
	Elset          string
	Offset         float64
	Composite      bool
	NodalThickness bool
	Property       [12]struct {
		Thickness float64
		Material  string
	}
}

func (ss ShellSection) String() string {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "*SHELL SECTION")
	fmt.Fprintf(&buf, ", ELSET=%s", ss.Elset)
	fmt.Fprintf(&buf, ", OFFSET=%f", ss.Offset)
	if ss.NodalThickness {
		fmt.Fprintf(&buf, ", NODAL THICKNESS")
	}
	if ss.Composite {
		fmt.Fprintf(&buf, ", COMPOSITE")
		fmt.Fprintf(&buf, "\n")
		for _, row := range ss.Property {
			if math.Abs(row.Thickness) < 1e-8 {
				continue
			}
			fmt.Fprintf(&buf, "%.8e,, %s\n", row.Thickness, row.Material)
		}
	} else {
		fmt.Fprintf(&buf, ", MATERIAL=%s", ss.Property[0].Material)
		fmt.Fprintf(&buf, "\n")
		fmt.Fprintf(&buf, "%.8f\n", ss.Property[0].Thickness)
	}
	return buf.String()
}

// *SHELL SECTION,MATERIAL=steel,ELSET=Eall,,OFFSET=0
// 6.2500E-02
func (f *Model) parseShellSection(block []string) (ok bool, err error) {
	if !isHeader(block[0], "*SHELL SECTION") {
		return false, nil
	}
	var ss ShellSection
	split := fields(block[0])[1:]
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
			ss.Elset = s
		case strings.HasPrefix(s, "OFFSET"):
			index := strings.Index(s, "=")
			s = strings.TrimSpace(s[index+1:])
			ss.Offset, err = parseFloat(s)
			if err != nil {
				return
			}
		case strings.HasPrefix(s, "NODAL THICKNESS"):
			ss.NodalThickness = true
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
			ss.Property[pos].Thickness, err = parseFloat(fields[0])
			if err != nil {
				err = fmt.Errorf("%v : %v", block, err)
				return
			}
			ss.Property[pos].Material = fields[1]
		}
	} else {
		line := strings.TrimSpace(block[1])
		ss.Property[0].Thickness, err = parseFloat(line)
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
//
// *STEP,NLGEOM
// *STATIC,DIRECT
// 1.,1.
// *CLOAD
// N1,3,80.33333
// N2,3,40.16666
// N3,3,20.08333
// N4,3,-80.33333
// N5,3,-160.66666
// *NODE PRINT,NSET=NALL
// U
// *EL PRINT,ELSET=EALL
// S
// *END STEP
func (f *Model) parseStep(block []string) (ok bool, err error) {
	var s Step
	if !isHeader(block[0], "*STEP") {
		return false, nil
	}
	if !isHeader(block[len(block)-1], "*END STEP") {
		return false, nil
	}
	defer func() {
		f.Steps = append(f.Steps, s)
	}()
	{ // parse first line
		fs := fields(block[0])[1:]
		for _, part := range fs {
			switch {
			case strings.Contains(part, "NLGEOM"):
				part = strings.ReplaceAll(part, "NLGEOM", "")
				part = strings.ReplaceAll(part, "=", "")
				part = strings.TrimSpace(part)
				switch part {
				case "":
					s.Nlgeom = true
				case "NO":
					s.Nlgeom = false
				default:
					err = fmt.Errorf("not valid NLGEOM: %v", part)
					return
				}

			case strings.HasPrefix(part, "INC="):
				part = part[4:]
				var i64 int64
				i64, err = strconv.ParseInt(part, 10, 64)
				if err != nil {
					return
				}
				s.Inc = int(i64)
			default:
				panic(part)
			}
		}
	}
	// remove corner lines
	block = block[1 : len(block)-1]
	blocks := splitByBlocks(block)

	et := errors.New("parse step")
	for _, block := range blocks {
		err := blockParser(block, []func(block []string) (ok bool, err error){
			s.parseBuckle,
			s.parseStatic,
			func(block []string) (ok bool, err error) {
				return s.parsePrint(block, "*NODE FILE", &(s.NodeFiles))
			},
			func(block []string) (ok bool, err error) {
				return s.parsePrint(block, "*EL FILE", &(s.ElFiles))
			},
			func(block []string) (ok bool, err error) {
				return s.parsePrint(block, "*NODE PRINT", &(s.NodePrints))
			},
			func(block []string) (ok bool, err error) {
				return s.parsePrint(block, "*EL PRINT", &(s.ElPrints))
			},
			s.parseCload,
			s.parseDload,
			parseBoundary(&s.Boundaries),
		})
		if err != nil {
			_ = et.Add(err)
		}
	}
	if et.IsError() {
		err = et
	}
	if err == nil {
		ok = true
	}
	return
}

// First line:
//
//	*BUCKLE
//
// Second line:
//
//	Number of buckling factors desired (usually 1).
//	Accuracy desired (default: 0.01).
//	# Lanczos vectors calculated in each iteration (default: 4 * #eigenvalues).
//	Maximum # of iterations (default: 1000).
func (s *Step) parseBuckle(block []string) (ok bool, err error) {
	if !isHeader(block[0], "*BUCKLE") {
		return false, nil
	}
	if len(block) > 2 {
		err = fmt.Errorf("multiline block: %s", strings.Join(block, "\n"))
		return
	}
	fs := fields(block[1])
	var i64 int64
	i64, err = strconv.ParseInt(fs[0], 10, 64)
	if err != nil {
		return
	}
	s.Buckle.Number = int(i64)
	if 1 < len(fs) {
		var acc float64
		acc, err = parseFloat(fs[1])
		if err != nil {
			return
		}
		s.Buckle.Accuracy = acc
	}
	return true, nil
}

type File struct {
	Options []string
}

type Print struct {
	SetName        string
	Frequency      string
	Output         string
	TimePoints     string
	TotalOnly      bool
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
//
//	*NODE FILE,TIME POINTS=T1
//	RF,NT
//
// requests the storage of reaction forces and temperatures in the .frd file for
// all time points defined by the T1 time points sequence
func (s *Step) parsePrint(block []string, prefix string, pr *[]Print) (ok bool, err error) {
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
			// var i64 int64
			// i64, err = strconv.ParseInt(s, 10, 64)
			// if err != nil {
			// 	return
			// }
			np.Frequency = s // int(i64)
		case strings.HasPrefix(s, "OUTPUT"):
			index := strings.Index(s, "=")
			s = strings.TrimSpace(s[index+1:])
			np.Output = s
		default:
			err = fmt.Errorf("parsePrint cannot parse: `%s`", s)
			return
		}
	}
	if len(block) == 2 {
		np.Options = strings.Fields(strings.Replace(block[1], ",", " ", -1))
	}
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
//   - Initial time increment. This value will be modified due to automatic in-
//     crementation, unless the parameter DIRECT was specified (default 1.).
//   - Time period of the step (default 1.).
//   - Minimum time increment allowed. Only active if DIRECT is not specified.
//
// Default is the initial time increment or 1.e-5 times the time period of the
// step, whichever is smaller.
//   - Maximum time increment allowed. Only active if DIRECT is not specified.
//     Default is 1.e+30
//   - Initial time increment for CFD applications (default 1.e-2)
func (s *Step) parseStatic(block []string) (ok bool, err error) {
	if !isHeader(block[0], "*STATIC") {
		return false, nil
	}
	s.IsStatic = true
	if len(block) == 1 {
		return true, nil
	}
	if len(block) != 2 {
		err = fmt.Errorf("not valid: %s", strings.Join(block, "\n"))
		return
	}
	fields := strings.Fields(strings.Replace(block[1], ",", " ", -1))
	if len(fields) != 2 {
		panic(block)
	}

	s.Static.TimeInc, err = parseFloat(fields[0])
	if err != nil {
		err = fmt.Errorf("%v : %v", block, err)
		return
	}

	s.Static.TimePeriod, err = parseFloat(fields[1])
	if err != nil {
		err = fmt.Errorf("%v : %v", block, err)
		return
	}

	return true, nil
}

type Cload struct {
	Position  string
	Direction int
	Value     float64
}

func (load Cload) String() string {
	return fmt.Sprintf("*CLOAD\n%s, %3d, %.8e\n",
		load.Position, load.Direction, load.Value)
}

// [*CLOAD 5, 1, 5000.0]
// [*CLOAD 2,3,0.0025]
// [*CLOAD LOAD,3,-3.3112583E+00]
func (s *Step) parseCload(block []string) (ok bool, err error) {
	if !isHeader(block[0], "*CLOAD") {
		return false, nil
	}
	for _, line := range block[1:] {
		line = strings.Replace(line, ",", " ", -1)
		fields := strings.Fields(line)
		if len(fields) != 3 {
			panic(line)
		}
		var l Cload
		l.Position = fields[0]

		var i64 int64
		i64, err = strconv.ParseInt(fields[1], 10, 64)
		if err != nil {
			return
		}
		l.Direction = int(i64)

		l.Value, err = parseFloat(fields[2])
		if err != nil {
			return
		}

		s.Cloads = append(s.Cloads, l)
	}

	return true, nil
}

type Dload struct {
	Values []string
}

func (load Dload) String() string {
	return fmt.Sprintf("*DLOAD\n%s\n",
		strings.Join(load.Values, " ,"))
}

// [*DLOAD EALL,GRAV,9.81,0.,0.,-1.]
// [*DLOAD 3,P,0.01]
// [*DLOAD 3,P,0.01]
func (s *Step) parseDload(block []string) (ok bool, err error) {
	if !isHeader(block[0], "*DLOAD") {
		return false, nil
	}
	if len(block) != 2 {
		return false, fmt.Errorf("not valid Dload")
	}
	var load Dload
	load.Values = fields(block[1])
	s.Dloads = append(s.Dloads, load)
	return true, nil
}

func (f *Model) parseTimePoint(block []string) (ok bool, err error) {
	if !isHeader(block[0], "*TIME POINTS") {
		return false, nil
	}

	for _, s := range strings.Split(block[0], ",")[1:] {
		s = strings.TrimSpace(s)
		switch {
		case strings.HasPrefix(s, "NAME="):
			s = s[5:]
			f.TimePoint.Name = s
		case strings.HasPrefix(s, "GENERATE"):
			f.TimePoint.Generate = true
		default:
			panic(s)
		}
	}

	line := strings.Replace(block[1], ",", " ", -1)
	fields := strings.Fields(line)
	var t float64

	t, err = parseFloat(fields[0])
	if err != nil {
		return
	}
	f.TimePoint.Time = append(f.TimePoint.Time, t)

	t, err = parseFloat(fields[1])
	if err != nil {
		return
	}
	f.TimePoint.Time = append(f.TimePoint.Time, t)

	t, err = parseFloat(fields[2])
	if err != nil {
		return
	}
	f.TimePoint.Time = append(f.TimePoint.Time, t)

	return true, nil
}

func (f *Model) parsePlastic(block []string) (ok bool, err error) {
	if !isHeader(block[0], "*PLASTIC") {
		return false, nil
	}

	if len(f.Materials) == 0 {
		f.Materials = make([]Material, 1)
	}

	for _, s := range strings.Split(block[0], ",")[1:] {
		s = strings.TrimSpace(s)
		prefixH := "HARDENING="
		switch {
		case strings.HasPrefix(s, prefixH):
			s = s[len(prefixH):]
			f.Materials[0].Plastic.Hardening = s
		default:
			panic(s)
		}
	}

	for pos, line := range block[1:] {
		line = strings.Replace(line, ",", " ", -1)
		fields := strings.Fields(line)

		f.Materials[0].Plastic.Data[pos].StressVonMises, err = parseFloat(fields[0])
		if err != nil {
			return
		}
		f.Materials[0].Plastic.Data[pos].PlasticStrain, err = parseFloat(fields[1])
		if err != nil {
			return
		}
		if len(fields) == 2 {
			continue
		}
		f.Materials[0].Plastic.Data[pos].Temperature, err = parseFloat(fields[2])
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

func blockParser(block []string, parsers []func(block []string) (ok bool, err error)) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%s\n%s",
				strings.Join(block, "\n"),
				string(debug.Stack()))
		}
	}()
	if len(block) == 0 {
		return
	}
	et := errors.New("parse string block")
	var found bool
	for pos := range parsers {
		var ok bool
		ok, err = parsers[pos](block)
		if err != nil {
			_ = et.Add(fmt.Errorf("№ %d: %v", pos, err))
			continue
		}
		found = found || ok
		if ok {
			block = nil
			break
		}
	}
	if !found {
		if len(block) > 3 {
			block = block[:3]
		}
		err = fmt.Errorf("Not found block : %v", strings.Join(block, "\n"))
		_ = et.Add(err)
	}
	if et.IsError() {
		err = et
	}
	return
}

func splitByBlocks(lines []string) (blocks [][]string) {
	for _, s := range lines {
		s = strings.TrimSpace(s)
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
	return
}

func Parse(content []byte) (f *Model, err error) {
	// split into lines
	var lines []string
	{
		dat := string(content)
		dat = strings.ReplaceAll(dat, "\r", "")
		dat = strings.ToUpper(dat)
		dat = strings.ReplaceAll(dat, "  ", " ")
		lines = strings.Split(dat, "\n")
	}
	// split into block
	blocks := splitByBlocks(lines)
	pair := [2]string{"*STEP", "*END STEP"}
	for i := range blocks {
		if len(blocks[i]) == 0 {
			continue
		}
		if !strings.Contains(blocks[i][0], pair[0]) {
			continue
		}
		for k := i + 1; k < len(blocks); k++ {
			blocks[i] = append(blocks[i], blocks[k]...)
			if strings.Contains(blocks[k][0], pair[1]) {
				blocks[k] = nil
				break
			}
			blocks[k] = nil
		}
	}

	// parsing
	f = new(Model)

	et := errors.New("Parse")
	for _, block := range blocks {
		err := blockParser(block, []func(block []string) (ok bool, err error){
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
			parseBoundary(&f.Boundaries),
			f.parseMaterial,
			ignore("*SURFACE"),
			f.parseBeamSection,
			f.parseSolidSection,
			f.parseShellSection,
			f.parseStep,
			f.parsePlastic,
			f.parseTimePoint,
			// ignore("*END STEP"),
			// ignore("*HEAT TRANSFER"),
			// ignore("*CONDUCTIVITY"),
			// ignore("*FLUID"),
			// ignore("*SPECIFIC GAS CONSTANT"),
			// ignore("*SPECIFIC HEAT"),
			// ignore("*PHYSICAL CONSTANTS"),
		})
		if err != nil {
			_ = et.Add(err)
		}
	}
	if et.IsError() {
		err = et
	}
	return
}

// type Buckle struct {
// 	Factor        float64
// 	Displacements []Node
// }
//
// type Frd struct {
// 	Nodes   []Node
// 	Buckles []Buckle
// }

//
//     2C                          5418                                     1
//  -1         1 6.00000E+00 0.00000E+00 0.00000E+00
//  -1         2 5.99980E+00 4.83855E-02 0.00000E+00
//  -1         3 3.00000E+00 5.19615E+00 0.00000E+00
//  -1         4-2.95800E+00 5.22018E+00 0.00000E+00
//  -1         5-3.00000E+00 5.19615E+00 0.00000E+00
//
//     1PSTEP                         1           1           1
//   100CL  101 0.00000E+00        5418                     4    1           1
//  -4  DISP        4    1
//
//     1PSTEP                         2           1           1
//   100CL  102 536.3893407        5418                     4    2           1
//  -4  DISP        4    1
//  -5  D1          1    2    1    0
//  -5  D2          1    2    2    0
//  -5  D3          1    2    3    0
//  -5  ALL         1    2    0    0    1ALL
//  -1         1 0.00000E+00 0.00000E+00 0.00000E+00
//  -1         2 0.00000E+00 0.00000E+00 0.00000E+00
//  -1      7586-1.51592E-05 9.16009E-06 3.94705E-08
//  -1      7588-3.28491E-05 1.94252E-05-1.75749E-08
//
// func ParseFrd(content []byte) (frd *Frd, err error) {
// 	frd = new(Frd)
//
// 	lines := strings.Split(string(content), "\n")
// 	for i := range lines {
// 		line := strings.TrimSpace(lines[i])
// 		if !strings.Contains(line, "1PSTEP") {
// 			continue
// 		}
// 		i++
// 		line = strings.TrimSpace(lines[i])
// 		fields := strings.Fields(line)
//
// 		var factor float64
// 		factor, err = parseFloat(fields[2])
// 		if err != nil {
// 			return
// 		}
// 		if factor == 0 {
// 			continue
// 		}
//
// 		frd.Buckles = append(frd.Buckles, Buckle{Factor: factor})
// 	}
// 	// sort buckle
//
// 	return
// }

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
// func (f *Model) ChangeTypeFiniteElement(from *FiniteElement, to *FiniteElement) (err error) {
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

// func (f *Model) foundPointCIndexInLineGroup(p1, p2 int, group *[]lineGroup) (middlePoint int, err error) {
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

// func (f *Model) createMiddlePoint(fe *FiniteElement) (group []lineGroup, err error) {
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

// func (f *Model) foundByIndex(index int) (node [3]float64, err error) {
// 	i := sort.Search(len(f.Nodes), func(a int) bool { return f.Nodes[a].Index >= index })
// 	if i < len(f.Nodes) && f.Nodes[i].Index == index {
// 		// index is present at nodes
// 		return f.Nodes[i].Coord, nil
// 	}
// 	// index is not present in nodes,
// 	// but i is the index where it would be inserted.
// 	return node, fmt.Errorf("Cannot found in sort.Search : %v, but i = %v", index, i)
// }

// func (f *Model) changeFEfromQuadraticToTriangle(from *FiniteElement, to *FiniteElement) {
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
// func (f *Model) AddUniqueIndexToElements() {
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
// func (f *Model) AddNamedNodesOnLevel(level float64, name string) int {
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
// func (inp *Model) Open(file string) (err error) {
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

// func saveNamedNode(namedNode NamedNode, inp *Model) {
// 	if len(namedNode.Nodes) == 0 {
// 		return
// 	}
// 	inp.NodesWithName = append(inp.NodesWithName, namedNode)
// }
//
// func saveElement(element Element, inp *Model) {
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
// func (f Model) Save(filename string) (err error) {
// 	if len(f.Name) == 0 {
// 		f.Name = filename
// 	}
// 	panic(" 	err = utils.CreateNewFile(filename, f.SaveINPtoLines())")
// 	return err
//
// }
//
// // SaveINPtoLines - converting
// func (f Model) SaveINPtoLines() (lines []string) {
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
// 			time, err := parseFloat(strings.TrimSpace(s[1]), 64)
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
// 			factor, err := parseFloat(s[index], 64)
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

func fields(str string) (fs []string) {
	fs = strings.Split(str, ",")
	for i := range fs {
		fs[i] = strings.TrimSpace(fs[i])
	}
	return
}

func parseInt(str string) (v int, err error) {
	str = strings.TrimSpace(str)
	str = strings.ReplaceAll(str, "D", "e")
	v64, err := strconv.ParseInt(str, 10, 64)
	return int(v64), err
}

func parseFloat(str string) (v float64, err error) {
	str = strings.TrimSpace(str)
	str = strings.ReplaceAll(str, "D", "e")
	v, err = strconv.ParseFloat(str, 64)
	return
}

type Dat struct {
	BucklingFactors    []float64
	Temperatures       []Single
	Displacements      []Record
	EigenDisplacements [][]Record
	Stresses           []Stress
	Forces             []Record
	TotalForces        []Record
	EqPlasticStrain    []Pe
}

func (d Dat) MaxTime() (mt float64) {
	list := [][]Record{
		d.Displacements,
		d.Forces,
		d.TotalForces,
	}
	list = append(list, d.EigenDisplacements...)
	for i := range list {
		for j := range list[i] {
			mt = math.Max(mt, list[i][j].Time)
		}
	}
	for i := range d.Stresses {
		mt = math.Max(mt, d.Stresses[i].Time)
	}
	for i := range d.EqPlasticStrain {
		mt = math.Max(mt, d.EqPlasticStrain[i].Time)
	}
	return
}

type Single struct {
	Name  string
	Time  float64
	Node  int
	Value float64
}

type Pe struct {
	Name     string
	Time     float64
	Elem     int
	IntegPnt int
	Value    float64
}

type Record struct {
	Name string
	Time float64
	Node int

	// (vx,vy,vz)
	// (fx,fy,fz)
	Values [3]float64
}

type Stress struct {
	Name       string
	Time       float64
	Node       int
	IntegPnt   int
	Values     [6]float64 // (elem, integ.pnt.,sxx,syy,szz,sxy,sxz,syz)
	SecondName string
}

func (s Stress) StressIV() float64 {
	v := s.Values
	return math.Sqrt(0.5 *
		(pow.E2(v[0]-v[1]) + pow.E2(v[1]-v[2]) + pow.E2(v[2]-v[0]) +
			6*(pow.E2(v[3])+pow.E2(v[4])+pow.E2(v[5]))))
}

func ParseDat(content []byte) (dat *Dat, err error) {
	dat = new(Dat)

	et := errors.New("ParseDat")
	defer func() {
		if et.IsError() {
			err = et
		}
	}()

	lines := func() []string {
		dat := strings.ReplaceAll(string(content), "\r", "")
		dat = strings.ReplaceAll(dat, "\r", "")
		lines := strings.Split(dat, "\n")
		for i := range lines {
			lines[i] = strings.TrimSpace(lines[i])
		}
		return lines
	}()
	if lines[0] != "" {
		err = fmt.Errorf("not valid first line: `%s`", lines[0])
	}
	lines = lines[1:]

	for pos, err := range []error{
		dat.cleanDat(&lines),
		dat.parseBucklingFactor(&lines),
		dat.parseEigen(&lines),
		dat.parseRecord("displacements (vx,vy,vz)", &dat.Displacements, &lines),
		dat.parseRecord("forces (fx,fy,fz)", &dat.Forces, &lines),
		dat.parseRecord("total force (fx,fy,fz)", &dat.TotalForces, &lines),
		dat.parseSingle("temperatures", &dat.Temperatures, &lines),
		dat.parsePe(&lines),
		dat.parseStresses(&lines),
	} {
		if err != nil {
			_ = et.Add(fmt.Errorf("Pos: %d. %v", pos, err))
		}
	}

	counter := 0
	for i := range lines {
		if lines[i] == "" {
			continue
		}
		_ = et.Add(fmt.Errorf("not parse pos %d: %s", i, lines[i]))
		counter++
		if 5 < counter {
			break
		}
	}

	return
}

func (d *Dat) cleanDat(lines *[]string) error {
	//
	// KNOT1
	// tra      7991  0.2840E-06  0.3126E-06 -0.5253E-08
	// rot     93633 -0.1991E-05 -0.4091E-05 -0.7867E-06
	// exp     93634 -0.1501E-07
	//
	for i := 0; i < len(*lines); i++ {
		if !strings.Contains((*lines)[i], "KNOT1") {
			continue
		}
		(*lines)[i+0] = ""
		(*lines)[i+1] = ""
		(*lines)[i+2] = ""
		(*lines)[i+3] = ""
	}
	return nil
}

// ParseBucklingFactor in file for example `shell2.dat` and return
// slice of buckling factors.
//
//	    B U C K L I N G   F A C T O R   O U T P U T
//
//	MODE NO       BUCKLING
//	               FACTOR
//
//	     1   0.4185108E+03
//	     2   0.4196190E+03
//	     3   0.4200342E+03
//	     4   0.4212441E+03
func (d *Dat) parseBucklingFactor(lines *[]string) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("parseBucklingFactor: %v", err)
		}
	}()
	for i := 0; i < len(*lines); i++ {
		if (*lines)[i] != "B U C K L I N G   F A C T O R   O U T P U T" {
			continue
		}
		counter := 0
		for ; i < len(*lines); i++ {
			(*lines)[i] = ""
			counter++
			if counter == 5 {
				break
			}
		}
		for i += 1; i < len(*lines); i++ {
			if (*lines)[i] == "" {
				break
			}
			fields := strings.Fields((*lines)[i])
			if len(fields) != 2 {
				err = fmt.Errorf("not valid line: `%s`", (*lines)[i])
				return
			}
			var factor float64
			factor, err = parseFloat(fields[1])
			if err != nil {
				return
			}
			d.BucklingFactors = append(d.BucklingFactors, factor)
			(*lines)[i] = ""
		}
	}
	return
}

//	E I G E N V A L U E    N U M B E R     1
//
// displacements (vx,vy,vz) for set NSUMMARY and time  0.0000000E+00
//
//	1  0.000000E+00  0.000000E+00  0.000000E+00
//	2  0.000000E+00  0.000000E+00  0.000000E+00
func (d *Dat) parseEigen(lines *[]string) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("parseBucklingFactor: %v", err)
		}
	}()
	header := "E I G E N V A L U E    N U M B E R"
	for i := 0; i < len(*lines); i++ {
		if !strings.HasPrefix((*lines)[i], header) {
			continue
		}
		mode := (*lines)[i][len(header):]
		mode = strings.TrimSpace(mode)

		var modeNumber int64
		modeNumber, err = strconv.ParseInt(mode, 10, 64)
		if err != nil {
			return
		}
		_ = modeNumber

		(*lines)[i] = ""

		for ; i < len(*lines); i++ {
			if (*lines)[i] != "" {
				break
			}
		}

		var end int = i + 2
		for ; end < len(*lines); end++ {
			if (*lines)[end] == "" {
				break
			}
		}
		sublines := (*lines)[i:end]

		var recs []Record
		err = d.parseRecord("displacements (vx,vy,vz)", &recs, &sublines)
		if err != nil {
			return
		}
		for i := range recs {
			recs[i].Time = float64(modeNumber)
		}
		d.EigenDisplacements = append(d.EigenDisplacements, recs)
		i = end
		for p := i - 1; p <= end; p++ {
			(*lines)[p] = ""
		}
	}

	return
}

// equivalent plastic strain (elem, integ.pnt.,pe)for set ELSUMMARY and time  0.1000000E+00
//
//	1   1  0.000000E+00
//	1   2  0.000000E+00
//	1   3  0.000000E+00
//	1   4  0.000000E+00
//	1   5  0.000000E+00
//	1   6  0.000000E+00
//	1   7  0.000000E+00
//	1   8  0.000000E+00
//	1   9  0.000000E+00
//	2   1  0.000000E+00
func (d *Dat) parsePe(lines *[]string) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("parsePe: %v", err)
		}
	}()
	prefix := "equivalent plastic strain (elem, integ.pnt.,pe)for set"
	for i := 0; i < len(*lines); i++ {
		if !strings.Contains((*lines)[i], prefix) {
			continue
		}
		(*lines)[i] = strings.ReplaceAll((*lines)[i], prefix, "")
		// parse
		var name string
		var time float64
		{
			fs := strings.Fields((*lines)[i])
			if len(fs) != 4 {
				err = fmt.Errorf("not valid: `%s`", (*lines)[i])
				return
			}
			name = fs[0]
			time, err = parseFloat(fs[3])
			if err != nil {
				return
			}
			(*lines)[i] = ""
			(*lines)[i+1] = ""
			i += 2
		}
		for ; i < len(*lines); i++ {
			if (*lines)[i] == "" {
				break
			}
			fields := strings.Fields((*lines)[i])
			var elem int
			elem, err = parseInt(fields[0])
			if err != nil {
				return
			}
			var interpnt int
			interpnt, err = parseInt(fields[1])
			if err != nil {
				return
			}
			var value float64
			value, err = parseFloat(fields[2])
			if err != nil {
				return
			}

			d.EqPlasticStrain = append(d.EqPlasticStrain, Pe{
				Name: name, Time: time,
				Elem: elem, IntegPnt: interpnt, Value: value,
			})
			(*lines)[i] = ""
		}
	}
	return
}

// displacements (vx,vy,vz) for set NALL and time  0.1000000E+01
//
//	1  1.352000E-16 -2.498494E-17  4.970090E-16
//	2  5.598820E-16 -1.501735E-16  1.454478E-15
//
// forces (fx,fy,fz) for set NALL and time  0.1000000E+01
//
//	1 -5.000000E+03 -2.571121E-11  3.895106E-12
//	2 -2.423295E-11  3.754330E-12  8.677503E-12
//
// total force (fx,fy,fz) for set SUPALL and time  0.2500000E+00
//
//	-2.370143E-09  1.371588E+04  1.044280E-11
//
// total force (fx,fy,fz) for set FIX and time  0.1000000E+00
//
//	-8.198390E+02 -3.551087E+02  9.499363E+06
func (d *Dat) parseRecord(header string, recs *[]Record, lines *[]string) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("parseRecord `%s`: %v", header, err)
		}
	}()

	for i := 0; i < len(*lines); i++ {
		if !strings.Contains((*lines)[i], header) {
			continue
		}
		// parse
		fs := strings.Fields((*lines)[i])

		var name string

		for p := range fs {
			if fs[p] == "set" {
				name = fs[p+1]
			}
		}

		var time float64
		time, err = parseFloat(fs[len(fs)-1])
		if err != nil {
			return
		}
		(*lines)[i] = ""
		(*lines)[i+1] = ""
		i += 2
		for ; i < len(*lines); i++ {
			if (*lines)[i] == "" {
				break
			}
			fields := strings.Fields((*lines)[i])
			counter := 0
			var node int
			if !strings.Contains(header, "total force") {
				counter++
				node, err = parseInt(fields[0])
				if err != nil {
					return
				}
			}
			var values [3]float64
			for k := range values {
				values[k], err = parseFloat(fields[k+counter])
				if err != nil {
					return
				}
			}
			(*recs) = append((*recs), Record{
				Name: name, Node: node, Values: values, Time: time})
			(*lines)[i] = ""
		}
	}
	return
}

// temperatures ...
// 1 20.08333
// 2 32.32323
func (d *Dat) parseSingle(header string, recs *[]Single, lines *[]string) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("parseRecord `%s`: %v", header, err)
		}
	}()

	for i := 0; i < len(*lines); i++ {
		if !strings.Contains((*lines)[i], header) {
			continue
		}
		// parse
		fs := strings.Fields((*lines)[i])

		var name string

		for p := range fs {
			if fs[p] == "set" {
				name = fs[p+1]
			}
		}

		var time float64
		time, err = parseFloat(fs[len(fs)-1])
		if err != nil {
			return
		}
		(*lines)[i] = ""
		(*lines)[i+1] = ""
		i += 2
		for ; i < len(*lines); i++ {
			if (*lines)[i] == "" {
				break
			}
			fields := strings.Fields((*lines)[i])
			var node int
			node, err = parseInt(fields[0])
			if err != nil {
				return
			}
			var value float64
			value, err = parseFloat(fields[1])
			if err != nil {
				return
			}
			(*recs) = append((*recs), Single{
				Name: name, Node: node, Value: value, Time: time})
			(*lines)[i] = ""
		}
	}
	return
}

// stresses (elem, integ.pnt.,sxx,syy,szz,sxy,sxz,syz) for set EALL and time  0.1000000E+01
//
//	9   1  1.924780E+01  2.364342E+00  1.339050E+00 -1.499755E+00 -1.734234E+01 -5.334068E+00
//	9   2  1.822731E+01 -1.114173E-01 -1.199380E+01  1.650582E+00 -1.518002E+01  9.306785E-01
//	9   3  1.617674E+01  1.623984E+00 -1.026618E+01  4.567034E+00 -1.570035E+01  7.537470E+00
//	9   4 -1.031847E+01 -1.821434E+00 -6.997414E+00  1.421124E+00  1.661227E+00 -4.074531E+00
//
//	1   1 -4.365744E+02 -1.138162E+02 -1.460370E+02 -9.350144E+00 -7.660134E-01  2.364488E+01 _shell_0000000001`
func (d *Dat) parseStresses(lines *[]string) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("parseStresses: %v", err)
		}
	}()
	for i := 0; i < len(*lines); i++ {
		if !strings.Contains((*lines)[i], "stresses (elem, integ.pnt.,sxx,syy,szz,sxy,sxz,syz)") {
			continue
		}
		// parse
		var name string
		var time float64
		{
			fs := strings.Fields((*lines)[i])
			if len(fs) != 9 {
				err = fmt.Errorf("not valid: `%s`", (*lines)[i])
				return
			}
			name = fs[5]
			time, err = parseFloat(fs[8])
			if err != nil {
				return
			}
			(*lines)[i] = ""
			(*lines)[i+1] = ""
			i += 2
		}
		for ; i < len(*lines); i++ {
			if (*lines)[i] == "" {
				break
			}
			fields := strings.Fields((*lines)[i])
			var node int
			node, err = parseInt(fields[0])
			if err != nil {
				return
			}
			var interpnt int
			interpnt, err = parseInt(fields[1])
			if err != nil {
				return
			}

			var values [6]float64
			for k := range values {
				values[k], err = parseFloat(fields[k+2])
				if err != nil {
					return
				}
			}
			var secName string
			if len(fields) == 9 {
				secName = fields[8]
			}
			d.Stresses = append(d.Stresses, Stress{
				Name: name, IntegPnt: interpnt,
				Node: node, Values: values, Time: time,
				SecondName: secName})
			(*lines)[i] = ""
		}
	}
	return
}
