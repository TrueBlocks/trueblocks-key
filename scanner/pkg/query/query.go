package query

import (
	"fmt"
	"log"
	"sync"

	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/base"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/index"
	"trueblocks.io/searcher/pkg/query/bloom"
	"trueblocks.io/searcher/pkg/query/chunk"
)

func Find(chain string, address base.Address, runEnv RunEnv, results chan index.AppearanceRecord) (err error) {
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

func QueryBlooms(chain string, address base.Address, foundRangesCh chan string, runEnv RunEnv) error {
	blooms, err := runEnv.Blooms(chain)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup

	for rawFileName, fileName := range blooms {
		fileRange, err := base.RangeFromFilenameE(fileName)
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
			b, err := bloom.NewBloom(f, fileRange.String())
			if err != nil {
				panic(err)
			}

			v, err := b.IsMember(address)
			if err != nil {
				panic(err)
			}

			if v {
				foundRangesCh <- rfn
				log.Println("Bloom match:", rfn)
			}
		}()

	}

	wg.Wait()
	close(foundRangesCh)

	return nil
}

func Extract(chain string, fileName string, address base.Address, runEnv RunEnv) (result []index.AppearanceRecord, err error) {
	f, err := runEnv.ReadChunk(chain, fileName)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	chunk, err := chunk.NewChunkData(f, fileName)
	if err != nil {
		return
	}
	found := chunk.GetAppearanceRecords(address)
	if found == nil {
		return
	}
	if found.AppRecords == nil {
		return
	}
	if found.Err != nil {
		return nil, found.Err
	}
	result = *found.AppRecords
	return
}
