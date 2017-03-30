package inp

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
