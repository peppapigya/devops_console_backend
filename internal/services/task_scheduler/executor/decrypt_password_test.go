package executor

import (
	"testing"
)

func TestDecryptPassword(t *testing.T) {
	password := "+RN69XTFMCBfqYg=="
	decrypted := DecryptPassword(password)
	println(decrypted)
}
