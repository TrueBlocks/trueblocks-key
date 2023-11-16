package chunk

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"sort"

	"github.com/TrueBlocks/trueblocks-key/searcher/pkg/blkrange"
)

const (
	// HeaderWidth - size of Header Record
	HeaderWidth     = 44
	AddrRecordWidth = 28
	AppRecordWidth  = 8
)

type AddressRecord struct {
	Address [20]byte `json:"address"`
	Offset  uint32   `json:"offset"`
	Count   uint32   `json:"count"`
}

type AppearanceRecord struct {
	BlockNumber   uint32 `json:"blockNumber"`
	TransactionId uint32 `json:"transactionIndex"`
}

func (addressRec *AddressRecord) ReadAddress(reader io.ReadSeeker) (err error) {
	return binary.Read(reader, binary.LittleEndian, addressRec)
}

type ChunkData struct {
	Reader         io.ReadSeeker
	Header         IndexHeaderRecord
	Range          [2]uint64
	AddrTableStart int64
	AppTableStart  int64
}

type IndexHeaderRecord struct {
	Magic           uint32
	Hash            Hash
	AddressCount    uint32
	AppearanceCount uint32
}

type Hash [32]byte

// NewChunkData returns an ChunkData with an opened file pointer to the given fileName. The HeaderRecord
// for the chunk has been populated and the file position to the two tables are ready for use.
func NewChunkData(reader io.ReadSeeker, fileName string) (chunk *ChunkData, err error) {
	blkRange, err := blkrange.FromFilename(fileName)
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

func readIndexHeader(fl io.ReadSeeker) (header IndexHeaderRecord, err error) {
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

func (chunk *ChunkData) GetAppearanceRecords(address string) ([]AppearanceRecord, error) {
	// ret := index.AppearanceResult{Address: address, Range: chunk.Range}

	foundAt := chunk.searchForAddressRecord(address)
	if foundAt == -1 {
		return nil, nil
	}

	startOfAddressRecord := int64(HeaderWidth + (foundAt * AddrRecordWidth))
	_, err := chunk.Reader.Seek(startOfAddressRecord, io.SeekStart)
	if err != nil {
		return nil, err
	}

	addressRecord := AddressRecord{}
	err = addressRecord.ReadAddress(chunk.Reader)
	if err != nil {
		return nil, err
	}

	appearances, err := chunk.ReadAppearanceRecords(&addressRecord)
	if err != nil {
		return nil, err
	}

	return appearances, nil
}

func (chunk *ChunkData) searchForAddressRecord(address string) int {
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

		addrBytes, err := hex.DecodeString(address[2:])
		if err != nil {
			fmt.Println(err)
			return false
		}

		return bytes.Compare(addressRec.Address[:], addrBytes) >= 0
	}

	pos := sort.Search(int(chunk.Header.AddressCount), compareFunc)

	readLocation := int64(HeaderWidth + pos*AddrRecordWidth)
	_, _ = chunk.Reader.Seek(readLocation, io.SeekStart)
	rec := AddressRecord{}
	err := rec.ReadAddress(chunk.Reader)
	if err != nil {
		return -1
	}

	addrBytes, err := hex.DecodeString(address[2:])
	if err != nil {
		fmt.Println(err)
		return -1
	}

	if !bytes.Equal(rec.Address[:], addrBytes) {
		return -1
	}

	return pos
}

func (chunk *ChunkData) ReadAppearanceRecords(addrRecord *AddressRecord) (apps []AppearanceRecord, err error) {
	readLocation := int64(HeaderWidth + AddrRecordWidth*chunk.Header.AddressCount + AppRecordWidth*addrRecord.Offset)

	_, err = chunk.Reader.Seek(readLocation, io.SeekStart)
	if err != nil {
		return
	}

	apps = make([]AppearanceRecord, addrRecord.Count)
	err = binary.Read(chunk.Reader, binary.LittleEndian, &apps)

	return
}
