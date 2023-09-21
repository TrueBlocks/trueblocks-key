package chunk

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"sort"

	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/base"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/index"
)

const (
	// HeaderWidth - size of Header Record
	HeaderWidth     = 44
	AddrRecordWidth = 28
	AppRecordWidth  = 8
)

type AddressRecord struct {
	Address base.Address `json:"address"`
	Offset  uint32       `json:"offset"`
	Count   uint32       `json:"count"`
}

func (addressRec *AddressRecord) ReadAddress(reader io.ReadSeeker) (err error) {
	return binary.Read(reader, binary.LittleEndian, addressRec)
}

type ChunkData struct {
	Reader         io.ReadSeeker
	Header         index.IndexHeaderRecord
	Range          base.FileRange
	AddrTableStart int64
	AppTableStart  int64
}

// NewChunkData returns an ChunkData with an opened file pointer to the given fileName. The HeaderRecord
// for the chunk has been populated and the file position to the two tables are ready for use.
func NewChunkData(reader io.ReadSeeker, fileName string) (chunk *ChunkData, err error) {
	blkRange, err := base.RangeFromFilenameE(fileName)
	if err != nil {
		return
	}

	header, err := readIndexHeader(reader)
	if err != nil {
		return
	}

	chunk = &ChunkData{
		Reader:         reader,
		Header:         header,
		AddrTableStart: HeaderWidth,
		AppTableStart:  int64(HeaderWidth + (header.AddressCount * AddrRecordWidth)),
		Range:          blkRange,
	}

	return
}

func readIndexHeader(fl io.ReadSeeker) (header index.IndexHeaderRecord, err error) {
	err = binary.Read(fl, binary.LittleEndian, &header)
	// if err != nil {
	// 	return
	// }

	// Because we call this frequently, we only check that the magic number is correct
	// we let the caller check the hash if needed
	// if header.Magic != file.MagicNumber {
	// 	return header, fmt.Errorf("magic number in file %s is incorrect, expected %d, got %d", fl.Name(), file.MagicNumber, header.Magic)
	// }

	return
}

func (chunk *ChunkData) GetAppearanceRecords(address base.Address) *index.AppearanceResult {
	ret := index.AppearanceResult{Address: address, Range: chunk.Range}

	foundAt := chunk.searchForAddressRecord(address)
	if foundAt == -1 {
		return &ret
	}

	startOfAddressRecord := int64(HeaderWidth + (foundAt * AddrRecordWidth))
	_, err := chunk.Reader.Seek(startOfAddressRecord, io.SeekStart)
	if err != nil {
		ret.Err = err
		return &ret
	}

	addressRecord := AddressRecord{}
	err = addressRecord.ReadAddress(chunk.Reader)
	if err != nil {
		ret.Err = err
		return &ret
	}

	appearances, err := chunk.ReadAppearanceRecords(&addressRecord)
	if err != nil {
		ret.Err = err
		return &ret
	}

	ret.AppRecords = &appearances
	return &ret
}

func (chunk *ChunkData) searchForAddressRecord(address base.Address) int {
	compareFunc := func(pos int) bool {
		if pos == -1 {
			return false
		}

		if pos == int(chunk.Header.AddressCount) {
			return true
		}

		readLocation := int64(HeaderWidth + pos*AddrRecordWidth)
		_, err := chunk.Reader.Seek(readLocation, io.SeekStart)
		if err != nil {
			fmt.Println(err)
			return false
		}

		addressRec := AddressRecord{}
		err = addressRec.ReadAddress(chunk.Reader)
		if err != nil {
			fmt.Println(err)
			return false
		}

		return bytes.Compare(addressRec.Address.Bytes(), address.Bytes()) >= 0
	}

	pos := sort.Search(int(chunk.Header.AddressCount), compareFunc)

	readLocation := int64(HeaderWidth + pos*AddrRecordWidth)
	_, _ = chunk.Reader.Seek(readLocation, io.SeekStart)
	rec := AddressRecord{}
	err := rec.ReadAddress(chunk.Reader)
	if err != nil {
		return -1
	}

	if !bytes.Equal(rec.Address.Bytes(), address.Bytes()) {
		return -1
	}

	return pos
}

func (chunk *ChunkData) ReadAppearanceRecords(addrRecord *AddressRecord) (apps []index.AppearanceRecord, err error) {
	readLocation := int64(HeaderWidth + AddrRecordWidth*chunk.Header.AddressCount + AppRecordWidth*addrRecord.Offset)

	_, err = chunk.Reader.Seek(readLocation, io.SeekStart)
	if err != nil {
		return
	}

	apps = make([]index.AppearanceRecord, addrRecord.Count)
	err = binary.Read(chunk.Reader, binary.LittleEndian, &apps)

	return
}
