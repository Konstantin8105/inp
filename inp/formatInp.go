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

/*
*MATERIAL, NAME=stell
*ELASTIC
2.9E+07,0.28

*DENSITY
7.35E-4
*EXPANSION
7.228E-6

*SHELL SECTION, MATERIAL=stell,ELSET=shell,OFFSET=0
0.005

*BOUNDARY
Bottom,1,1,0

*STEP
*BUCKLE
5
*CLOAD
Top,2,-100

*NODE FILE
U
*NODE PRINT
U,NT,RF
*NODE FILE
U,NT,RF
*EL FILE
U,S
*EL PRINT
S
*END STEP
*/

// Node - coordinate in inp format
type Node struct {
	Index int
	Coord [3]float64
}

// ElementData - inp elements
type ElementData struct {
	Index  int
	IPoint []int
}

// Element - inp element
type Element struct {
	Name string
	FE   *FiniteElement
	Data []ElementData
}

// NamedNode - list of nodes with specific name
type NamedNode struct {
	Name  string
	Nodes []int
}

// ShellSection - add thickness for shell elements
type ShellSection struct {
	ElementName string
	Thickness   float64
}

// BoundaryProperty - fixed point
// For structures:
// – 1: translation in the local x-direction
// – 2: translation in the local y-direction
// – 3: translation in the local z-direction
// – 4: rotation about the local x-axis (only for nodes belonging to beams or shells)
// – 5: rotation about the local y-axis (only for nodes belonging to beams or shells)
// – 6: rotation about the local z-axis (only for nodes belonging to beams or shells)
type BoundaryProperty struct {
	NodesByName   string
	StartFreedom  int
	FinishFreedom int
	Value         float64
}

//StepProperty - property of load case
type StepProperty struct {
	AmountBucklingShapes int
	Loads                []Load
}

// Load - load
type Load struct {
	NodesByName string
	Direction   int
	LoadValue   float64
}

// Format - summary inp format
type Format struct {
	Name          string
	Nodes         []Node
	Elements      []Element
	NodesWithName []NamedNode
	ShellSections []ShellSection
	Boundary      []BoundaryProperty
	Step          StepProperty
}

var materialProperty string
var stepProperty string

func init() {
	materialProperty = `
*MATERIAL, NAME=stell
*ELASTIC
2.9E+07,0.28

*DENSITY
7.35E-4
*EXPANSION
7.228E-6
`
	stepProperty = `
*NODE FILE
U
*NODE PRINT
U,NT,RF
*NODE FILE
U,NT,RF
*EL FILE
U,S
*EL PRINT
S
`
}
