package sessions

import (
	"testing"
	"os"
	"fmt"
	"encoding/base64"
	"time"
	"sync"
	"github.com/tjbrains/TeaGo/Tea"
	"github.com/tjbrains/TeaGo/logs"
)

func TestFileSessionManager_Init(t *testing.T) {
	t.Log(os.TempDir())
}

func TestFileSessionManagerEncrypt(t *testing.T) {
	key := "123456"
	key = fmt.Sprintf("%-32s", key)
	data, err := encrypt([]byte("Hello, World"), []byte(key))
	if err != nil {
		t.Error(err)
	} else {
		dataString := base64.StdEncoding.EncodeToString(data)
		t.Log(dataString)
	}
}

func TestFileSessionManagerDecrypt(t *testing.T) {
	key := "123456"
	key = fmt.Sprintf("%-32s", key)

	data, err := base64.StdEncoding.DecodeString("M0LZVKTUSgCfEmcV8kA1icpq+SPsIqFVOrC5qUIkj7Z4JmMu8YtOkw==")
	if err != nil {
		t.Error(err)
		return
	}
	data, err = decrypt(data, []byte(key))
	if err != nil {
		t.Error(err)
		return
	}

	t.Log(string(data))
}

func TestFileSessionManagerEncryptData(t *testing.T) {
	var manager = NewFileSessionManager(1200, "123456")
	data, err := manager.encryptData(&FileSessionData{
		Sid:       "123",
		ExpiredAt: uint(time.Now().Unix() + 1200),
		Values: map[string]string{
			"hello": "World",
		},
	})
	t.Log(data, err)

	if err == nil {
		session, err := manager.decryptData(data)
		t.Log(session, err)
	}
}

func TestFileSessionManager_WriteItem(t *testing.T) {
	var manager = NewFileSessionManager(1200, "123456")
	if manager.WriteItem("123", "hello", "world") {
		t.Log("Write OK")
	} else {
		t.Log("Write Fail")
	}
}

func TestFileSessionManager_Read(t *testing.T) {
	var manager = NewFileSessionManager(1200, "123456")
	t.Log(manager.Read("123"))
}

func TestFileSessionManager_Delete(t *testing.T) {
	var manager = NewFileSessionManager(1200, "123456")
	t.Log(manager.Delete("123"))
}

func TestFileSessionManagerWriteConcurrent(t *testing.T) {
	var manager = NewFileSessionManager(1200, "123456")
	manager.SetDir(Tea.TmpDir())
	wg := sync.WaitGroup{}
	wg.Add(1000)
	for i := 0; i < 1000; i ++ {
		go func() {
			b := manager.WriteItem("123", "a", "b")
			if !b {
				logs.Println("fail")
			}
			wg.Done()
		}()
	}
	wg.Wait()
	t.Log("ok")
}

func TestFileSessionManagerReadConcurrent(t *testing.T) {
	var manager = NewFileSessionManager(1200, "123456")
	manager.SetDir(Tea.TmpDir())
	time.Sleep(1 * time.Second)
	wg := sync.WaitGroup{}
	wg.Add(1000)
	for i := 0; i < 1000; i ++ {
		go func() {
			b := manager.Read("123")
			logs.Println(b)
			wg.Done()
		}()
	}
	wg.Wait()
	t.Log("ok")
}
