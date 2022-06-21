package api

import (
	"fmt"
	"io/ioutil"
)

func GetImage(path string) []byte {
	availableExts := []string{"", ".png", ".jpg", ".jpeg"}
	var bytes []byte = nil
	for _, ext := range availableExts {
		p := fmt.Sprintf("%s%s", path, ext)
		// println("getting file: " + p)
		fileBytes, err := ioutil.ReadFile(p)
		if err != nil {
			// println("not found")
			continue
		}
		// println("found!")
		bytes = fileBytes
		break
	}
	return bytes
}

func GetImageWithDefault(path string, defaultPath string) []byte {
	bytes := GetImage(path)
	if bytes == nil {
		b, err := ioutil.ReadFile(defaultPath)
		if err != nil {
			println("Default file not found")
			return nil
		}
		bytes = b
	}
	return bytes
}
