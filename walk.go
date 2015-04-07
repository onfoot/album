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
	"strings"
)

var extensions = []string{".jpg", ".jpeg", ".JPG", ".JPEG"}
var metaDir = ".album"

var root = flag.String("root", "", "Album root")
var metaRoot string

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

			log.Println("Hashing", normalizedPath)

			file, fileErr := os.Open(path)

			if fileErr != nil {
				return fileErr
			}

			sha := sha1.New()
			io.Copy(sha, file)

			file.Close()

			os.MkdirAll(filepath.Dir(metaDataPath), 0755)
			hashFile, hashFileErr = os.Create(metaDataPath)

			if hashFileErr == nil {
				hashFile.Write(sha.Sum(nil))
				hashFile.Close()
			} else {
				log.Println("Could not write hash file: ", hashFileErr)
			}

			if fileErr != nil {
				return fileErr
			}

			break
		}
	}

	return nil
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

	finish := make(chan int)
	go func() {
		walkErr := filepath.Walk(*root, walker)

		if walkErr != nil {
			log.Fatal(walkErr.Error())
		}
		close(finish)
	}()

	<-finish
}
