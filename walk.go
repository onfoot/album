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
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/nfnt/resize"
	"github.com/rwcarlsen/goexif/exif"
)

var extensions = []string{".jpg", ".jpeg", ".JPG", ".JPEG"}
var root = flag.String("root", ".", "Root content directory")
var metaDir = ".album"
var testMode = flag.Bool("test", false, "Test mode")
var httpAddress = flag.String("http", ":8080", "Default listening http address")

func HashPath(metaRoot string, hash string) string {
	hashPath := filepath.Join(metaRoot, "hash", hash) + ".sha1"
	return hashPath
}

func ThumbPath(metaRoot string, hash string) string {
	hashPath := filepath.Join(metaRoot, "thumbs", hash) + ".jpg"
	return hashPath
}

type Info struct {
	Path     string
	FileInfo os.FileInfo
	Hash     []byte
}

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

func Walker() <-chan Info {
	out := make(chan Info)

	go func() {
		filepath.Walk(*root, func(path string, info os.FileInfo, err error) error {

			if info.IsDir() {

				if path == "." {
					return nil
				}

				if strings.HasPrefix(info.Name(), ".") {
					log.Println("Skipping hidden dir", info.Name())
					return filepath.SkipDir
				}

				if strings.HasSuffix(path, ".app") {
					log.Println("Skipping app")
					return filepath.SkipDir
				}

				if strings.HasSuffix(path, ".bundle") {
					log.Println("Skipping bundle")
					return filepath.SkipDir
				}

				if strings.HasSuffix(path, ".photoslibrary") {
					log.Println("Skipping iPhoto library bundle")
					return filepath.SkipDir
				}

				if strings.HasSuffix(path, ".photolibrary") {
					log.Println("Skipping Photos library bundle")
					return filepath.SkipDir
				}

				return nil
			}

			for _, ext := range extensions {
				if ext == filepath.Ext(info.Name()) {
					out <- Info{Path: path, FileInfo: info}
					break
				}
			}

			return nil
		})
		close(out)
	}()

	return out
}

func Hash(info Info) ([]byte, error) {

	file, err := os.Open(info.Path)
	if err != nil {
		return []byte{}, err
	}

	sha := sha1.New()
	io.Copy(sha, file)

	file.Close()

	sum := sha.Sum(nil)

	return sum, nil
}

func Hasher(in <-chan Info) <-chan Info {
	out := make(chan Info)
	done := make(chan bool)

	count := runtime.NumCPU() * 2

	for i := 0; i != count; i++ {
		go func(in <-chan Info) {
			for {
				info, ok := <-in

				if !ok {
					done <- true
					return
				}

				result, err := Hash(info)

				if err != nil {
					log.Fatalf("Hashing error for %v: %v", info.Path, err)
				}

				hashHex := hex.EncodeToString(result)

				hashPath := HashPath(metaRoot, hashHex)

				currentHashHex, err := ioutil.ReadFile(hashPath)

				info.Hash = result

				if err != nil || string(currentHashHex) != hashHex {

					if !(*testMode) {
						os.MkdirAll(filepath.Dir(hashPath), 0755)
						file, err := os.Create(hashPath)

						if err != nil {
							log.Fatalf("Could not create hash file: %v", err)
						} else {
							if _, err := file.WriteString(info.Path); err != nil {
								log.Fatalf("Could not write to hash file %v", err)
							}

							file.Close()
						}
					}
				}

				out <- info
			}

		}(in)
	}

	go func() {
		for i := 0; i != count; i++ {
			<-done
		}
		close(out)
	}()

	return out
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
		return 0, 0
	}

	orientation, err := tag.Int(0)

	if err != nil {
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

func Thumbnail(info Info) (image.Image, error) {

	file, err := os.Open(info.Path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var exifBuffer bytes.Buffer

	img, imgErr := jpeg.Decode(io.TeeReader(file, &exifBuffer))

	if imgErr != nil {
		log.Printf("Jpeg decode error: %v", imgErr)
		return nil, imgErr
	}

	m := resize.Thumbnail(800, 800, img, resize.Bilinear)
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

	return m, nil
}

func Thumbnailer(in <-chan Info) <-chan Info {
	out := make(chan Info)
	done := make(chan bool)

	count := runtime.NumCPU() * 2

	for i := 0; i != count; i++ {
		go func(in <-chan Info) {
			for {
				info, ok := <-in

				if !ok {
					done <- true
					return
				}

				thumbPath := ThumbPath(metaRoot, hex.EncodeToString(info.Hash))
				thumbFile, err := os.Open(thumbPath)

				thumbFile.Close()

				if err != nil {
					thumb, resultErr := Thumbnail(info)

					if resultErr != nil {
						log.Fatalf("Could not create a thumbnail: %v", resultErr)
					}

					if !(*testMode) {
						os.MkdirAll(filepath.Dir(thumbPath), 0755)
						file, err := os.Create(thumbPath)

						if err != nil {
							log.Fatalf("Could not create thumbnail file: %v", err)
						} else {
							jpeg.Encode(file, thumb, &jpeg.Options{Quality: 85})
							file.Close()
						}
					}
				}

				out <- info
			}

		}(in)
	}

	go func() {
		for i := 0; i != count; i++ {
			<-done
		}
		close(out)
	}()

	return out
}

var Usage = func() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	flag.PrintDefaults()
}

var metaRoot string

func main() {
	flag.Parse()

	*root = filepath.Clean(*root)
	metaRoot = filepath.Join(*root, metaDir)

	if _, rootErr := ioutil.ReadDir(*root); rootErr != nil {
		log.Fatal("Root directory could not be read")
	}

	log.Println("Meta dir: " + metaDir)
	log.Println("Root: " + *root)

	walk := Walker()
	hash := Hasher(walk)
	thumb := Thumbnailer(hash)

	photos := []string{}

	for info := range thumb {
		photos = append(photos, hex.EncodeToString(info.Hash))
	}

	log.Printf("Processed %v files", len(photos))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "<!doctype html><html><body>Hi, you have %d photos</body></html>", len(photos))
	})

	log.Println("Listening on", *httpAddress)
	log.Fatal(http.ListenAndServe(*httpAddress, nil))
}
