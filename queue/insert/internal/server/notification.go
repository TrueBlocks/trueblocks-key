package server

import (
	"fmt"
	"strconv"

	coreNotify "github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/notify"
	queueItem "github.com/TrueBlocks/trueblocks-key/queue/consume/pkg/item"
)

type Message string

const (
	MessageChunkWritten Message = "chunkWritten"
	MessageStageUpdated Message = "stageUpdated"
	MessageAppearance   Message = "appearance"
)

func CidRange(p *coreNotify.NotificationPayloadChunkWritten) (string, string) {
	return p.Cid, p.Range
}

func Appearances[T coreNotify.NotificationPayload](n *coreNotify.Notification[T]) (apps []*queueItem.Appearance, err error) {
	payload, ok := any(n.Payload).([]coreNotify.NotificationPayloadAppearance)
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
			Address:          item.Address,
			BlockNumber:      uint32(bn),
			TransactionIndex: item.TransactionIndex,
		})
	}
	return
}
