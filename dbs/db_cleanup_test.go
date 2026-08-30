package dbs

import (
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
