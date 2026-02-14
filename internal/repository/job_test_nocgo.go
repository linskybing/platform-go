//go:build !cgo
// +build !cgo

package repository

import "testing"

func TestJobRepo_CGODisabled(t *testing.T) {
	t.Skip("job repo tests require CGO for sqlite3")
}
