package memdb

import (
	"testing"
	"time"
)

func TestNewToken(t *testing.T) {
	token := newToken(1, time.Second, "test")

	token.Set("1", "1")
	token.Set("2", "2")
	token.Set("3", "3")

	data, ok := token.Get("1", false)
	if !ok {
		t.Fatal(ok)
	}
	t.Log("1=", data)
	data, ok = token.Get("2", false)
	if !ok {
		t.Fatal(ok)
	}
	t.Log("2=", data)
	data, ok = token.Get("3", false)
	if !ok {
		t.Fatal(ok)
	}
	t.Log("3=", data)

	items := token.Lst("")
	if len(items) != 3 {
		t.Fatal(items)
	}
	t.Log(items)

	time.Sleep(30 * time.Second)
	token.Del("1")
	data, ok = token.Get("1", false)
	if ok {
		t.Fatal(ok)
	}
	data, ok = token.Get("2", true)
	if !ok {
		t.Fatal(ok)
	}
	t.Log("2=", data)
	data, ok = token.Get("3", false)
	if !ok {
		t.Fatal(ok)
	}
	t.Log("3=", data)

	time.Sleep(35 * time.Second)
	data, ok = token.Get("2", false)
	if !ok {
		t.Fatal(ok)
	}
	t.Log("2=", data)
	data, ok = token.Get("3", false)
	if ok {
		t.Fatal(ok)
	}

}
