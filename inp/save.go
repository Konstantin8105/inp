package inp

import (
	"fmt"
	"sort"
	"strings"

	"github.com/Konstantin8105/Convert-INP-to-STD-format/utils"
)

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

type pp []Node

func (a pp) Len() int           { return len(a) }
func (a pp) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a pp) Less(i, j int) bool { return a[i].Index < a[j].Index }

// Save - convertor
func (f Format) Save(filename string) (err error) {
	if len(f.Name) == 0 {
		f.Name = filename
	}
	err = utils.CreateNewFile(filename, f.SaveINPtoLines())
	return err

}

// SaveINPtoLines - converting
func (f Format) SaveINPtoLines() (lines []string) {

	// TODO fast append strings
http: //stackoverflow.com/questions/1760757/how-to-efficiently-concatenate-strings-in-go
	// TODO check  benchmark next line
	lines = make([]string, 0, len(f.Elements)+len(f.Nodes)+10)

	lines = append(lines, "*HEADING")
	f.Name = strings.TrimSpace(f.Name)
	if len(f.Name) == 0 {
		f.Name = "Convertor"
	}
	lines = append(lines, f.Name)

	// sort points by index
	sort.Sort(pp(f.Nodes))

	lines = append(lines, "*NODE")
	for _, node := range f.Nodes {
		lines = append(lines, fmt.Sprintf("%v, %.10e, %.10e, %.10e", node.Index, node.Coord[0], node.Coord[1], node.Coord[2]))
	}

	lines = append(lines, "**** ELEMENTS ****")
	for _, element := range f.Elements {
		element.Name = strings.TrimSpace(element.Name)
		if len(element.Name) == 0 {
			element.Name = "Convertor"
		}
		lines = append(lines, fmt.Sprintf("*ELEMENT, type=%v, ELSET=%v", element.FE.Name, element.Name))
		for _, data := range element.Data {
			s := fmt.Sprintf("%v", data.Index)
			for _, point := range data.IPoint {
				s += fmt.Sprintf(",%v", point)
			}
			lines = append(lines, s)
		}
	}

	lines = append(lines, "**** Named nodes ****")
	for _, n := range f.NodesWithName {
		lines = append(lines, fmt.Sprintf("*NSET,NSET=%v", n.Name))
		for _, i := range n.Nodes {
			lines = append(lines, fmt.Sprintf("%v,", i))
		}
	}

	lines = append(lines, "**** Property of material ****")
	lines = append(lines, materialProperty)

	lines = append(lines, "**** Shell property ****")
	for _, s := range f.ShellSections {
		lines = append(lines, fmt.Sprintf("*SHELL SECTION,MATERIAL=steel,ELSET=%v,,OFFSET=0", s.ElementName))
		lines = append(lines, fmt.Sprintf("%.10e", s.Thickness))
	}

	lines = append(lines, "**** Boundary property ****")
	for _, b := range f.Boundary {
		lines = append(lines, "*BOUNDARY")
		lines = append(lines, fmt.Sprintf("%v,%v,%v,%v", b.NodesByName, b.StartFreedom, b.FinishFreedom, b.Value))
	}

	lines = append(lines, "**** STEP PROPERTY ****")
	if f.Step.AmountBucklingShapes > 0 || len(f.Step.Loads) > 0 {
		lines = append(lines, "*STEP")
		if f.Step.AmountBucklingShapes > 0 {
			lines = append(lines, "*BUCKLE")
			lines = append(lines, fmt.Sprintf("%v", f.Step.AmountBucklingShapes))
		}
		if len(f.Step.Loads) > 0 {
			for _, l := range f.Step.Loads {
				lines = append(lines, "*CLOAD")
				lines = append(lines, fmt.Sprintf("%v,%v,%.10e", l.NodesByName, l.Direction, l.LoadValue))
			}
		}
		lines = append(lines, stepProperty)
		lines = append(lines, "*END STEP")
	}

	return lines
}
