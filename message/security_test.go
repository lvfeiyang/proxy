package message

import "testing"

func TestNewKey(t *testing.T) {
	key := NewKey(34)
	if string(key) == "yao32bytes nameduo zenmexie aaaa" {
		t.Error("key no change")
	}
}

func TestAes(t *testing.T) {
	info := "hello world"
	e, _ := AesEn([]byte(info), NewKey(78))
	// t.Log(e)
	d, _ := AesDe(e, NewKey(78))
	if string(d) != info {
		t.Error("d is not same info")
	}
}

func TestSign(t *testing.T) {
	info_1 := []byte("hello world")
	// info_2 := []byte("hello leon")
	sign_1, _ := Sign(info_1, NewKey(99))
	// sign_2 := Sign(info_2)
	if Verify(sign_1, info_1, NewKey(99)) == false {
		t.Error("verify info 1 err")
	}
}
