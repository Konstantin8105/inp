package inp

import (
	"fmt"
	"sort"
)

// lineGroup - group of points
//  *------*------*
//  A      C      B
// points A,B - exist index of points
// point C - new point
type lineGroup struct {
	indexA, indexB int
	nodeC          Node
}

// for sorting by indexA
type byIndexA []lineGroup

func (l byIndexA) Len() int {
	return len(l)
}
func (l byIndexA) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}
func (l byIndexA) Less(i, j int) bool {
	return l[i].indexA < l[j].indexA
}

// ChangeTypeFiniteElement - change type finite element for example
// from S4 to S8
func (f *Format) ChangeTypeFiniteElement(from FiniteElement, to FiniteElement) (out Format, err error) {
	if from.Name != "S4" && to.Name != "S8" {
		return out, fmt.Errorf("Cannot change that finite element : from %v  to %v", from, to)
	}
	// divide middle point inside exist
	group, err := f.createMiddlePoint(from)
	if err != nil {
		return out, fmt.Errorf("Wrong in createMiddlePoint: %v", err)
	}

	// add points in format

	// modify finite element with middle point
}

func (f *Format) createMiddlePoint(fe FiniteElement) (group []lineGroup, err error) {
	// check slice of nodes inp format - index must by from less to more
	// if it is true, then we can use binary sort for fast found the point
	for index := range f.Nodes {
		if index == 0 {
			continue
		}
		if f.Nodes[index-1].Index >= f.Nodes[index].Index {
			return nil, fmt.Errorf("Please sort the nodes in inp format")
		}
	}

	// create slice of linegroup
	for _, element := range f.Elements {
		if element.FE.Name == fe.Name {
			for index := range element.Data {
				if index == 0 {
					continue
				}
				pointIndex1 := element.Data[index-1].Index
				var pointIndex2 int
				if index != len(element.Data)-1 {
					pointIndex2 = element.Data[index].Index
				} else {
					pointIndex2 = element.Data[0].Index
				}
				if pointIndex1 > pointIndex2 {
					group = append(group, lineGroup{indexA: pointIndex2, indexB: pointIndex1})
				} else {
					group = append(group, lineGroup{indexA: pointIndex1, indexB: pointIndex2})
				}
			}
		}
	}
	// sorting linegroup
	sort.Sort(byIndexA(group))

	// create unique slice : true - if unique
	unique := make([]bool, len(group), len(group))
	for index := range group {
		if index == 0 {
			unique[0] = true
		}
		unique[index] = group[index-1].indexA == group[index].indexA && group[index-1].indexB == group[index].indexB
	}

	fmt.Println("unique len = ", len(unique))
	amount := 0
	for _, u := range unique {
		if u {
			amount++
		}
	}
	fmt.Println("amount unique = ", amount)

	// 2-step for calculate middle point
	for index := range group {
		if !unique[index] {
			continue
		}
		// step 1: loop - add to nodeC coordinate of NodeA
		group[index].nodeC.Coord, err = f.foundByIndex(group[index].indexA)
		if err != nil {
			return nil, fmt.Errorf("Cannot found point with index : %v", group[index].indexA)
		}
		// step 2: loop - calculate nodeC = (nodeC+nodeB)/2.
		coord, err := f.foundByIndex(group[index].indexB)
		if err != nil {
			return nil, fmt.Errorf("Cannot found point with index : %v", group[index].indexB)
		}
		// calculate middle
		for i := 0; i < 3; i++ {
			group[index].nodeC.Coord[i] += coord[i]
			group[index].nodeC.Coord[i] /= 2.0
		}
	}

	// find maximal index of point
	maximalIndex := f.Nodes[0].Index
	for index := range f.Nodes {
		if maximalIndex < f.Nodes[index].Index {
			maximalIndex = f.Nodes[index].Index
		}
	}
	maximalIndex++

	// add index to indexC
	for index := range group {
		if !unique[index] {
			continue
		}
		group[index].nodeC.Index = maximalIndex
		maximalIndex++
	}

	// create unique linegroup
	var buffer []lineGroup
	for i, u := range unique {
		if u {
			buffer = append(buffer, group[i])
		}
	}

	return buffer, nil
}

func (f *Format) foundByIndex(index int) (node [3]float64, err error) {
	i := sort.Search(len(f.Nodes), func(a int) bool { return f.Nodes[a].Index >= index })
	if i < len(f.Nodes) && f.Nodes[i].Index == index {
		// index is present at nodes
		return f.Nodes[i].Coord, nil
	}
	// index is not present in nodes,
	// but i is the index where it would be inserted.
	return node, fmt.Errorf("Cannot found in sort.Search : %v, but i = %v", index, i)
}
