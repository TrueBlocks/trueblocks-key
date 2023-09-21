package main

import (
	"io"
	"os"

	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/config"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/index"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/walk"
)

type CmdRunEnv struct{}

func (c *CmdRunEnv) Blooms(chain string) (map[string]string, error) {
	bloomPath := config.GetPathToIndex(chain) + "blooms/"
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

		if !walk.IsCacheType(fileName, walk.Index_Bloom, true /* checkExt */) {
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
	indexFilename := config.GetPathToIndex(chain) + "finalized/" + index.ToIndexPath(blockRange)
	f, err := os.Open(indexFilename)
	if err != nil {
		return nil, err
	}

	return f, nil
}
