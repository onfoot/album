package main

import (
	"crypto/sha1"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

var extensions = []string{".jpg", ".jpeg", ".JPG", ".JPEG"}
var metaDir = ".album"

var root = flag.String("root", "", "Album root")
var metaRoot string

type HashingTask struct {
	MetaDataPath string
	FilePath     string
}

func walker(path string, info os.FileInfo, err error) error {

	if info.IsDir() {

		if path == "." {
			return nil
		}

		if path == ".." {
			return nil
		}

		if strings.HasPrefix(info.Name(), ".") {
			return filepath.SkipDir
		}

		return nil
	}

	for _, ext := range extensions {
		if ext == filepath.Ext(info.Name()) {
			normalizedPath := strings.TrimPrefix(path, *root)
			normalizedPath = strings.TrimPrefix(normalizedPath, "/")

			metaDataPath := filepath.Join(metaRoot, "hash", normalizedPath) + ".sha1"

			hashFile, hashFileErr := os.Open(metaDataPath)

			if hashFileErr == nil {
				hashFile.Close()
				return nil
			}

			hasher <- HashingTask{metaDataPath, path}
			// hasher

			break
		}
	}

	return nil
}

var Usage = func() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	flag.PrintDefaults()
}

var hasher = make(chan HashingTask, runtime.NumCPU())

func main() {
	flag.Parse()

	if *root == "" {
		Usage()
		return
	}

	log.Printf("Setting up for %d CPUs", runtime.NumCPU())

	*root = filepath.Clean(*root)
	metaRoot = filepath.Join(*root, metaDir)

	if _, rootErr := ioutil.ReadDir(*root); rootErr != nil {
		log.Fatal("Root directory could not be read")
	}

	log.Println("Meta dir: " + metaDir)
	log.Println("Root: " + *root)

	go func() {

		for task := range hasher {

			go func(task HashingTask) {
				log.Println("Task starting for", task.FilePath)
				file, fileErr := os.Open(task.FilePath)

				if fileErr != nil {
					return
				}

				sha := sha1.New()
				io.Copy(sha, file)

				file.Close()

				// sum := sha.Sum(nil)

				/*os.MkdirAll(filepath.Dir(task.MetaDataPath), 0755)
				hashFile, hashFileErr := os.Create(task.MetaDataPath)

				if hashFileErr == nil {
					hashFile.Write(sum)
					hashFile.Close()
				} else {
					log.Println("Could not write hash file: ", hashFileErr)
				}*/

				log.Println("Task done for", task.FilePath)
			}(task)
		}

	}()

	walkErr := filepath.Walk(*root, walker)

	if walkErr != nil {
		log.Fatal(walkErr.Error())
	}
}
