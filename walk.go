package main

import (
	"crypto/sha1"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

var extensions = []string{".jpg", ".jpeg", ".JPG", ".JPEG"}
var metaDir = ".album"

var root = flag.String("root", "", "Album root")
var testMode = flag.Bool("test", false, "Test mode")
var httpAddress = flag.String("http", ":8080", "Default listening http address")

var metaRoot string

type HashingTask struct {
	path     string
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

	workerCount := runtime.NumCPU()
	worker := make(chan HashingTask, workerCount)

	counting := make(chan int)

	photos := []string{}

	photoWalker := PhotoWalker{}
	photoWalker.walkFunc = func(path string, info os.FileInfo) {
		photos = append(photos, path)
	}

	walkErr := filepath.Walk(*root, photoWalker.photoWalker())
	if walkErr != nil {
		log.Fatal(walkErr.Error())
	}

	fileCount := len(photos)
	taskCount := 0

	finishing := make(chan int)

	for i := 0; i != workerCount; i++ {
		go func(i int, in chan HashingTask) {
			for {
				task, ok := <-in

				if !ok {
					break
				}

				file, fileErr := os.Open(task.path)

				if fileErr != nil {
					log.Println("Could not read photo file at", task.path, fileErr)
					counting <- 1
					break
				}

				sha := sha1.New()
				io.Copy(sha, file)

				file.Close()

				sum := sha.Sum(nil)

				hashPath := HashPath(task.path, task.root, task.metaRoot)

				if currentSum, hashErr := ioutil.ReadFile(hashPath); hashErr == nil {
					if string(currentSum) == string(sum) {
						log.Println("Skipping", task.path)
						counting <- 1
						break
					}

				}

				if !(*testMode) {
					os.MkdirAll(filepath.Dir(hashPath), 0755)
					if hashErr := ioutil.WriteFile(hashPath, sum, 0666); hashErr != nil {
						log.Println("Could not write hash file: ", hashErr)
					}
				}

				log.Println("processed", task.path)
				counting <- 1
			}
		}(i, worker)
	}

	go func() {
		for _, path := range photos {
			task := HashingTask{path, *root, metaRoot}
			worker <- task
		}
	}()

	go func() {
		for {
			count, ok := <-counting
			if !ok {
				break
			}

			taskCount += count

			if taskCount == fileCount {
				close(finishing)
				break
			}
		}
	}()

	<-finishing
	close(worker)
	close(counting)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "<!doctype html><html><body>Hi, you have %d photos</body></html>", len(photos))
	})

	log.Println("Listening on", *httpAddress)
	log.Fatal(http.ListenAndServe(*httpAddress, nil))
}
