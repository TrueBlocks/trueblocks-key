package convertNew

import (
	"context"
	"fmt"
	"log"
	"path"
	"sync/atomic"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/exp/mmap"
)

const batchSize = 5000 // 10000

var insert = `
WITH ids AS (
    INSERT INTO mainnet_addresses (address)
    VALUES ($1::varchar(42))
    ON CONFLICT DO NOTHING
    RETURNING id AS address_id
),
pids AS (
    SELECT address_id FROM ids
    UNION ALL
    SELECT id AS address_id FROM mainnet_addresses WHERE ADDRESS = $1 LIMIT 1
)
INSERT INTO mainnet_appearances (address_id, block_number, tx_id)
SELECT pids.address_id, $2, $3 FROM pids
ON CONFLICT DO NOTHING;
`

func ConvertDir(dirPath string, dsn string) {
	dbpool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Fatalln("unable to create connection pool:", err)
	}
	defer dbpool.Close()

	var doneApps atomic.Int32
	var doneSecs atomic.Int32
	progressContext, cancelProgress := context.WithCancel(context.Background())
	defer cancelProgress()

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				d := doneApps.Load()
				s := doneSecs.Add(1)
				aps := d / s
				fmt.Printf("\rApps done: %d (%d apps/sec)                        ", d, aps)
			case <-progressContext.Done():
				fmt.Println()
				return
			}
		}
	}()

	filePaths := make(chan string, 100)
	go DirFiles(dirPath, filePaths)

	for fileName := range filePaths {
		chunkName := path.Base(fileName)
		chunk, err := mmap.Open(fileName)
		if err != nil {
			log.Fatalln("mmap:", err)
		}
		defer chunk.Close()
		fileSize := chunk.Len()

		resuts := make(chan convertResult, 10000)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go func() {
			defer close(resuts)
			ConvertChunk(ctx, resuts, chunk, chunkName, fileSize)
		}()

		batch := &pgx.Batch{}
		for item := range resuts {
			if err := item.Err; err != nil {
				cancel()
				log.Fatalln("processing error:", err)
			}

			if args := item.Args; len(args) > 0 {
				batch.Queue(insert, args...)
				doneApps.Add(1)
				if batch.Len() >= batchSize {
					if err := saveApps(dbpool, batch); err != nil {
						cancel()
						log.Fatalln("batch insert:", err)
					}
					batch = &pgx.Batch{}
				}
			}
		}

		if err := saveApps(dbpool, batch); err != nil {
			cancel()
			log.Fatalln("batch insert remainder:", err)
		}
	}

	log.Println("Done:", doneApps.Load())
}

func saveApps(dbpool *pgxpool.Pool, batch *pgx.Batch) error {
	if batch.Len() == 0 {
		return nil
	}

	// _, err := dbpool.CopyFrom(
	// 	context.TODO(),
	// 	pgx.Identifier{"appearances"},
	// 	[]string{"id", "address", "block_number", "txid"},
	// 	pgx.CopyFromRows(batch),
	// )

	return dbpool.SendBatch(context.TODO(), batch).Close()
}
