package query

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
)

type RpcGetAppearancesParam struct {
	Address   string           `json:"address"`
	LastBlock *json.RawMessage `json:"lastBlock,omitempty"`
	PageId    json.RawMessage  `json:"pageId,omitempty"`
	PerPage   uint             `json:"perPage"`
}

func (r *RpcGetAppearancesParam) Limit() uint {
	return r.PerPage
}

func (r *RpcGetAppearancesParam) Validate() error {
	if err := validateAddress(r.Address); err != nil {
		return err
	}

	if err := validateLimit(r); err != nil {
		return err
	}

	return nil
}

// LastBlock returns nil for latest block, block number otherwise
func (r *RpcGetAppearancesParam) LastBlockNumber() (*uint, error) {
	if r.LastBlock == nil {
		return nil, nil
	}

	var special string
	err := json.Unmarshal(*r.LastBlock, &special)
	if err != nil {
		log.Println("last block is not special because of error:", err)
	} else {
		// it is special
		if special == LastBlockLatest {
			return nil, nil
		} else {
			// it's invalid
			return nil, ErrInvalidLastBlockSpecial
		}
	}

	// Try a number
	var blockNumber uint
	if err := json.Unmarshal(*r.LastBlock, &blockNumber); err != nil {
		return nil, ErrInvalidLastBlockInvalid
	}

	return &blockNumber, nil
}

func (r *RpcGetAppearancesParam) PageIdValue() (special PageIdSpecial, pageId *PageId, err error) {
	raw := r.PageId
	// no pageId means "latest"
	if len(raw) == 0 || string(raw) == "null" || string(raw) == `""` {
		special = PageIdLatest
		return
	}
	var specialPageId PageIdSpecial
	if ok := specialPageId.FromBytes(bytes.Trim(raw, `"`)); ok {
		special = specialPageId
		return
	}

	pageId = &PageId{}
	err = json.Unmarshal(raw, pageId)
	return
}

func (r *RpcGetAppearancesParam) SetPageId(specialPageId PageIdSpecial, pageId *PageId) error {
	var value any
	if specialPageId != "" {
		value = specialPageId
	} else {
		value = pageId
	}
	b, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("setting page id: %w", err)
	}

	raw := json.RawMessage(b)
	r.PageId = raw
	return nil
}
