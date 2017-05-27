package main

import (
	"crypto/sha512"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

func printFile(path string, info os.FileInfo, err error) error {
	if err != nil {
		log.Print(err)
		return nil
	}

	if info.IsDir() {
		return nil
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Print(err)
		return nil
	}

	digest := sha512.Sum512(data)
	log.Println(path, digest)

	fmt.Println(path)
	return nil
}

func main() {
	if len(os.Args) < 1 {
		log.Println("MIssing arguments")
		return
	}

	searchDir := os.Args[1]

	err := filepath.Walk(searchDir, printFile)

	if err != nil {
		log.Fatal(err)
	}
}
