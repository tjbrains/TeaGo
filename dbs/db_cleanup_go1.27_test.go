//go:build go1.27

package dbs

import (
	"runtime"
	"strings"
	"testing"
)

func TestAddCleanupClosesOverPtrPanics(t *testing.T) {
	// Same shape as the old NewInstanceFromConfig code that panicked on Go 1.27+:
	//   runtime.AddCleanup(db, func(int) { _ = db.Close() }, 1)
	db := &DB{stmtManager: NewStmtManager()}

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic when cleanup closes over ptr")
		}
		msg, ok := r.(string)
		if !ok {
			t.Fatalf("unexpected panic type %T: %v", r, r)
		}
		if !strings.Contains(msg, "closes over ptr") {
			t.Fatalf("unexpected panic: %s", msg)
		}
	}()

	runtime.AddCleanup(db, func(s int) {
		_ = db.Close()
	}, 1)
}
