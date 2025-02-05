//go:build unix

package utils

import (
	"fmt"
	"path"
	"path/filepath"
	"strconv"
	"unicode/utf8"
)

func MakeFilepath(directory string, rkey string, handle string, text string, mediaType string, i int, maxBytes int) string {
	filename := fmt.Sprintf("%s_%s_%s", rkey, handle, text)
	var filePath string
	if i > 0 {
		filename = FilenameLengthLimit(filename, maxBytes-(len(mediaType)+2+i))
		filePath = filepath.Join(directory, filename, strconv.Itoa(i))
		filePath += "." + mediaType
	} else {
		filename = FilenameLengthLimit(filename, maxBytes-(len(mediaType)+1))
		filePath = filepath.Join(directory, filename)
		filePath += "." + mediaType
	}
	return path.Clean(filePath)
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
