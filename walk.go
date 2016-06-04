package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/rwcarlsen/goexif/exif"
)

var extensions = []string{".jpg", ".jpeg", ".JPG", ".JPEG"}
var metaDir = ".album"

var root = flag.String("root", "", "Album root")
var testMode = flag.Bool("test", false, "Test mode")
var httpAddress = flag.String("http", ":8080", "Default listening http address")

var metaRoot string

type FlipMode int

const (
	FlipVertical FlipMode = 1 << iota
	FlipHorizontal
)

const (
	topLeftSide     = 1
	topRightSide    = 2
	bottomRightSide = 3
	bottomLeftSide  = 4
	leftSideTop     = 5
	rightSideTop    = 6
	rightSideBottom = 7
	leftSideBottom  = 8
)

type PhotoTask struct {
	path     string
	root     string
	metaRoot string
	hash     string
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

			if strings.HasSuffix(path, ".app") {
				return filepath.SkipDir
			}

			if strings.HasSuffix(path, ".bundle") {
				return filepath.SkipDir
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

func ThumbPath(hash string, root string, metaRoot string) string {
	hashPath := filepath.Join(metaRoot, "thumbs", hash) + ".jpg"
	return hashPath
}

var Usage = func() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	flag.PrintDefaults()
}

func ExifOrientation(r io.Reader) (int, FlipMode) {
	var (
		angle int
		flip  FlipMode
	)

	ex, err := exif.Decode(r)
	if err != nil {
		log.Println("Could not decode Exif information")
		return 0, 0
	}

	tag, err := ex.Get(exif.Orientation)
	if err != nil {
		log.Println("No EXIF orientation tag: ", err)
		return 0, 0
	}

	orientation, err := tag.Int(0)

	if err != nil {
		log.Printf("EXIF error %v", err)
		return 0, 0
	}

	switch orientation {
	case topLeftSide:
		// no change required
	case topRightSide:
		flip = FlipHorizontal
	case bottomRightSide:
		angle = 180
	case bottomLeftSide:
		angle = 180
		flip = FlipHorizontal
	case leftSideTop:
		angle = -90
		flip = FlipHorizontal
	case rightSideTop:
		angle = -90
	case rightSideBottom:
		angle = 90
		flip = FlipHorizontal
	case leftSideBottom:
		angle = 90
	}

	return angle, flip
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

	photos := []string{}
	hashes := make(map[string]string)

	photoWalker := PhotoWalker{}
	photoWalker.walkFunc = func(path string, info os.FileInfo) {
		photos = append(photos, path)
	}

	walkErr := filepath.Walk(*root, photoWalker.photoWalker())
	if walkErr != nil {
		log.Fatal(walkErr.Error())
	}

	fileCount := len(photos)
	counting := make(chan int)

	ticking := make(chan int, workerCount)
	finishing := make(chan bool)

	hashWorker := make(chan PhotoTask, workerCount)

	go func(work chan PhotoTask) {
		for task := range work {
			ticking <- 1

			task := task

			go func() {
				log.Println("Starting work on", task.path)
				file, fileErr := os.Open(task.path)

				if fileErr != nil {
					log.Println("Could not read photo file at", task.path, fileErr)
					<-ticking
					counting <- 1

					return
				}

				sha := sha1.New()
				io.Copy(sha, file)

				file.Close()

				sum := sha.Sum(nil)

				hashPath := HashPath(task.path, task.root, task.metaRoot)

				sumHex := hex.EncodeToString(sum)

				hashes[task.path] = sumHex

				if currentSumHex, hashErr := ioutil.ReadFile(hashPath); hashErr == nil {
					if string(currentSumHex) == sumHex {
						log.Println("Skipping", task.path)

						<-ticking
						counting <- 1

						return
					}
				}

				if !(*testMode) {
					os.MkdirAll(filepath.Dir(hashPath), 0755)
					file, fileErr := os.Create(hashPath)

					if fileErr != nil {
						log.Println("Could not create hash file: ", fileErr)
					} else {
						defer file.Close()

						if _, hashErr := file.WriteString(sumHex); hashErr != nil {
							log.Println("Could not write hash file: ", hashErr)
						}
					}
				}

				<-ticking
				counting <- 1

			}()

		}
	}(hashWorker)

	go func() {
		for _, path := range photos {
			task := PhotoTask{path: path, root: *root, metaRoot: metaRoot}
			hashWorker <- task
		}
	}()

	go func() {
		taskCount := 0

		for count := range counting {
			taskCount += count

			if taskCount == fileCount {
				close(counting)
				finishing <- true
				return
			}

		}
	}()

	<-finishing
	close(ticking)

	counting = make(chan int)

	ticking = make(chan int, workerCount)
	finishing = make(chan bool)

	photoWorker := make(chan PhotoTask, workerCount)

	go func(work chan PhotoTask) {
		for task := range work {
			ticking <- 1

			task := task

			go func() {
				log.Printf("Starting thumbnail work on %s (%s)", task.path, task.hash)
				file, fileErr := os.Open(task.path)

				if fileErr != nil {
					log.Println("Could not read photo file at", task.path, fileErr)
					<-ticking
					counting <- 1

					return
				}

				var exifBuffer bytes.Buffer

				img, imgErr := jpeg.Decode(io.TeeReader(file, &exifBuffer))

				file.Close()

				if imgErr != nil {
					log.Printf("Jpeg decode error: %v", imgErr)
					<-ticking
					counting <- 1
					return
				}

				maxDimension := 800.0
				width := float64(img.Bounds().Max.X - img.Bounds().Min.X)
				height := float64(img.Bounds().Max.Y - img.Bounds().Min.Y)

				var dstWidth float64
				var dstHeight float64

				biggerDimension := math.Max(width, height)

				if width < maxDimension && height < maxDimension {
					dstWidth = width
					dstHeight = height
				} else {
					scaleFactor := biggerDimension / maxDimension

					dstWidth = width / scaleFactor
					dstHeight = height / scaleFactor
				}

				idstWidth := int(math.Ceil(dstWidth))
				idstHeight := int(math.Ceil(dstHeight))

				var m image.Image
				if dstWidth == width && dstHeight == height {
					m = img
				} else {
					m = imaging.Resize(img, idstWidth, idstHeight, imaging.Box)
				}

				angle, flip := ExifOrientation(&exifBuffer)

				switch angle {
				case 90:
					m = imaging.Rotate90(m)
				case -90:
					m = imaging.Rotate270(m)
				case 180:
					m = imaging.Rotate180(m)
				}

				switch flip {
				case FlipHorizontal:
					m = imaging.FlipH(m)
				case FlipVertical:
					m = imaging.FlipV(m)
				}

				thumbPath := ThumbPath(task.hash, task.root, task.metaRoot)

				if !(*testMode) {
					os.MkdirAll(filepath.Dir(thumbPath), 0755)

					thumbFile, thumbErr := os.Create(thumbPath)
					if thumbErr == nil {

						jpeg.Encode(thumbFile, m, &jpeg.Options{Quality: 85})

						thumbFile.Close()
					}
				}

				log.Printf("Processed image, %0.fx%0.f, thumbnail is %0.fx%0.f", width, height, dstWidth, dstHeight)
				<-ticking

				counting <- 1
			}()

		}
	}(photoWorker)

	go func() {
		for _, path := range photos {
			task := PhotoTask{path: path, root: *root, metaRoot: metaRoot, hash: hashes[path]}
			photoWorker <- task
		}
	}()

	go func() {
		taskCount := 0

		for count := range counting {
			taskCount += count

			if taskCount == fileCount {
				close(counting)
				finishing <- true
				return
			}

		}
	}()

	<-finishing
	close(ticking)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "<!doctype html><html><body>Hi, you have %d photos</body></html>", len(photos))
	})

	log.Println("Listening on", *httpAddress)
	log.Fatal(http.ListenAndServe(*httpAddress, nil))
}
