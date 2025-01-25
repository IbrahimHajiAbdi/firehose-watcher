package utils

import (
	"fmt"
	"os"
)

type FileSystem interface {
	OpenFile(name string, flag int, perm os.FileMode) (File, error)
}

type File interface {
	Write(data []byte) (int, error)
	Close() error
}

type DefaultFileSystem struct{}

func (dfs *DefaultFileSystem) OpenFile(name string, flag int, perm os.FileMode) (File, error) {
	return os.OpenFile(name, flag, perm)
}

type DefaultFile struct {
	file *os.File
}

func (df *DefaultFile) Write(data []byte) (int, error) {
	return df.file.Write(data)
}

func (df *DefaultFile) Close() error {
	return df.file.Close()
}

func WriteFile(fs FileSystem, filepath string, data *[]byte) error {
	f, err := fs.OpenFile(filepath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer f.Close()

	_, err = f.Write(*data)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}
