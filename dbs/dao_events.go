// Copyright 2024 FlexCDN root@flexcdn.cn. All rights reserved. Official site: https://flexcdn.cn .

package dbs

var daoInitBeforeCallback func(dao DAOInterface)
var daoInitErrorCallback func(dao DAOInterface, err error) error

func OnDAOInitBefore(f func(dao DAOInterface)) {
	daoInitBeforeCallback = f
}

func OnDAOInitError(f func(dao DAOInterface, err error) error) {
	daoInitErrorCallback = f
}
