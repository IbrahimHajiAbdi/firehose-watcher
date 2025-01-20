package utils

import (
	"fmt"
	"os"
)

func WriteFile(filepath string, data *[]byte) error {
	f, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println(err)
		return err
	}

	_, err = f.Write(*data)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}
