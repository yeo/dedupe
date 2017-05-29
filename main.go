package main

// Find file with a certain extension, and keep the oldest version since that maybe
// the original one
//
// Usage:
//
// To run in dry mode
// dedupe targetdirectory -extension=,list,by,comma
// When ready:
// dedupe targetdirectory -extension=,list,by,comma -move-to=dir
// Dedupe don't really delete file, it's move file into a folder in your home directory call ~/.dedupe/tmp/
import (
	"crypto/sha512"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"os/user"

	"path/filepath"
	"strings"

	"github.com/syndtr/goleveldb/leveldb"
)

const (
	maxQueuSize = 50000
)

type File struct {
	Path   string
	Digest [64]byte
}

var (
	searchDir string
	db        *leveldb.DB
	queue     chan *File
	done      chan bool

	moveTo    string
	dry       bool
	extension []string
)

func ignore(path string) bool {
	if strings.HasSuffix(path, ".git") || strings.HasPrefix(path, ".") {
		return true
	}
	for _, t := range extension {
		if strings.HasSuffix(path, t) {
			return false
		}
	}

	return true
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
		data, err := db.Get(file.Digest[:], nil)
		if err == leveldb.ErrNotFound {
			err = db.Put(file.Digest[:], []byte(file.Path), nil)
		} else {
			if string(data) != file.Path {
				clean(data, file)
			}
		}
	}
	done <- true
}

func clean(source []byte, dupe *File) {
	if dry == true {
		log.Println(string(source), "has dup", dupe.Path)
	} else {
		//os.Rename(dupe.Path, moveTo+"/"+dupe.Path)
	}
}

func homedir() string {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	return usr.HomeDir
}

func setup() {
	searchDir = os.Args[1]
	queue = make(chan *File, maxQueuSize)
	done = make(chan bool)

	if len(os.Args) < 1 {
		log.Fatal("MIssing arguments")
		return
	}

	dry = true
	m := flag.String("move-to", "", "direcory to keep dup file(for backup)")
	moveTo = *m
	fileTypes := flag.String("extension", "jpg,png,gif", "file type")

	if moveTo != "" {
		dry = false
		log.Println("!!! ACTUALLY DELETE FILE NOW !!!")
	}

	extension = strings.Split(*fileTypes, ",")

	var err error
	db, err = leveldb.OpenFile(homedir()+"/.dedupe/work", nil)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("============================")
	log.Println("Dedupe")
	log.Println("  -> dry mode", dry)
	log.Println("  -> searchDir", searchDir)
	log.Println("  -> extension", extension)
	log.Println("   -> moveTo", moveTo)
	log.Println("============================\n\n")

	go inspect()
}

func teardown() {
	db.Close()
}

func run() {
	if err := filepath.Walk(searchDir, walk); err != nil {
		log.Fatal(err)
	}
	close(queue)
}

func main() {
	setup()
	run()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	select {
	case <-done:
		teardown()
	case <-c:
		log.Println("Force shutdown")
		teardown()
		os.Exit(1)
	}
}
