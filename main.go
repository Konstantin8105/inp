package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/Konstantin8105/Convert-INP-to-STD-format/convertorInp"
)

// Convert all INP files to STD files
func main() {
	// menu
	line()
	fmt.Println("|==================================|")
	fmt.Println("| Software:                        |")
	fmt.Println("| Convert file from format INP to  |")
	fmt.Println("| format STD.                      |")
	fmt.Println("|                                  |")
	fmt.Println("| Created by : Konstantin.I        |")
	fmt.Println("| License : MIT. Absolute free to  |")
	fmt.Println("| use, copy, modify.               |")
	fmt.Println("|==================================|")
	line()

	files, err := getFilesInp()
	if err != nil {
		fmt.Println(err)
		return
	}
	for i, file := range files {
		fmt.Printf("| %3v. %v\n", i, file)
	}
	line()

	reader := bufio.NewReader(os.Stdin)
	text, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println(err)
		return
	}
	text = strings.Replace(text, "\n", "", -1)
	text = strings.TrimSpace(text)

	if len(text) == 0 {
		fmt.Println("You don`t enter the position of file")
		return
	}

	index, err := strconv.ParseUint(text, 10, 64)
	if err != nil {
		fmt.Println(err)
		return
	}

	if index >= 0 && index < uint64(len(files)) {
		fmt.Println("Convert file :", files[index])
		err := convertorInp.Convert(files[index], files[index]+".std")
		if err != nil {
			fmt.Println(err)
		}
	}
}

func line() {
	fmt.Println("|==================================|")
}

func getFilesInp() (files []string, err error) {
	fileList, err := ioutil.ReadDir("./")
	if err != nil {
		return
	}
	for _, file := range fileList {
		var del bool
		if file.IsDir() {
			del = true
		}
		if !strings.HasSuffix(file.Name(), ".inp") {
			del = true
		}
		if !del {
			files = append(files, file.Name())
		}
	}
	return files, err
}
