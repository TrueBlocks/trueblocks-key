package main

import (
	"io"
	"os"
	"path"
	"strings"
)

type CmdRunEnv struct {
	IndexPath string
}

func (c *CmdRunEnv) Blooms(chain string) (map[string]string, error) {
	bloomPath := path.Join(c.IndexPath, "blooms/")
	files, err := os.ReadDir(bloomPath)
	if err != nil {
		return nil, err
	}

	result := make(map[string]string, len(files))

	for _, info := range files {
		if info.IsDir() {
			continue
		}

		rawFileName := info.Name()
		fileName := bloomPath + "/" + rawFileName

		if !strings.Contains(fileName, ".bloom") {
			continue // sometimes there are .gz files in this folder, for example
		}

		result[rawFileName] = fileName
	}

	return result, nil
}

func (c *CmdRunEnv) ReadBloom(fileName string) (io.ReadSeekCloser, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (c *CmdRunEnv) ReadChunk(chain string, blockRange string) (io.ReadSeekCloser, error) {
	indexFilename := path.Join(c.IndexPath, "finalized/", blockRange+".bin")
	// indexFilename := config.GetPathToIndex(chain) + "finalized/" + index.ToIndexPath(blockRange)
	f, err := os.Open(indexFilename)
	if err != nil {
		return nil, err
	}

	return f, nil
}
