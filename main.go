package main

import (
	"crypto/sha512"
	"io/ioutil"
	"log"
	"os"
	"os/user"

	"path/filepath"
	"strings"

	"github.com/syndtr/goleveldb/leveldb"
)

type File struct {
	Path   string
	Digest [64]byte
}

var (
	db    *leveldb.DB
	queue chan *File
)

func ignore(path string) bool {
	return strings.HasSuffix(path, ".git") || strings.HasPrefix(path, ".")
}

func walk(path string, info os.FileInfo, err error) error {
	if err != nil {
		log.Print(err)
		return nil
	}

	if info.IsDir() || ignore(path) {
		return nil
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Print("Read file error: ", err)
		return nil
	}

	digest := sha512.Sum512(data)
	queue <- &File{path, digest}
	return nil
}

func inspect() {
	for file := range queue {
		log.Println("Receive ", file.Path)
		data, err := db.Get(file.Digest[:], nil)
		if err == leveldb.ErrNotFound {
			err = db.Put(file.Digest[:], []byte(file.Path), nil)
		} else {
			log.Println(string(data), "has dup", file.Path)
		}
	}
}

func homedir() string {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	return usr.HomeDir
}

func main() {
	queue = make(chan *File, 50000)

	if len(os.Args) < 1 {
		log.Println("MIssing arguments")
		return
	}

	var err error
	db, err = leveldb.OpenFile(homedir()+"/.dedupe", nil)
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	searchDir := os.Args[1]

	go inspect()
	err = filepath.Walk(searchDir, walk)
	if err != nil {
		log.Fatal(err)
	}
}
