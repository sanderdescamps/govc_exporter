package main

import (
	"errors"
	"fmt"
	"os"
	"time"
)

func createPProfFile(name string) (*os.File, error) {
	dirPath := fmt.Sprintf("./dumps/%s", time.Now().Format("20060201"))
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		err = os.MkdirAll(dirPath, 0775)
		if err != nil {
			return nil, err
		}
	}
	filePath := fmt.Sprintf("%s/%s.pprof", dirPath, name)
	for i := 0; true; i++ {
		if _, err := os.Stat(filePath); errors.Is(err, os.ErrNotExist) {
			break
		}
		filePath = fmt.Sprintf("%s/%s-%d.pprof", dirPath, name, i)
		fmt.Printf("pprof path: %s", filePath)
	}

	file, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}
	return file, nil
}
