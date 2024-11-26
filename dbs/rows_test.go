// Copyright 2024 FlexCDN root@flexcdn.cn. All rights reserved. Official site: https://flexcdn.cn .

package dbs_test

import (
	"github.com/tjbrains/TeaGo/dbs"
	"github.com/tjbrains/TeaGo/rands"
	"testing"
)

func TestRows_Columns(t *testing.T) {
	setupUserQuery()

	stmt, err := testDBInstance.Prepare("SELECT * FROM users")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = stmt.Close()
	}()

	rows, err := stmt.FindRows()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = rows.Close()
	}()

	t.Log(rows.Columns())
	t.Log(rows.Columns())
}

func BenchmarkRows_Columns(b *testing.B) {
	setupUserQuery()

	stmt, err := testDBInstance.Prepare("SELECT * FROM users")
	if err != nil {
		b.Fatal(err)
	}
	defer func() {
		_ = stmt.Close()
	}()

	rows, err := stmt.FindRows()
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, execErr := rows.Columns()
			if execErr != nil {
				b.Fatal(execErr)
			}
		}
	})
}

func BenchmarkRows_FindOnes(b *testing.B) {
	setupUserQuery()

	const count = 1

	var stmts [count]*dbs.Stmt

	for i := range count {
		stmt, err := testDBInstance.Prepare("SELECT * FROM users")
		if err != nil {
			b.Fatal(err)
		}
		stmts[i] = stmt
	}

	b.ReportAllocs()

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var stmt = stmts[rands.Int(0, count-1)]
			rows, execErr := stmt.FindRows()
			if execErr != nil {
				b.Fatal(execErr)
			}
			execErr = rows.Close()
			if execErr != nil {
				b.Fatal(execErr)
			}
		}
	})
}
