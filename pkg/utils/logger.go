package utils

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

func SetupLogger(f *os.File) {
	logger := slog.New(slog.NewJSONHandler(f, nil))
	slog.SetDefault(logger)
}

func MakeLogFile(dir string) (*os.File, error) {
	f, err := os.OpenFile(filepath.Join(dir, makeLogFilename()), os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return f, nil
}

func makeLogFilename() string {
	now := time.Now()
	formattedTime := now.Format("02012006_150405")
	return fmt.Sprintf("%s_%s.log", "fw", formattedTime)
}
