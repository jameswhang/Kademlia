package main

import (
	"fmt"
	// "io/ioutil"
	"os"
	"bufio"
)

func main() {
	writeFile()
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

func handle(e error, msg string) {
	if e != nil {
		fmt.Println(msg)
		panic(e)
	}
}

func readFile() {
	f, err := os.Open("./nodes/write.txt")
	handle(err, "")
	lines := readLinesFromFile(f)
	for _, l := range lines {
		fmt.Println(l)
	}
	f.Close()
}

func writeFile() {
	path := "./nodes/write.txt"
	f, err := os.OpenFile(path, os.O_RDWR|os.O_APPEND, 0660)
	handle(err, "Cannot open file to append.")
	text := "\nLets get one more in there"
	appendToFile(f, text)
	f.Close()
}

func readLinesFromFile(f *os.File) []string {
	scanner := bufio.NewScanner(f)
	list := make([]string, 0)
	for scanner.Scan() {
		list = append(list, scanner.Text())
	}
	return list
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

func appendToFile(f *os.File, s string) {
	writer := bufio.NewWriter(f)
	_, err := writer.WriteString(s)
	handle(err, "Error writing appending to file")
	err = writer.Flush()
	handle(err, "Error flushing after append.")
}
