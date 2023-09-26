package bloom

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"unsafe"

	"trueblocks.io/searcher/pkg/blkrange"
)

const (
	// The number of bytes in a single BloomByte structure
	BLOOM_WIDTH_IN_BYTES = (BLOOM_WIDTH_IN_BITS / 8)
	// The number of bits in a single BloomByte structure
	BLOOM_WIDTH_IN_BITS = (1048576)
	// The maximum number of addresses to add to a BloomBytes before creating a new one
	MAX_ADDRS_IN_BLOOM = 50000
)

type bitChecker struct {
	whichBits [5]uint32
	offset    uint32
	bit       uint32
	bytes     []byte
}

type Bloom struct {
	Reader     io.ReadSeeker
	Range      [2]uint64
	Header     BloomHeader
	HeaderSize int64
	Count      int32
	Blooms     []BloomBytes
}

type BloomBytes struct {
	NInserted uint32 // Do not change the size of this field, it's stored on disc
	Bytes     []byte
}

type BloomHeader struct {
	Magic uint16 `json:"magic"`
	Hash  Hash   `json:"hash"`
}

type Hash [32]byte

func NewBloom(reader io.ReadSeeker, bloomName string) (b *Bloom, err error) {
	r, err := blkrange.FromFilename(bloomName)
	if err != nil {
		return
	}

	b = &Bloom{
		Reader: reader,
		Range:  r,
	}

	// Read header. We ignore it in this PoC
	if err = binary.Read(b.Reader, binary.LittleEndian, &b.Header); err != nil {
		return
	}

	if err = binary.Read(b.Reader, binary.LittleEndian, &b.Count); err != nil {
		return
	}
	b.HeaderSize = int64(unsafe.Sizeof(b.Header))
	_, _ = b.Reader.Seek(int64(b.HeaderSize), io.SeekStart) // Point to the start of Count
	b.Blooms = make([]BloomBytes, 0, b.Count)

	return
}

func (bl *Bloom) WhichBits(addr string) (bits [5]uint32, err error) {
	slice, err := hex.DecodeString(addr[2:]) // addr.Bytes()
	if err != nil {
		return
	}
	if len(slice) != 20 {
		err = errors.New("address is not 20 bytes long - should not happen")
		return
	}

	cnt := 0
	for i := 0; i < len(slice); i += 4 {
		bytes := slice[i : i+4]
		bits[cnt] = (binary.BigEndian.Uint32(bytes) % uint32(BLOOM_WIDTH_IN_BITS))
		cnt++
	}

	return
}

func (b *Bloom) IsMember(addr string) (bool, error) {
	whichBits, err := b.WhichBits(addr)
	if err != nil {
		return false, err
	}
	offset := uint32(b.HeaderSize) + 4 // the end of Count
	for j := 0; j < int(b.Count); j++ {
		offset += uint32(4) // Skip over NInserted
		var tester = bitChecker{offset: offset, whichBits: whichBits}
		v, err := b.isMember(&tester)
		if err != nil {
			return false, err
		}
		if v {
			return true, nil
		}
		offset += BLOOM_WIDTH_IN_BYTES
	}
	return false, nil
}

func (bl *Bloom) isMember(tester *bitChecker) (bool, error) {
	for _, bit := range tester.whichBits {
		tester.bit = bit
		v, err := bl.isBitLit(tester)
		if err != nil {
			return false, err
		}
		if !v {
			return false, nil
		}
	}
	return true, nil
}

func (bl *Bloom) isBitLit(tester *bitChecker) (bool, error) {
	which := uint32(tester.bit / 8)
	index := uint32(BLOOM_WIDTH_IN_BYTES - which - 1)

	whence := uint32(tester.bit % 8)
	mask := byte(1 << whence)

	var res uint8
	if tester.bytes != nil {
		// In some cases, we've already read the bytes into memory, so use them if they're here
		byt := tester.bytes[index]
		res = byt & mask

	} else {
		var byt uint8
		_, err := bl.Reader.Seek(int64(tester.offset+index), io.SeekStart)
		if err != nil {
			return false, fmt.Errorf("Seek error: %w", err)
		}

		err = binary.Read(bl.Reader, binary.LittleEndian, &byt)
		if err != nil {
			return false, fmt.Errorf("Read error: %w", err)
		}

		res = byt & mask
	}

	return (res != 0), nil
}
