package convertorInp

import (
	"fmt"
	"os"
)

func createNewFile(filename string, lines []string) (err error) {
	if len(filename) == 0 {
		return fmt.Errorf("Filename is zero: %v", filename)
	}
	if _, err := os.Stat(filename); !os.IsNotExist(err) {
		return fmt.Errorf("File %v is exist. Please change the name for saving data", filename)
	}
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("Cannot create the file: %v.\nError: %v", filename, err)
	}

	defer func() {
		errFile := file.Close()
		if errFile != nil {
			if err != nil {
				err = fmt.Errorf("%v ; %v", err, errFile)
			} else {
				err = errFile
			}
		}
	}()

	for _, line := range lines {
		fmt.Fprintf(file, "%s\n", line)
	}

	return nil
}
