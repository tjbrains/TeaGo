// Copyright 2026 FlexCDN root@flexcdn.cn. All rights reserved. Official site: https://flexcdn.cn .

package logs

type TestingInterface interface {
	Log(args ...any)
	Logf(format string, args ...any)
	Fatal(args ...any)
	Fail()
}
