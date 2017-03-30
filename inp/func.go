package inp

import "math"

// AddUniqueIndexToElements - add unique index for element with Index == -1
func (f *Format) AddUniqueIndexToElements() {
	var maxIndexElement int
	for _, element := range f.Elements {
		for _, data := range element.Data {
			if data.Index > maxIndexElement {
				maxIndexElement = data.Index
			}
		}
	}
	if maxIndexElement <= 0 {
		maxIndexElement = 1
	}
	// add unique index only if "Index == -1"
	for iE, element := range f.Elements {
		for iD, data := range element.Data {
			if data.Index == -1 {
				maxIndexElement++
				f.Elements[iE].Data[iD].Index = maxIndexElement
			}
		}
	}
}

// AddNamedNodesOnLevel - add named nodes on specific elevation with name
func (f *Format) AddNamedNodesOnLevel(level float64, name string) {
	eps := 1e-8
	var n NamedNode
	n.Name = name
	for _, node := range f.Nodes {
		y := node.Coord[1]
		if math.Abs(y-level) <= eps {
			n.Nodes = append(n.Nodes, node.Index)
		}
	}
	if len(n.Nodes) > 0 {
		f.NodesWithName = append(f.NodesWithName, n)
	}
}
