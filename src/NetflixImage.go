package main

import (
	"bufio"
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
)

// get largest image if same name occurs more, save prev pixel value
var nameSize = make(map[string]int)
var saveFolder string

// Main Controller for instagram image download
func main() {

	argsWithProg := os.Args
	if len(argsWithProg) > 1 {
		saveFolder = argsWithProg[1]
		saveFolder = strings.Replace(saveFolder, "\\", "/", -1)
		if !strings.HasSuffix(saveFolder, "/") {
			saveFolder += "/"
		}
	} else {
		saveFolder = "c:/tmp/"
	}
	_ = os.MkdirAll(saveFolder, os.ModePerm)

	for {

		webSource := readConsole()
		nameSize = make(map[string]int)

		// remove escape letter
		webSource = strings.Replace(webSource, `\/`, `/`, -1)

		// image/video URL matching
		imgUrlEx := regexp.MustCompile(`https://scontent.+?instagram.com[" ]`)
		imgUrls := imgUrlEx.FindAllString(webSource, -1)

		// save image
		for ix, u := range imgUrls {
			u = u[:len(u)-1]
			name := getName(u)
			imageBytes, xSize, ySize := getImage(name, u)
			pixels := xSize * ySize
			if pixels <= 350*350 {
				continue
			}
			prevPixels := nameSize[name]
			if prevPixels < pixels {
				_ = ioutil.WriteFile(saveFolder+name, imageBytes, 0644)
				nameSize[name] = pixels
				fmt.Println(ix, xSize, ySize, name)
			}
		}
	}
}

// extract name from url
func getName(u string) (name string) {
	iQuestion := strings.Index(u, "?")

	for islash := iQuestion; islash > 0; islash-- {
		if u[islash] == '/' {
			name = u[islash+1 : iQuestion]
			break
		}
	}

	return name
}

// download image and return image bytes, width, height
func getImage(name string, url string) ([]byte, int, int) {
	response, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
		return nil, 0, 0
	}
	bodyBytes, _ := ioutil.ReadAll(response.Body)
	_ = response.Body.Close()

	// bypass image whose size is less than (500,y) or (x,500)
	ext := name[strings.LastIndex(name, ".")+1:]
	imageWidth := 9999
	imageHeight := 9999
	if ext != "mp4" {
		getImage, _, err := image.DecodeConfig(bytes.NewReader(bodyBytes))
		if err == nil {
			imageWidth = getImage.Width
			imageHeight = getImage.Height
		}
	}

	return bodyBytes, imageWidth, imageHeight
}

// read html code from console
func readConsole() string {
	fmt.Println("Save folder is", saveFolder, "\n",
		"Enter html source, at end type enter. ~0 exits, ~1 get images.")
	text := ""
	reader := bufio.NewReader(os.Stdin)
	for {
		oneLine, _ := reader.ReadString('\n')
		oneLine = strings.Replace(oneLine, "\n", "", -1)
		switch {
		case strings.HasPrefix(oneLine, "~0"):
			// terminate program
			os.Exit(0)
		case strings.HasPrefix(oneLine, "~1"):
			// process and retry
			return text
		default:
			text += oneLine
		}
	}
}
