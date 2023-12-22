package convertNew

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
)

const recordSize = 28 // address table record size in bytes

type convertResult struct {
	Err  error
	Args []any
}

type addressResult []any

func ConvertChunk(ctx context.Context, out chan<- convertResult, chunk io.ReaderAt, chunkName string, fileSize int) {
	h, err := NewHeader(chunk)
	if err != nil {
		reportError(out, fmt.Errorf("reading header: %w", err), chunkName)
		return
	}

	workerCount := 100
	if h.AddressCount < 100 {
		workerCount = int(h.AddressCount)
	}
	addrPerWorker := int(h.AddressCount) / workerCount
	lastWorkerExtra := int(h.AddressCount) % workerCount

	// log.Println("Addr records:", h.AddressCount, "Worker cnt:", workerCount, "Addr/worker:", addrPerWorker, "Extra:", lastWorkerExtra)

	var wg sync.WaitGroup

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func(workerIndex int) {
			defer wg.Done()

			startByte := 44 + (workerIndex * recordSize * addrPerWorker) // header + ...
			addrToRead := addrPerWorker
			if workerIndex == workerCount-1 {
				addrToRead += lastWorkerExtra
			}

			records := make([]addressRecord, addrToRead)
			if err := readBytes(chunk, int64(startByte), (addrToRead * recordSize), &records); err != nil {
				reportError(out, fmt.Errorf("reading records: %w", err), chunkName)
				return
			}

			appTableStart := 44 + recordSize*h.AddressCount // header + address table

			for i := 0; i < addrToRead; i++ {
				select {
				case <-ctx.Done():
					return
				default:
					record := records[i]
					addrStr := strings.ToLower(record.Address.Hex())

					apps := make([]appearanceRecord, record.Count)
					if err := readBytes(chunk, int64(appTableStart+8*record.Offset), 8*int(record.Count), &apps); err != nil {
						reportError(out, fmt.Errorf("reading appearances: %w", err), chunkName)
						return
					}
					for j := 0; j < len(apps); j++ {
						app := apps[j]
						// bnStr := strconv.FormatUint(uint64(app.BlockNumber), 10)
						// txidStr := strconv.FormatUint(uint64(app.TransactionId), 10)

						out <- convertResult{
							Args: []any{
								addrStr,
								app.BlockNumber,
								app.TransactionId,
							},
						}
					}
				}
			}
		}(i)
	}

	wg.Wait()
}

func reportError(out chan<- convertResult, err error, fileName string) {
	out <- convertResult{Err: fmt.Errorf("%s: %w", fileName, err)}
}
