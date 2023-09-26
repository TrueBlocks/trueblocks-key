package blkrange

import (
	"path"
	"strconv"
	"strings"
)

func FromFilename(fileName string) (blkRange [2]uint64, err error) {
	rangeStr := path.Base(fileName)
	rangeStr = strings.Replace(rangeStr, ".bin", "", 1)
	rangeStr = strings.Replace(rangeStr, ".bloom", "", 1)

	numbers := strings.Split(rangeStr, "-")
	blkRange[0], err = strconv.ParseUint(numbers[0], 10, 64)
	if err != nil {
		return
	}
	blkRange[1], err = strconv.ParseUint(numbers[1], 10, 64)
	if err != nil {
		return
	}

	return blkRange, nil
}
