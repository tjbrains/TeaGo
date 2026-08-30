package dbs

import (
	"runtime"
	"strings"
	"testing"
)

func TestNewInstanceFromConfig_AddCleanupDoesNotPanic(t *testing.T) {
	db, err := NewInstanceFromConfig(&DBConfig{
		Driver: "mysql",
		Dsn:    "user:pass@tcp(127.0.0.1:1)/unused?timeout=1s",
	})
	if err != nil {
		t.Fatal(err)
	}
	if db == nil || db.Raw() == nil {
		t.Fatal("expected db with raw handle")
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestCloseUnreachableDB_NilSafe(t *testing.T) {
	closeUnreachableDB(dbCleanupArg{})
}

func TestAddCleanupClosesOverPtrPanics(t *testing.T) {
	// Documents the runtime rule that the old
	//   runtime.AddCleanup(db, func(int) { _ = db.Close() }, 1)
	// tripped: the cleanup must not keep ptr reachable.
	type obj struct{ n int }
	p := &obj{n: 1}

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

	runtime.AddCleanup(p, func(int) {
		_ = p.n
	}, 1)
}
