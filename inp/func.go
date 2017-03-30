package inp

func (f Format) AddUniqueIndexToElements() {
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
	for _, element := range f.Elements {
		for _, data := range element.Data {
			if data.Index == -1 {
				maxIndexElement++
				data.Index = maxIndexElement
			}
		}
	}
}
