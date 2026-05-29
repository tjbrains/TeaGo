// Copyright 2026 FlexCDN root@flexcdn.cn. All rights reserved. Official site: https://flexcdn.cn .

package bootsrap_test

import (
	"testing"

	"github.com/tjbrains/TeaGo/Tea"
	_ "github.com/tjbrains/TeaGo/bootstrap"
)

func TestBootstrapInit(t *testing.T) {
	t.Log(Tea.Root)
}
