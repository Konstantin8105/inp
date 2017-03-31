package inp

import "fmt"

// FiniteElementShape - const of finite element shape
type FiniteElementShape int

// Shapes of finite element
const (
	Triangle FiniteElementShape = iota
	Quadratic
	Beam
)

// FiniteElement - information about finite element
type FiniteElement struct {
	Shape       FiniteElementShape
	AmountNodes int
	Name        string
	Description string
}

// FiniteElementDatabase - information about all allowable finite elements
var FiniteElementDatabase []FiniteElement

func init() {
	FiniteElementDatabase = []FiniteElement{
		FiniteElement{
			Shape:       Triangle,
			AmountNodes: 3,
			Name:        "CPS3",
			Description: "Three-node plane stress element",
		},
		FiniteElement{
			Shape:       Beam,
			AmountNodes: 2,
			Name:        "T3D2",
			Description: "Two-node truss element",
		},
		FiniteElement{
			Shape:       Triangle,
			AmountNodes: 3,
			Name:        "S3",
			Description: "Three-node shell element",
		},
		FiniteElement{
			Shape:       Quadratic,
			AmountNodes: 4,
			Name:        "S4",
			Description: "Four-node shell element",
		},
		FiniteElement{
			Shape:       Quadratic,
			AmountNodes: 4,
			Name:        "S4R",
			Description: "Four-node shell element",
		},
		FiniteElement{
			Shape:       Triangle,
			AmountNodes: 6,
			Name:        "S6",
			Description: "Six-node shell element",
		},
		FiniteElement{
			Shape:       Quadratic,
			AmountNodes: 8,
			Name:        "S8",
			Description: "Eight-node shell element",
		},
		FiniteElement{
			Shape:       Quadratic,
			AmountNodes: 8,
			Name:        "S8R",
			Description: "Eight-node shell element",
		},
	}
}

func GetFiniteElementByName(name string) (fe *FiniteElement, err error) {
	for i := range FiniteElementDatabase {
		if name == FiniteElementDatabase[i].Name {
			return &FiniteElementDatabase[i], nil
		}
	}
	return nil, fmt.Errorf("Cannot found finite element by name - %v", name)
}
