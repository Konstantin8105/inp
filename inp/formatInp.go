package inp

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

// Node - coordinate in inp format
type Node struct {
	Index int
	Coord [3]float64
}

// ElementType - type inp of beam, triangle...
type ElementType uint

// Types of inp elements
const (
	TypeT3D2 ElementType = iota // beam
	TypeCPS3                    // triangle
)

// ElementData - inp elements
type ElementData struct {
	Index  int
	IPoint []int
}

// Element - inp element
type Element struct {
	ElType ElementType
	Name   string
	Data   []ElementData
}

// Format - summary inp format
type Format struct {
	Name     string
	Nodes    []Node
	Elements []Element
}
