package queuetest

import (
	"fmt"

	"trueblocks.io/queue/consume/pkg/appearance"
)

type MockQueue struct {
	items []*appearance.Appearance
}

func (m *MockQueue) Init() error {
	return nil
}

func (m *MockQueue) Add(app *appearance.Appearance) (string, error) {
	m.items = append(m.items, app)
	return fmt.Sprintf("%d", m.Len()), nil
}

func (m *MockQueue) Len() int {
	return len(m.items)
}

func (m *MockQueue) Get(index int) *appearance.Appearance {
	return m.items[index]
}
