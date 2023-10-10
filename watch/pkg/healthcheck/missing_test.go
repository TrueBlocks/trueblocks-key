//go:build integration
// +build integration

package healthcheck

import (
	"testing"

	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/base"
	database "trueblocks.io/database/pkg"
	"trueblocks.io/database/pkg/dbtest"
)

var conn *database.Connection

func init() {
	var err error
	conn, err = dbtest.NewTestConnection()
	if err != nil {
		panic(err)
	}
}

func TestMissingAppearances(t *testing.T) {
	blockRange := [2]uint64{
		4039898,
		4053179,
	}
	dbtx := conn.Db().Create(&database.Appearance{
		BlockRangeStart: blockRange[0],
		BlockRangeEnd:   blockRange[1],
		Address:         "0xf503017d7baf7fbc0fff7492b751025c6a78179b",
		BlockNumber:     uint32(blockRange[0]),
		TransactionId:   1,
	})
	if err := dbtx.Error; err != nil {
		t.Fatal(err)
	}

	var err error
	var diff int
	var expected uint

	expected = 1
	diff, err = MissingAppearances(
		conn,
		base.BlockRange{
			First: blockRange[0],
			Last:  blockRange[1],
		},
		expected,
	)
	if err != nil {
		t.Fatal(err)
	}
	if diff != 0 {
		t.Fatal("wrong value:", diff, "expected", expected)
	}

	expected = 5
	diff, err = MissingAppearances(
		conn,
		base.BlockRange{
			First: blockRange[0],
			Last:  blockRange[1],
		},
		expected,
	)
	if err != nil {
		t.Fatal(err)
	}
	if diff != -4 {
		t.Fatal("wrong value:", diff, "expected", expected)
	}

	expected = 0
	diff, err = MissingAppearances(
		conn,
		base.BlockRange{
			First: blockRange[0],
			Last:  blockRange[1],
		},
		expected,
	)
	if err != nil {
		t.Fatal(err)
	}
	if diff != 1 {
		t.Fatal("wrong value:", diff, "expected", expected)
	}
}
