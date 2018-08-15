package watcher

import (
	"bytes"
	"io/ioutil"
	"math/big"
	"os"
	"sync"

	logger "github.com/alecthomas/log4go"
	"errors"
)

var digNoRWMutex sync.RWMutex

func WriteCheckpointBlockNumberToFile(filePath string, blkNumber *big.Int) error {
	digNoRWMutex.Lock()
	defer digNoRWMutex.Unlock()
	return ioutil.WriteFile(filePath, []byte(blkNumber.String()), 0755)
}

func ReadBlockNumberFromFile(filePath string) (*big.Int, error) {
	digNoRWMutex.Lock()
	defer digNoRWMutex.Unlock()
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		logger.Debug("file not found, %v", err)
		return big.NewInt(0), nil
	}

	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	data = bytes.TrimSpace(data)
	delta, isOk := big.NewInt(0).SetString(string(data), 10)
	if isOk == false {
		return big.NewInt(0), errors.New("data is nil")
	}


	return delta, nil
}
