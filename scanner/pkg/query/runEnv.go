package query

import "io"

type RunEnv interface {
	Blooms(chain string) (map[string]string, error)
	ReadBloom(fileName string) (io.ReadSeekCloser, error)
	ReadChunk(chain string, fileName string) (io.ReadSeekCloser, error)
}
