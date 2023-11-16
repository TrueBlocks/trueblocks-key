package integration_test

import (
	"os"
	"testing"

	"github.com/TrueBlocks/trueblocks-key/test/integration/helpers"
)

func TestMain(m *testing.M) {
	cancelSam, waitSam := helpers.StartSam()

	status := m.Run()
	cancelSam()
	waitSam()
	os.Exit(status)
}
