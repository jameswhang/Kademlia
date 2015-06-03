package kademlia

import (
	"fmt"
	"io/ioutil"
	"bufio"
	"os"
)

func main_in_this() {
	readFile()
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
	path := "./nodes/6aded190011951343d3e6e8f19a550d833a3ac7f.txt"
	// f, err := os.Open("./nodes/77bdbb08c418bab6de1d838da01a60e8632b75a8.txt")
	// handle(err, "")
	line := readFromFile(path)
	fmt.Println(line)
	// f.Close()
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

func readFromFile(filename string) []byte {
	
	b, err := ioutil.ReadFile(filename)
	handle(err, "Reading from file error.")
	fmt.Println(len(b))
	return b
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
