package integration_test

import (
	"os"
	"testing"

	"trueblocks.io/test/integration/helpers"
)

func TestMain(m *testing.M) {
	cancelSam, waitSam := helpers.StartSam()

	status := m.Run()
	cancelSam()
	waitSam()
	os.Exit(status)
}
