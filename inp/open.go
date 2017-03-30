package inp

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

/*
// INP stage of file format
*/

//type handler func(f *Format, scan *bufio.Scanner, line string) (ok bool, err error)

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
		stageElementT3D2
		stageElementCPS3
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
				switch element.ElType {
				case TypeT3D2:
					stage = stageElementT3D2
				case TypeCPS3:
					stage = stageElementCPS3
				}
			case strings.Contains(s, "NSET"):
				saveNamedNode(namedNode, inp)
				namedNode, err = convertNamedNode(line)
				if err != nil {
					return err
				}
				if len(namedNode.Name) == 0 {
					return fmt.Errorf("NSET without name - %v", line)
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
		case stageElementT3D2:
			el, err := convertStringToT3D2(line)
			if err != nil {
				return err
			}
			element.Data = append(element.Data, el)
		case stageElementCPS3:
			el, err := convertStringToCPS3(line)
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
	{
		r := strings.Split(s[1], "=")
		for i := range r {
			r[i] = strings.TrimSpace(r[i])
		}
		if len(r) != 2 {
			return el, fmt.Errorf("Wrong in second element - %v", line)
		}
		switch r[1] {
		case "T3D2":
			el.ElType = TypeT3D2
		case "CPS3":
			el.ElType = TypeCPS3
		default:
			return el, fmt.Errorf("Wrong type of element - %v", line)
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
func convertStringToT3D2(line string) (c ElementData, err error) {
	s := separate(line)
	if len(s) != 3 {
		return c, fmt.Errorf("Wrong string line for convert coordinate of point.\nLine = %v.\nSlice=%v", line, s)
	}
	i, err := strconv.ParseInt(s[0], 10, 64)
	if err != nil {
		return
	}
	c.Index = int(i)
	for i := 0; i < 2; i++ {
		point, err := strconv.ParseInt(s[1+i], 10, 64)
		if err != nil {
			return c, err
		}
		c.IPoint = append(c.IPoint, int(point))
	}
	return c, err
}

// *ELEMENT, type=CPS3, ELSET=Surface17
// 1906, 39, 234, 247
func convertStringToCPS3(line string) (c ElementData, err error) {
	s := separate(line)
	if len(s) != 4 {
		return c, fmt.Errorf("Wrong string line for convert coordinate of point.\nLine = %v.\nSlice=%v", line, s)
	}
	i, err := strconv.ParseInt(s[0], 10, 32)
	c.Index = int(i)
	if err != nil {
		return c, err
	}
	for i := 0; i < 3; i++ {
		point, err := strconv.ParseInt(s[1+i], 10, 64)
		if err != nil {
			return c, err
		}
		c.IPoint = append(c.IPoint, int(point))
	}
	return c, err
}

/*
	handlerList := []handler{
		emptyHandler,
		commentHandler,
		headingHandler,
	}

	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		for _, handler := range handlerList {
			ok, err := handler(&inp, scanner, line)
			if err != nil {
				return err
			}
			if ok {
				goto NextLine
			}
		}
		return fmt.Errorf("No found handler for that line : %v", line)
	NextLine:
	}
	return nil
}

func emptyHandler(f *Format, scan *bufio.Scanner, line string) (ok bool, err error) {
	if len(line) < 1 {
		return true, nil
	}
	return false, nil
}

// ** COMMENTS
func commentHandler(f *Format, scan *bufio.Scanner, line string) (ok bool, err error) {
	if len(line) < 2 {
		return false, nil
	}
	if line[0] == '*' && line[1] == '*' {
		return true, nil
	}
	return false, nil
}

// *Heading
//  cone.inp
func headingHandler(f *Format, scan *bufio.Scanner, line string) (ok bool, err error) {
	word := "HEADING"
	if len(line) <= len(word) {
		return false, nil
	}
	if line[0] != '*' {
		return false, nil
	}
	if strings.ToUpper(strings.TrimSpace(line[1:])) != word {
		return false, nil
	}
	// scanning
	if scan.Scan() {
		line = scan.Text()
		line = strings.TrimSpace(line)
		f.Name = strings.TrimSpace(line)
		return true, nil
	}
	return false, fmt.Errorf("Cannot scan in *heading")
}

// *NODE
// 1, 0, 0, 0

// *ELEMENT, type=T3D2, ELSET=Line1
// 7, 1, 7

// *ELEMENT, type=CPS3, ELSET=Surface17
// 1906, 39, 234, 247
*/
