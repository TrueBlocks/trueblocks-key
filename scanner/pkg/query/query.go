package query

import (
	"fmt"
	"log"
	"strings"
	"sync"

	"trueblocks.io/searcher/pkg/blkrange"
	"trueblocks.io/searcher/pkg/query/bloom"
	"trueblocks.io/searcher/pkg/query/chunk"
)

func Find(chain string, address string, runEnv RunEnv, results chan chunk.AppearanceRecord) (err error) {
	defer close(results)
	foundRangesCh := make(chan string, 100)

	go func() {
		if err := QueryBlooms(chain, address, foundRangesCh, runEnv); err != nil {
			panic(fmt.Errorf("querying bloom filters: %w", err))
		}
	}()

	for fileRange := range foundRangesCh {
		apps, err := Extract(chain, fileRange, address, runEnv)
		if err != nil {
			panic(err)
		}
		for _, app := range apps {
			results <- app
		}
	}

	return
}

func QueryBlooms(chain string, address string, foundRangesCh chan string, runEnv RunEnv) error {
	blooms, err := runEnv.Blooms(chain)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup

	for rawFileName, fileName := range blooms {
		fileRange, err := blkrange.FromFilename(fileName)
		if err != nil {
			// don't respond further -- there may be foreign files in the folder
			fmt.Println(err)
			continue
		}

		// Run a go routine for each index file
		wg.Add(1)
		fn := fileName
		rfn := rawFileName
		go func() {
			f, err := runEnv.ReadBloom(fn)
			if err != nil {
				panic(err)
			}
			defer f.Close()
			defer wg.Done()
			b, err := bloom.NewBloom(f, fmt.Sprintf("%d-%d", fileRange[0], fileRange[1]))
			if err != nil {
				panic(err)
			}

			v, err := b.IsMember(address)
			if err != nil {
				panic(err)
			}

			if v {
				nfn := strings.Replace(rfn, ".bloom", "", 1)
				foundRangesCh <- nfn
				log.Println("Bloom match:", rfn)
			}
		}()

	}

	wg.Wait()
	close(foundRangesCh)

	return nil
}

func Extract(chain string, fileName string, address string, runEnv RunEnv) (result []chunk.AppearanceRecord, err error) {
	f, err := runEnv.ReadChunk(chain, fileName)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	chunk, err := chunk.NewChunkData(f, fileName)
	if err != nil {
		return
	}
	found, err := chunk.GetAppearanceRecords(address)
	if err != nil {
		return nil, err
	}
	if found == nil {
		return
	}
	result = found
	return
}
