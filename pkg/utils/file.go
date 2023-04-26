package utils

import (
	"os"
	"path/filepath"
)

func ReadFiles(source string, extension string, recursive bool) ([]string, error) {
	filePaths := make([]string, 0)
	if recursive {
		err := filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if filepath.Ext(info.Name()) == extension {
				filePaths = append(filePaths, path)
			}
			return nil
		})
		if err != nil {
			return filePaths, err
		}
	} else {
		if filepath.Ext(source) == extension {
			filePaths = append(filePaths, source)
		} else {
			files, err := os.ReadDir(source)
			if err != nil {
				return filePaths, err
			}
			for _, info := range files {
				if filepath.Ext(info.Name()) == extension {
					path := filepath.Join(source, info.Name())
					filePaths = append(filePaths, path)
				}
			}
		}
	}
	return filePaths, nil
}
