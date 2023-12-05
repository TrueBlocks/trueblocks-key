//go:build integration
// +build integration

package dbtest

import "testing"

func TestDbtest(t *testing.T) {
	_, done, err := NewTestConnection()
	if err != nil {
		t.Fatal(err)
	}

	done()
}
