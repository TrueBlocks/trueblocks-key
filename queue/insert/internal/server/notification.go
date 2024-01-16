package server

import (
	"fmt"
	"strconv"

	queueItem "github.com/TrueBlocks/trueblocks-key/queue/consume/pkg/item"
)

type Message string

const (
	MessageChunkWritten Message = "chunkWritten"
	MessageStageUpdated Message = "stageUpdated"
	MessageAppearance   Message = "appearance"
)

// TODO: remove and link trueblocks-core/pkg instead when PR is merged: https://github.com/TrueBlocks/trueblocks-core/pull/3404
type Notification[T NotificationPayload] struct {
	Msg     Message        `json:"msg"`
	Meta    map[string]any `json:"meta"`
	Payload T              `json:"payload"`
}

type NotificationPayload interface {
	[]NotificationPayloadAppearance |
		[]NotificationPayloadChunkWritten |
		NotificationPayloadChunkWritten |
		string
}

type NotificationPayloadAppearance struct {
	Address          string `json:"address"`
	BlockNumber      string `json:"blockNumber"`
	TransactionIndex uint32 `json:"txid"`
}

type NotificationPayloadChunkWritten struct {
	Cid    string `json:"cid"`
	Range  string `json:"range"`
	Author string `json:"author"`
}

func (p *NotificationPayloadChunkWritten) CidRange() (string, string) {
	return p.Cid, p.Range
}

func (n *Notification[T]) Appearances() (apps []*queueItem.Appearance, err error) {
	payload, ok := any(n.Payload).([]NotificationPayloadAppearance)
	if !ok {
		err = fmt.Errorf("notification is not appearance notification: %s", n.Msg)
		return
	}

	apps = make([]*queueItem.Appearance, 0, len(payload))
	for _, item := range payload {
		bn, err := strconv.ParseUint(item.BlockNumber, 10, 32)
		if err != nil {
			return nil, err
		}
		apps = append(apps, &queueItem.Appearance{
			Address:       item.Address,
			BlockNumber:   uint32(bn),
			TransactionId: item.TransactionIndex,
		})
	}
	return
}
