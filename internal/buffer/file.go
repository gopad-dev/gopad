package buffer

import (
	"fmt"
	"os"
)

func readFile(name string) (*os.File, error) {
	file, err := os.Open(name)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return file, nil
}

func writeFile(name string) (*os.File, error) {
	file, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return nil, fmt.Errorf("error creating file: %w", err)
	}

	return file, nil
}

func deleteFile(name string) error {
	return os.Remove(name)
}

func renameFile(oldName, newName string) error {
	return os.Rename(oldName, newName)
}
