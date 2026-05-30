// Copyright 2026 FlexCDN root@flexcdn.cn. All rights reserved. Official site: https://flexcdn.cn .

package dbs

import (
	"database/sql"
	"errors"
	"strconv"
	"sync"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/tjbrains/TeaGo/containers"
	"github.com/tjbrains/TeaGo/logs"
)

func IsPrepareError(err error) bool {
	if err == nil {
		return false
	}
	mysqlErr, isMySQLErr := err.(*mysql.MySQLError)
	return isMySQLErr && mysqlErr.Number == 1461
}

type sqlPreparer interface {
	Prepare(query string) (*sql.Stmt, error)
}

var timestamp = time.Now().Unix()
var timeTicker = time.NewTicker(500 * time.Millisecond)

func init() {
	go func() {
		for range timeTicker.C {
			timestamp = time.Now().Unix()
		}
	}()
}

func unixTime() int64 {
	return timestamp
}

type StmtManager struct {
	stmtMap map[string]*Stmt        // key => *Stmt
	subMap  map[int64][]string      // id => [cache keys]
	stmtLRU *containers.LRU[string] // key

	mu sync.RWMutex

	isClosed bool
}

func NewStmtManager() *StmtManager {
	var maxCount = 1 << 10
	var manager = &StmtManager{
		stmtMap: map[string]*Stmt{},
		stmtLRU: containers.NewLRU[string](maxCount, containers.LRURoundTouch[string]()),
		subMap:  map[int64][]string{},
	}

	manager.stmtLRU.OnEvict(manager.onEvict)

	return manager
}

func (this *StmtManager) SetMaxCount(count int) {
	if count > 0 {
		this.stmtLRU.SetCapacity(count)
	}
}

// Prepare statement
func (this *StmtManager) Prepare(preparer sqlPreparer, querySQL string) (*Stmt, error) {
	if this.isClosed {
		return nil, errors.New("prepare failed: connection is closed")
	}

	if ShowPreparedStatements {
		logs.Println("[DB]prepare " + querySQL)
	}

	sqlStmt, err := preparer.Prepare(querySQL)
	if err != nil {
		if IsPrepareError(err) {
			// lock for concurrent operation
			this.mu.Lock()
			this.purge()
			this.mu.Unlock()

			// retry
			sqlStmt, err = preparer.Prepare(querySQL)
		}
		if err != nil {
			return nil, err
		}
	}

	return NewStmt(sqlStmt), nil
}

// PrepareOnce prepare statement for reuse
func (this *StmtManager) PrepareOnce(preparer sqlPreparer, querySQL string, parentId int64) (resultStmt *Stmt, wasCached bool, err error) {
	var cacheKey string
	if parentId == 0 {
		cacheKey = "0$" + querySQL
	} else {
		cacheKey = strconv.FormatInt(parentId, 10) + "$" + querySQL
	}

	// check if exists
	this.mu.RLock()
	stmt, ok := this.stmtMap[cacheKey]
	if ok {
		this.mu.RUnlock()

		this.stmtLRU.TryTouch(cacheKey)

		return stmt, true, nil
	}
	this.mu.RUnlock()

	if ShowPreparedStatements {
		logs.Println("[DB]prepare " + querySQL)
	}

	sqlStmt, err := preparer.Prepare(querySQL)
	if err != nil {
		if IsPrepareError(err) {
			// purge once
			this.purge()

			// retry
			sqlStmt, err = preparer.Prepare(querySQL)
		}
		if err != nil {
			return nil, false, err
		}
	}
	stmt = NewStmt(sqlStmt)

	this.mu.Lock()
	defer this.mu.Unlock()

	// exists, check again
	_, exists := this.stmtMap[cacheKey]
	if exists {
		return stmt, false, nil
	}

	// touch first to hold a position
	this.stmtLRU.Touch(cacheKey)

	// put stmt into cache map
	this.stmtMap[cacheKey] = stmt
	if parentId > 0 {
		this.subMap[parentId] = append(this.subMap[parentId], cacheKey)
	}

	return stmt, true, nil
}

func (this *StmtManager) Close() error {
	this.isClosed = true

	this.mu.Lock()
	var stmtMap = this.stmtMap
	this.stmtMap = map[string]*Stmt{}
	this.stmtLRU.Clear()
	this.mu.Unlock()

	var firstError error
	for _, stmt := range stmtMap {
		err := stmt.Close()
		if err != nil && firstError == nil {
			firstError = err
		}
	}

	return firstError
}

func (this *StmtManager) CloseId(parentId int64) error {
	// collect dirty stmts
	this.mu.Lock()

	cacheKeys, ok := this.subMap[parentId]
	if !ok {
		this.mu.Unlock()
		return nil
	}
	delete(this.subMap, parentId)

	var dirtyStmts = []*Stmt{}
	for _, cacheKey := range cacheKeys {
		stmt, stmtOk := this.stmtMap[cacheKey]
		if stmtOk {
			dirtyStmts = append(dirtyStmts, stmt)

			this.stmtLRU.Delete(cacheKey)
			delete(this.stmtMap, cacheKey)
		}
	}

	this.mu.Unlock()

	// close dirty stmts
	var firstError error
	for _, stmt := range dirtyStmts {
		err := stmt.Close()
		if err != nil && firstError == nil {
			firstError = err
		}
	}

	return firstError
}

func (this *StmtManager) Len() int {
	this.mu.Lock()
	defer this.mu.Unlock()
	return len(this.stmtMap)
}

func (this *StmtManager) purge() {
	this.stmtLRU.Evict(4)
}

func (this *StmtManager) onEvict(keys []string) {
	if len(keys) == 0 {
		return
	}

	var dirtyStmts = make([]*Stmt, 0, len(keys))

	this.mu.Lock()
	for _, key := range keys {
		stmt, ok := this.stmtMap[key]
		if ok {
			delete(this.stmtMap, key)
			dirtyStmts = append(dirtyStmts, stmt)
		}
	}
	this.mu.Unlock()

	for _, stmt := range dirtyStmts {
		_ = stmt.Close()
	}
}
