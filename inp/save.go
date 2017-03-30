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
	err = utils.CreateNewFile(filename, f.saveINPtoLines())
	return err

}

func (f Format) saveINPtoLines() (lines []string) {
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
		lines = append(lines, fmt.Sprintf("*ELEMENT, type=%v, ELSET=%v", element.ElType, element.Name))
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
	return lines
}
