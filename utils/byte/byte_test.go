package byteutil_test

import (
	"encoding/base64"
	byteutil "github.com/tjbrains/TeaGo/utils/byte"
	"testing"
)

func TestEncrypt(t *testing.T) {
	encrypted, err := byteutil.Encrypt([]byte("Hello, World"), []byte("hello"))
	if err != nil {
		t.Fatal(err)
	}

	t.Log(len(encrypted), base64.StdEncoding.EncodeToString(encrypted))

	decrypted, err := byteutil.Decrypt(encrypted, []byte("hello"))
	if err != nil {
		t.Fatal(err)
	}
	t.Log(len(decrypted), string(decrypted))
}
