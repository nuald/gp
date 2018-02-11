package cmd

import "testing"

func TestEncrypt(t *testing.T) {
	src := "test"
	key := "123"
	if _, err := encrypt(src, key); err == nil {
		t.Error("Should fail with the invalid key")
	}

	key = "12345678901234567890123456789012"
	encrypted, err := encrypt(src, key)
	if err != nil {
		t.Error(err)
	} else {
		dst, err := decrypt(encrypted, key)
		if err != nil {
			t.Error(err)
		} else if src != string(dst) {
			t.Error("Expected", src, "got", dst)
		}
	}
}
