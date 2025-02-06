//go:build windows

package utils

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

var replacer = strings.NewReplacer(
	"\"", "",
	"\\", "",
	"/", "",
	"|", "",
	":", "",
	"<", "",
	">", "",
	"?", "",
	"*", "",
	"\n", " ",
)

func MakeFilepath(directory string, rkey string, handle string, text string, mediaType string, i int, maxBytes int) string {
	filename := fmt.Sprintf("%s_%s_%s", rkey, handle, text)
	filename = replacer.Replace(filename)
	var filePath string
	if i > 0 {
		filename = FilenameLengthLimit(filename, maxBytes-(len(mediaType)+2+i))
		filePath = filepath.Join(directory, filename)
		filePath = fmt.Sprintf("%s_%d.%s", filePath, i, mediaType)
	} else {
		filename = FilenameLengthLimit(filename, maxBytes-(len(mediaType)+1))
		filePath = filepath.Join(directory, filename)
		filePath = fmt.Sprintf("%s.%s", filePath, mediaType)
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
