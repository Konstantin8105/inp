package calculixResult

// SupportForceSummary - type of summary force
type SupportForceSummary struct {
	Time     float64
	NodeName string
	Load     [3]float64
}

// SupportForcesSummary - return summary force on support
func SupportForcesSummary(datLines []string) (summaryForce []SupportForceSummary, err error) {
	s, err := SupportForces(datLines)
	if err != nil {
		return nil, err
	}
	for _, force := range s {
		var summ SupportForceSummary
		summ.Time = force.Time
		summ.NodeName = force.NodeName
		for _, f := range force.Forces {
			for i := 0; i < 3; i++ {
				summ.Load[i] += f.Load[i]
			}
		}
		summaryForce = append(summaryForce, summ)
	}
	return summaryForce, nil
}
