package calculixResult

import (
	"fmt"
	"strconv"
	"strings"
)

// SupportForce - struct for saving information in support force
type SupportForce struct {
	Time     float64
	NodeName string
	Forces   []Force
}

// Force - force
type Force struct {
	NodeIndex int
	Load      [3]float64
}

// SupportForces - return forces on support
// Examples in dat file:
// forces (fx,fy,fz) for set FIX and time  0.4000000E-01
// forces (fx,fy,fz) for set LOAD and time  0.2000000E-01
// 204  3.485854E+00  1.025290E+01  3.092803E+01
func SupportForces(datLines []string) (supportForces []SupportForce, err error) {
	headerPrefix := "forces (fx,fy,fz) for set"
	headerMiddle := "and time"

	type stage int
	const (
		undefined stage = iota
		header
		load
	)

	present := undefined
	for _, line := range datLines {
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			if present == load {
				present = undefined
			}
			continue
		}
		if present != header && strings.HasPrefix(line, headerPrefix) {
			present = header
			var support SupportForce
			line = line[len(headerPrefix):]
			s := strings.Split(line, headerMiddle)
			support.NodeName = strings.TrimSpace(s[0])
			time, err := strconv.ParseFloat(strings.TrimSpace(s[1]), 64)
			if err != nil {
				return supportForces, fmt.Errorf("line = %v\nerr=%v", line, err)
			}
			support.Time = time
			supportForces = append(supportForces, support)
			continue
		}
		if present == header || present == load {
			present = load
			f, err := parseForce(line)
			if err != nil {
				return supportForces, err
			}
			supportForces[len(supportForces)-1].Forces = append(supportForces[len(supportForces)-1].Forces, f)
		}
	}
	return supportForces, nil
}

func parseForce(line string) (force Force, err error) {
	s := strings.Split(line, " ")
	for i := range s {
		s[i] = strings.TrimSpace(s[i])
	}

	var index int

	for index = 0; index < len(s); index++ {
		if len(s[index]) == 0 {
			continue
		}
		i, err := strconv.ParseInt(s[index], 10, 64)
		if err != nil {
			return force, fmt.Errorf("Error: string parts - %v, error - %v, in line - %v", s, err, line)
		}
		force.NodeIndex = int(i)
		break
	}

	for position := 0; position < 3; position++ {
		for index++; index < len(s); index++ {
			if len(s[index]) == 0 {
				continue
			}
			factor, err := strconv.ParseFloat(s[index], 64)
			if err != nil {
				return force, fmt.Errorf("Error: string parts - %v, error - %v, in line - %v", s, err, line)
			}
			force.Load[position] = factor
			break
		}
	}

	return force, nil
}
