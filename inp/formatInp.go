package inp

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
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

// Coordinate - coordinate in inp format
type Coordinate struct {
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

// Element - inp element
type Element struct {
	ElType ElementType
	Index  int
	IPoint []int
}

// Format - summary inp format
type Format struct {
	Nodes    []Coordinate
	Elements []Element
}

// INP stage of file format
type stageReading uint

const (
	stageBug stageReading = iota
	stageHeading
	stageNode
	stageElementT3D2
	stageElementCPS3
)

// ReadInp - read file in inp format
func (inp Format) ReadInp(file string) (err error) {
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

	// stage == 0 - Heading
	// stage == 1 - Node
	// stage == 2 - Element T3D2
	// stage == 3 - Element CPS3
	var stage stageReading

	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case len(line) == 0:
			continue
		case line[0] == '*':
			stage = getStage(line)
			continue
		}

		switch stage {
		case stageHeading:
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
			inp.Elements = append(inp.Elements, el)
		case stageElementCPS3:
			el, err := convertStringToCPS3(line)
			if err != nil {
				return err
			}
			inp.Elements = append(inp.Elements, el)
		case stageBug:
			return fmt.Errorf("stageBug. Cannot convert line = %v", line)
		}
	}

	return
}

func getStage(line string) stageReading {
	switch {
	case strings.HasPrefix(line, "*Heading"):
		return stageHeading
	case strings.HasPrefix(line, "*NODE"):
		return stageNode
	case strings.HasPrefix(line, "*ELEMENT"):
		switch {
		case strings.Contains(line, "T3D2"):
			return stageElementT3D2
		case strings.Contains(line, "CPS3"):
			return stageElementCPS3
		}
	}
	return stageBug
}

// *NODE
// 1, 0, 0, 0
func convertStringToNode(line string) (c Coordinate, err error) {
	s := strings.Split(line, ",")
	if len(s) != 4 {
		return c, fmt.Errorf("Wrong string line for convert coordinate of point.\nLine = %v.\nSlice=%v", line, s)
	}
	for i := range s {
		s[i] = strings.TrimSpace(s[i])
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
func convertStringToT3D2(line string) (c Element, err error) {
	c.ElType = TypeT3D2
	s := strings.Split(line, ",")
	if len(s) != 3 {
		return c, fmt.Errorf("Wrong string line for convert coordinate of point.\nLine = %v.\nSlice=%v", line, s)
	}
	for i := range s {
		s[i] = strings.TrimSpace(s[i])
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
func convertStringToCPS3(line string) (c Element, err error) {
	c.ElType = TypeCPS3
	s := strings.Split(line, ",")
	if len(s) != 4 {
		return c, fmt.Errorf("Wrong string line for convert coordinate of point.\nLine = %v.\nSlice=%v", line, s)
	}
	for i := range s {
		s[i] = strings.TrimSpace(s[i])
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
