package main

import (
	"fmt"
	// "io/ioutil"
	"os"
	"bufio"
)

func main() {
	createFile()
}

func createFile() {
	f, _ := os.Create("./nodes/write.txt")
	s := "This is Nevil writing to a file using bufio tralalala"
	err := writeStringToFile(f, s)
	if err != nil {
		panic(err)
	}
	f.Close()
}

func check(err error) {
	if err != nil {
		fmt.Println("error occurred")
	}
}

func readFile() {
	f, err := os.Open("./nodes/hello.txt")
	check(err)
	scanner := bufio.NewScanner(f)
	defer func() {
		err := f.Close()
		if err != nil {
			check(err)
		}
	}()

	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}
}

func writeStringToFile(f *os.File, text string) error {
	writer := bufio.NewWriter(f)
	_, err := writer.WriteString(text)
	if err != nil {
		fmt.Println("String incompletely written to file.")
		return err
 	}
 	err = writer.Flush()
 	if err != nil {
 		fmt.Println("Flush error.")
		return err
 	}
 	return nil
}