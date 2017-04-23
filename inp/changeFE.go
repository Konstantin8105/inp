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
func (f *Format) ChangeTypeFiniteElement(from *FiniteElement, to *FiniteElement) (err error) {
	if from == to {
		return nil
	}

	if from.Shape == to.Shape && from.AmountNodes == to.AmountNodes {
		// modify finite element with middle point
		for elemenentI := range f.Elements {
			if f.Elements[elemenentI].FE.Name != from.Name {
				continue
			}
			f.Elements[elemenentI].FE = to
		}
		return nil
	}

	if from.Shape != to.Shape {
		if from.AmountNodes == 4 && to.AmountNodes == 3 {
			s3, _ := GetFiniteElementByName("S3")
			f.changeFEfromQuadraticToTriangle(from, s3)
			return nil
		}
		if from.AmountNodes == 4 && to.AmountNodes == 6 {
			s3, _ := GetFiniteElementByName("S3")
			err = f.ChangeTypeFiniteElement(from, s3)
			if err != nil {
				return err
			}
			err = f.ChangeTypeFiniteElement(s3, to)
			if err != nil {
				return err
			}
			return nil
		}
	}

	if from.Shape == to.Shape && from.AmountNodes*2 == to.AmountNodes {

		// divide middle point inside exist
		group, err := f.createMiddlePoint(from)
		if err != nil {
			return fmt.Errorf("Wrong in createMiddlePoint: %v", err)
		}

		// add points in format
		for _, node := range group {
			f.Nodes = append(f.Nodes, node.nodeC)
		}

		// modify finite element with middle point
		for elemenentI := range f.Elements {
			if f.Elements[elemenentI].FE.Name != from.Name {
				continue
			}
			f.Elements[elemenentI].FE = to
			for iData := range f.Elements[elemenentI].Data {
				iPoints := f.Elements[elemenentI].Data[iData].IPoint
				// modification
				var newPoints []int
				for index := range iPoints {
					var pointIndex1 int
					if index == 0 {
						pointIndex1 = iPoints[len(iPoints)-1]
					} else {
						pointIndex1 = iPoints[index-1]
					}
					pointIndex2 := iPoints[index]
					var newPoint int
					if pointIndex1 > pointIndex2 {
						newPoint, err = f.foundPointCIndexInLineGroup(pointIndex2, pointIndex1, &group)
					} else {
						newPoint, err = f.foundPointCIndexInLineGroup(pointIndex1, pointIndex2, &group)
					}
					if err != nil {
						return fmt.Errorf("Cannot found point in lineGroup : %v", err)
					}
					newPoints = append(newPoints, newPoint)
				}
				// end of modification
				for i := range newPoints {
					if i == len(newPoints)-1 {
						f.Elements[elemenentI].Data[iData].IPoint = append(f.Elements[elemenentI].Data[iData].IPoint, newPoints[0])
					} else {
						f.Elements[elemenentI].Data[iData].IPoint = append(f.Elements[elemenentI].Data[iData].IPoint, newPoints[i+1])
					}
				}
			}
		}

		// NodeNames changes
		if len(f.NodesWithName) != 0 {
			return fmt.Errorf("Cannot work with Named nodes")
		}

		return nil
	}

	return fmt.Errorf("Cannot change FE from %v to %v", from, to)
}

func (f *Format) foundPointCIndexInLineGroup(p1, p2 int, group *[]lineGroup) (middlePoint int, err error) {
	if p1 > p2 {
		return -1, fmt.Errorf("Case p1 < p2 is not correct")
	}
	for _, g := range *group {
		if g.indexA == p1 && g.indexB == p2 {
			return g.nodeC.Index, nil
		}
	}
	return -1, fmt.Errorf("Cannot found in group with point %v,%v\nGroup = %v", p1, p2, *group)
}

func (f *Format) createMiddlePoint(fe *FiniteElement) (group []lineGroup, err error) {
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
		if element.FE.Name != fe.Name {
			continue
		}

		for _, data := range element.Data {
			for index := range data.IPoint {
				var pointIndex1 int
				if index == 0 {
					pointIndex1 = data.IPoint[len(data.IPoint)-1]
				} else {
					pointIndex1 = data.IPoint[index-1]
				}
				pointIndex2 := data.IPoint[index]
				var g lineGroup
				if pointIndex1 > pointIndex2 {
					g = lineGroup{indexA: pointIndex2, indexB: pointIndex1}
				} else {
					g = lineGroup{indexA: pointIndex1, indexB: pointIndex2}
				}
				group = append(group, g)
			}
		}
	}

	// sorting linegroup
	sort.Sort(byIndexA(group))
	for {
		var isChange bool
		for i := range group {
			if i == 0 {
				continue
			}
			if group[i-1].indexA != group[i].indexA {
				continue
			}
			if group[i-1].indexB > group[i].indexB {
				// swap
				group[i-1].indexB, group[i].indexB = group[i].indexB, group[i-1].indexB
				isChange = true
			}
		}
		if !isChange {
			break
		}
	}

	// create unique slice : true - if unique
	unique := make([]bool, len(group), len(group))
	for index := range group {
		if index == 0 {
			unique[0] = true
			continue
		}
		unique[index] = !(group[index-1].indexA == group[index].indexA && group[index-1].indexB == group[index].indexB)
	}

	amount := 0
	for _, u := range unique {
		if u {
			amount++
		}
	}

	// create unique linegroup
	var buffer []lineGroup
	for i, u := range unique {
		if u {
			buffer = append(buffer, group[i])
		}
	}
	group = buffer

	// 2-step for calculate middle point
	for index := range group {

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
		group[index].nodeC.Index = maximalIndex
		maximalIndex++
	}

	return group, nil
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

func (f *Format) changeFEfromQuadraticToTriangle(from *FiniteElement, to *FiniteElement) {
	var maximalIndex int
	for _, element := range f.Elements {
		for _, data := range element.Data {
			if maximalIndex < data.Index {
				maximalIndex = data.Index
			}
		}
	}
	maximalIndex++

	// add new elements
	for elemenentI := range f.Elements {
		if f.Elements[elemenentI].FE.Name != from.Name {
			continue
		}
		var newElement Element
		newElement.Name = f.Elements[elemenentI].Name
		newElement.FE = to
		for iData := range f.Elements[elemenentI].Data {
			newElement.Data = append(newElement.Data, ElementData{Index: maximalIndex, IPoint: []int{
				f.Elements[elemenentI].Data[iData].IPoint[0],
				f.Elements[elemenentI].Data[iData].IPoint[1],
				f.Elements[elemenentI].Data[iData].IPoint[2],
			}})
			maximalIndex++
			newElement.Data = append(newElement.Data, ElementData{Index: maximalIndex, IPoint: []int{
				f.Elements[elemenentI].Data[iData].IPoint[2],
				f.Elements[elemenentI].Data[iData].IPoint[3],
				f.Elements[elemenentI].Data[iData].IPoint[0],
			}})
			maximalIndex++
		}
		f.Elements = append(f.Elements, newElement)
	}
	// remove old FE
AGAIN:
	for elemenentI := range f.Elements {
		if f.Elements[elemenentI].FE.Name != from.Name {
			continue
		}
		f.Elements = append(f.Elements[:elemenentI], f.Elements[(elemenentI+1):]...)
		goto AGAIN
	}

	return
}
