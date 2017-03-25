package convertorInp

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
type inpCoordinate struct {
	index uint64
	coord [3]float64
}

type inpElementType uint

const (
	inpTypeT3D2 inpElementType = iota
	inpTypeCPS3
)

type inpElement struct {
	elType inpElementType
	index  uint64
	iPoint []uint64
}

type inpFormat struct {
	nodes    []inpCoordinate
	elements []inpElement
}

// INP stage of file format
type inpStageReading uint

const (
	stageBug inpStageReading = iota
	stageHeading
	stageNode
	stageElementT3D2
	stageElementCPS3
)

func readInp(file string) (inp inpFormat, err error) {
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
	var stage inpStageReading

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
				return inp, err
			}
			inp.nodes = append(inp.nodes, node)
		case stageElementT3D2:
			el, err := convertStringToT3D2(line)
			if err != nil {
				return inp, err
			}
			inp.elements = append(inp.elements, el)
		case stageElementCPS3:
			el, err := convertStringToCPS3(line)
			if err != nil {
				return inp, err
			}
			inp.elements = append(inp.elements, el)
		case stageBug:
			return inp, fmt.Errorf("stageBug. Cannot convert line = %v", line)
		}
	}

	return
}

func getStage(line string) inpStageReading {
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
func convertStringToNode(line string) (c inpCoordinate, err error) {
	s := strings.Split(line, ",")
	if len(s) != 4 {
		return c, fmt.Errorf("Wrong string line for convert coordinate of point.\nLine = %v.\nSlice=%v", line, s)
	}
	for i := range s {
		s[i] = strings.TrimSpace(s[i])
	}
	c.index, err = strconv.ParseUint(s[0], 10, 64)
	if err != nil {
		return c, err
	}
	for i := 0; i < 3; i++ {
		c.coord[i], err = strconv.ParseFloat(s[1+i], 64)
		if err != nil {
			return c, err
		}
	}
	return c, err
}

// *ELEMENT, type=T3D2, ELSET=Line1
// 7, 1, 7
func convertStringToT3D2(line string) (c inpElement, err error) {
	c.elType = inpTypeT3D2
	s := strings.Split(line, ",")
	if len(s) != 3 {
		return c, fmt.Errorf("Wrong string line for convert coordinate of point.\nLine = %v.\nSlice=%v", line, s)
	}
	for i := range s {
		s[i] = strings.TrimSpace(s[i])
	}
	c.index, err = strconv.ParseUint(s[0], 10, 64)
	if err != nil {
		return
	}
	for i := 0; i < 2; i++ {
		point, err := strconv.ParseUint(s[1+i], 10, 64)
		if err != nil {
			return c, err
		}
		c.iPoint = append(c.iPoint, point)
	}
	return c, err
}

// *ELEMENT, type=CPS3, ELSET=Surface17
// 1906, 39, 234, 247
func convertStringToCPS3(line string) (c inpElement, err error) {
	c.elType = inpTypeCPS3
	s := strings.Split(line, ",")
	if len(s) != 4 {
		return c, fmt.Errorf("Wrong string line for convert coordinate of point.\nLine = %v.\nSlice=%v", line, s)
	}
	for i := range s {
		s[i] = strings.TrimSpace(s[i])
	}
	c.index, err = strconv.ParseUint(s[0], 10, 32)
	if err != nil {
		return c, err
	}
	for i := 0; i < 3; i++ {
		point, err := strconv.ParseUint(s[1+i], 10, 64)
		if err != nil {
			return c, err
		}
		c.iPoint = append(c.iPoint, point)
	}
	return c, err
}
