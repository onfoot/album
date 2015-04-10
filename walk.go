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
var testMode = flag.Bool("test", false, "Test mode")

var metaRoot string

type HashingTask struct {
	path     string
	info     os.FileInfo
	root     string
	metaRoot string
}

type PhotoWalker struct {
	walkFunc func(path string, info os.FileInfo)
}

func (w PhotoWalker) photoWalker() filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {

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
				w.walkFunc(path, info)
				return nil
			}
		}

		return nil
	}
}

func HashPath(path string, root string, metaRoot string) string {
	normalizedPath := strings.TrimPrefix(path, root)
	normalizedPath = strings.TrimPrefix(normalizedPath, "/")

	hashPath := filepath.Join(metaRoot, "hash", normalizedPath) + ".sha1"
	return hashPath
}

var Usage = func() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	flag.PrintDefaults()
}

// var hasher = make(chan HashingTask, runtime.NumCPU())

func main() {
	flag.Parse()

	if *root == "" {
		Usage()
		return
	}

	*root = filepath.Clean(*root)
	metaRoot = filepath.Join(*root, metaDir)

	if _, rootErr := ioutil.ReadDir(*root); rootErr != nil {
		log.Fatal("Root directory could not be read")
	}

	log.Println("Meta dir: " + metaDir)
	log.Println("Root: " + *root)

	var tasksCount int
	var fileCount int

	workerChan := make(chan HashingTask, runtime.NumCPU())
	counterChan := make(chan int)

	finishChan := make(chan int)

	go func() {
		for {
			val, ok := <-counterChan
			if !ok {
				return
			}

			if val > 0 {
				tasksCount++
			}

			if tasksCount == fileCount {
				finishChan <- 1
			}

		}
	}()

	for i := 0; i < runtime.NumCPU(); i++ {
		go func() {
			for {
				task := <-workerChan

				file, fileErr := os.Open(task.path)

				if fileErr != nil {
					log.Println("Could not read photo file at", task.path)
					continue
				}

				sha := sha1.New()
				io.Copy(sha, file)

				file.Close()

				sum := sha.Sum(nil)

				hashPath := HashPath(task.path, task.root, task.metaRoot)

				hashFile, hashFileErr := os.Open(hashPath)

				if hashFileErr == nil {
					currentSum := make([]byte, 20)
					hashFile.Read(currentSum)

					if string(currentSum) == string(sum) {
					}

					hashFile.Close()
					counterChan <- 1
					log.Println("Skipping", task.path)
					continue
				}

				if !(*testMode) {

					os.MkdirAll(filepath.Dir(hashPath), 0755)
					hashFile, hashFileErr := os.Create(hashPath)

					if hashFileErr == nil {
						hashFile.Write(sum)
						hashFile.Close()
					} else {
						log.Println("Could not write hash file: ", hashFileErr)
					}
				}

				log.Println("Task done for", task.path)
				counterChan <- 1
			}
		}()
	}

	photoWalker := PhotoWalker{}
	photoWalker.walkFunc = func(path string, info os.FileInfo) {
		fileCount++
		workerChan <- HashingTask{path, info, *root, metaRoot}
	}

	walkErr := filepath.Walk(*root, photoWalker.photoWalker())

	if walkErr != nil {
		log.Fatal(walkErr.Error())
	}

	<-finishChan
	close(workerChan)
	log.Printf("%d files processed", fileCount)
}
