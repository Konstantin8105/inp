package inp

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Open - open file in inp format
func (inp *Format) Open(file string) (err error) {
	inFile, err := os.Open(file)
	if err != nil {
		return
	}
	defer func() {
		errFile := inFile.Close()
		if errFile != nil {
			if err != nil {
				err = fmt.Errorf("%v ; %v", err, errFile)
			} else {
				err = errFile
			}
		}
	}()
	scanner := bufio.NewScanner(inFile)
	scanner.Split(bufio.ScanLines)

	type stageReading uint

	const (
		stageHeading stageReading = iota
		stageNode
		stageElement
		stageNamedNode
	)

	var stage stageReading
	var element Element
	var namedNode NamedNode

	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		// empty line
		if len(line) == 0 {
			continue
		}

		// comments
		if len(line) >= 2 && line[0] == '*' && line[1] == '*' {
			continue
		}

		// change stage
		if line[0] == '*' {
			s := strings.ToUpper(line)
			switch {
			case strings.Contains(s, "HEADING"):
				stage = stageHeading
			case strings.Contains(s, "NODE"):
				stage = stageNode
			case strings.Contains(s, "ELEMENT"):
				saveElement(element, inp)
				element, err = convertElement(line)
				if err != nil {
					return err
				}
				stage = stageElement
			case strings.Contains(s, "NSET"):
				saveNamedNode(namedNode, inp)
				namedNode, err = convertNamedNode(line)
				if err != nil {
					return err
				}
				stage = stageNamedNode
			default:
				return fmt.Errorf("Cannot found type for that line : %v", line)
			}
			continue
		}

		switch stage {
		case stageHeading:
			inp.Name = line
			continue
		case stageNode:
			node, err := convertStringToNode(line)
			if err != nil {
				return err
			}
			inp.Nodes = append(inp.Nodes, node)
		case stageElement:
			el, err := convertStringToElement(element, line)
			if err != nil {
				return err
			}
			element.Data = append(element.Data, el)
		case stageNamedNode:
			index, err := convertStringToNameIndex(line)
			if err != nil {
				return err
			}
			namedNode.Nodes = append(namedNode.Nodes, index)
		}
	}
	saveElement(element, inp)
	saveNamedNode(namedNode, inp)

	return nil
}

func saveNamedNode(namedNode NamedNode, inp *Format) {
	if len(namedNode.Nodes) == 0 {
		return
	}
	inp.NodesWithName = append(inp.NodesWithName, namedNode)
}

func saveElement(element Element, inp *Format) {
	if len(element.Data) == 0 {
		return
	}
	inp.Elements = append(inp.Elements, element)
}

// convert named nodes
// *NSET, NSET = name
func convertNamedNode(line string) (namedNode NamedNode, err error) {
	s := strings.Split(line, ",")
	for i := range s {
		s[i] = strings.TrimSpace(s[i])
	}
	{
		r := strings.Split(s[1], "=")
		for i := range r {
			r[i] = strings.TrimSpace(r[i])
		}
		if len(r) != 2 {
			return namedNode, fmt.Errorf("Wrong in second NSET - %v", line)
		}
		namedNode.Name = strings.TrimSpace(r[1])
		if len(namedNode.Name) == 0 {
			return namedNode, fmt.Errorf("Name is empty and this is not acceptable - %v", line)
		}
	}
	return namedNode, nil
}

// convert element
// *ELEMENT, type=CPS3, ELSET=shell
func convertElement(line string) (el Element, err error) {
	s := strings.Split(line, ",")
	for i := range s {
		s[i] = strings.TrimSpace(s[i])
	}
	// found the type
	{
		r := strings.Split(s[1], "=")
		for i := range r {
			r[i] = strings.ToUpper(strings.TrimSpace(r[i]))
		}
		if len(r) != 2 {
			return el, fmt.Errorf("Wrong in second element - %v", line)
		}
		var found bool
		for i, f := range FiniteElementDatabase {
			if r[1] == f.Name {
				el.FE = &FiniteElementDatabase[i]
				found = true
				break
			}
		}
		if !found {
			return el, fmt.Errorf("Cannot convert to finite element - %v", line)
		}
	}
	{
		r := strings.Split(s[2], "=")
		for i := range r {
			r[i] = strings.TrimSpace(r[i])
		}
		if len(r) != 2 {
			return el, fmt.Errorf("Wrong in 3 element - %v", line)
		}
		el.Name = r[1]
		if len(el.Name) == 0 {
			return el, fmt.Errorf("Name is empty and this is not acceptable - %v", line)
		}
	}
	return el, nil
}

// separate by , and trim
func separate(line string) (s []string) {
	s = strings.Split(line, ",")
	for i := range s {
		s[i] = strings.TrimSpace(s[i])
	}
	return s
}

// convert index of node in string to int
// 1,
// 5921,
func convertStringToNameIndex(line string) (index int, err error) {
	s := separate(line)
	i, err := strconv.ParseInt(s[0], 10, 64)
	if err != nil {
		return
	}
	return int(i), nil
}

// *NODE
// 1, 0, 0, 0
func convertStringToNode(line string) (c Node, err error) {
	s := separate(line)
	if len(s) != 4 {
		return c, fmt.Errorf("Wrong string line for convert coordinate of point.\nLine = %v.\nSlice=%v", line, s)
	}
	i, err := strconv.ParseInt(s[0], 10, 64)
	if err != nil {
		return c, err
	}
	c.Index = int(i)
	for i := 0; i < 3; i++ {
		c.Coord[i], err = strconv.ParseFloat(s[1+i], 64)
		if err != nil {
			return c, err
		}
	}
	return c, err
}

// *ELEMENT, type=T3D2, ELSET=Line1
// 7, 1, 7
// *ELEMENT, type=CPS3, ELSET=Surface17
// 1906, 39, 234, 247
func convertStringToElement(el Element, line string) (c ElementData, err error) {
	s := separate(line)
	if el.FE == nil {
		return c, fmt.Errorf("Error in convertStringToElement: element is nil")
	}
	if len(s) != el.FE.AmountNodes+1 {
		return c, fmt.Errorf("Wrong string line for convert coordinate of point.\nLine = %v.\nSlice=%v", line, s)
	}
	var array []int
	for i := 0; i < el.FE.AmountNodes+1; i++ {
		result, err := strconv.ParseInt(s[0], 10, 64)
		if err != nil {
			return c, fmt.Errorf("Cannot convert to int - %v on line - %v", s[i], line)
		}
		array = append(array, int(result))
	}

	c.Index = array[0]
	c.IPoint = array[1:]

	return c, err
}
