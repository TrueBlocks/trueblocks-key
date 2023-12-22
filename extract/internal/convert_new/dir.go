package convertNew

import (
	"log"
	"os"
	"path/filepath"
	"sort"
)

func DirFiles(dirPath string, out chan<- string) {
	defer close(out)

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		log.Fatalln("readdir:", err)
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() > entries[j].Name()
	})
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		fileName := entry.Name()
		if filepath.Ext(fileName) != ".bin" {
			continue
		}
		out <- filepath.Join(dirPath, fileName)
	}
}
