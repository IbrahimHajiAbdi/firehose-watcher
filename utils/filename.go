package utils

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

func MakeFilepath(directory string, rkey string, handle string, text string, mediaType string, maxBytes int) string {
	filename := fmt.Sprintf("%s_%s_%s.%s", rkey, handle, text, mediaType)
	filename = strings.Replace(filename, "/", "", -1)
	filepath := fmt.Sprintf("%s/%s", directory, filename)
	return FilenameLengthLimit(filepath, maxBytes)
}

func FilenameLengthLimit(filename string, maxBytes int) string {
	filenameBytes := []byte(filename)
	if len(filenameBytes) < maxBytes {
		return filename
	}
	var lastValidFilename []byte
	var currentFilename []byte
	for i := 0; i < maxBytes; i++ {
		currentFilename = append(currentFilename, filenameBytes[i])
		if utf8.Valid(currentFilename) {
			lastValidFilename = currentFilename
		}
	}
	return string(lastValidFilename)
}
