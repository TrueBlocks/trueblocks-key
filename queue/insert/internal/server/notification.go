package server

import (
	"strconv"

	"github.com/TrueBlocks/trueblocks-key/queue/consume/pkg/appearance"
)

// TODO: remove and link trueblocks-core/pkg instead when PR is merged: https://github.com/TrueBlocks/trueblocks-core/pull/3404
type Notification struct {
	Msg     string              `json:"msg"`
	Meta    map[string]any      `json:"meta"`
	Payload NotificationPayload `json:"payload"`
}

type NotificationPayload []struct {
	Address          string `json:"address"`
	BlockNumber      string `json:"blockNumber"`
	TransactionIndex uint32
}

func (n *Notification) Appearances() (apps []*appearance.Appearance, err error) {
	apps = make([]*appearance.Appearance, 0, len(n.Payload))
	for _, item := range n.Payload {
		bn, err := strconv.ParseUint(item.BlockNumber, 10, 32)
		if err != nil {
			return nil, err
		}
		apps = append(apps, &appearance.Appearance{
			Address:       item.Address,
			BlockNumber:   uint32(bn),
			TransactionId: item.TransactionIndex,
		})
	}
	return
}
