package query

import (
	"bytes"

	"encoding/base64"
	"encoding/binary"
	"encoding/json"

	database "github.com/TrueBlocks/trueblocks-key/database/pkg"
)

type PageId struct {
	DirectionNextPage bool
	LastBlock         uint32
	LastSeen          database.Appearance
	// BlockNumber       uint32
	// TransactionIndex  uint32
	LatestInSet   database.Appearance
	EarliestInSet database.Appearance
}

func (p *PageId) MarshalText() (text []byte, err error) {
	buf := &bytes.Buffer{}
	if err = binary.Write(buf, binary.LittleEndian, p); err != nil {
		return
	}
	result := base64.StdEncoding.EncodeToString(buf.Bytes())
	text = []byte(result)
	return
}

func (p *PageId) UnmarshalText(text []byte) (err error) {
	var b []byte
	b, err = base64.StdEncoding.DecodeString(string(text))
	if err != nil {
		return
	}

	var result PageId
	if err = binary.Read(bytes.NewReader(b), binary.LittleEndian, &result); err != nil {
		return
	}
	*p = result
	return
}

func (p *PageId) MarshalJSON() ([]byte, error) {
	b, err := p.MarshalText()
	if err != nil {
		return nil, err
	}
	return json.Marshal(b)
}

func (p *PageId) UnmarshalJSON(b []byte) (err error) {
	var enc []byte
	if err = json.Unmarshal(b, &enc); err != nil {
		return
	}

	return p.UnmarshalText(enc)
}

type PageIdSpecial string

const (
	PageIdLatest    PageIdSpecial = "latest"
	PageIdEarliest  PageIdSpecial = "earliest"
	PageIdNoSpecial PageIdSpecial = ""
)

func (p *PageIdSpecial) FromBytes(b []byte) bool {
	value := PageIdSpecial(b)
	switch value {
	// case PageIdNoSpecial:
	// 	fallthrough
	case PageIdEarliest:
		fallthrough
	case PageIdLatest:
		*p = value
		return true
	}

	return false
}
