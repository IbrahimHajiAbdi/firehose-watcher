package utils

import (
	"fmt"
	"path"
	"strings"
	"unicode/utf8"
)

func MakeFilepath(directory string, rkey string, handle string, text string, mediaType string, i int, maxBytes int) string {
	filename := fmt.Sprintf("%s_%s_%s", rkey, handle, text)
	filename = strings.Replace(filename, "/", "", -1)
	var filepath string
	if i > 0 {
		filename = FilenameLengthLimit(filename, maxBytes-(len(mediaType)+2+i))
		filepath = fmt.Sprintf("%s/%s_%d.%s", directory, filename, i, mediaType)
	} else {
		filename = FilenameLengthLimit(filename, maxBytes-(len(mediaType)+1))
		filepath = fmt.Sprintf("%s/%s.%s", directory, filename, mediaType)
	}
	return path.Clean(filepath)
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
