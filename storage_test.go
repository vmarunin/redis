package main

import (
	"fmt"
	"math"
	"sort"
	"testing"
	"time"
)

func TestCRUD(t *testing.T) {
	storage := NewAwfulRedisStorage()
	var v string
	var ok bool
	var l []string

	_, ok = storage.Get("aaa")
	if ok {
		t.Fatalf("Found key in empty storage")
	}
	_, ok = storage.Delete("aaa")
	if ok {
		t.Fatalf("Found key in empty storage")
	}
	_, ok = storage.Set("aaa", "aaa_val1", math.MaxInt)
	if ok {
		t.Fatalf("Found key in empty storage")
	}
	v, ok = storage.Get("aaa")
	if !ok || v != "aaa_val1" {
		t.Fatalf("Key not found or value wrong")
	}
	v, ok = storage.Set("aaa", "aaa_val2", math.MaxInt)
	if !ok || v != "aaa_val1" {
		t.Fatalf("Key not found or value wrong")
	}
	v, ok = storage.Get("aaa")
	if !ok || v != "aaa_val2" {
		t.Fatalf("Key not found or value wrong")
	}
	l, _ = storage.Keys("")
	if len(l) != 1 && l[0] != "aaa" {
		t.Fatalf("Wrong keys")
	}
	v, ok = storage.Delete("aaa")
	if !ok || v != "aaa_val2" {
		t.Fatalf("Key not found or value wrong")
	}
	l, _ = storage.Keys("")
	if len(l) != 0 {
		t.Fatalf("Wrong keys")
	}
}

func TestTTL(t *testing.T) {
	cnt := 4
	ttl := int(time.Now().Unix()) + 1
	storage := NewAwfulRedisStorage()
	var ok bool
	var l []string

	for i := 0; i < cnt; i++ {
		storage.Set(fmt.Sprintf("k_%d", i), fmt.Sprintf("v_%d", i), ttl)
	}

	time.Sleep(2000 * time.Millisecond)

	_, ok = storage.Get("k_0")
	if ok {
		t.Fatalf("Found expired key")
	}
	_, ok = storage.Set("k_1", "v_1_1", math.MaxInt)
	if ok {
		t.Fatalf("Found expired key")
	}
	_, ok = storage.Delete("k_2")
	if ok {
		t.Fatalf("Found expired key")
	}
	l, _ = storage.Keys("")
	if len(l) != 1 || l[0] != "k_1" {
		t.Fatalf("Key not found or value wrong")
	}
}

func TestPattern(t *testing.T) {
	storage := NewAwfulRedisStorage()
	var l []string

	keys := []string{"aaa", "aaaa", "aaab", "\\c?*[]"}
	tests := [][]string{ // pattern and result
		{"", "aaa", "aaaa", "aaab", "\\c?*[]"},
		{"*", "aaa", "aaaa", "aaab", "\\c?*[]"},
		{"a*", "aaa", "aaaa", "aaab"},
		{"*a", "aaa", "aaaa"},
		{"*b", "aaab"},
		{"???", "aaa"},
		{"????", "aaaa", "aaab"},
		{"b*b"},
		{"*[a-c]", "aaa", "aaaa", "aaab"},
		{"*[^c-z\\]]", "aaa", "aaaa", "aaab"},
		{"\\\\c\\?\\*\\[\\]", "\\c?*[]"},
	}
	for _, k := range keys {
		storage.Set(k, "1", math.MaxInt)
	}

	_, err := storage.Keys("[")
	if err == nil {
		t.Fatal("Expected error")
	}

	for _, tc := range tests {
		l, _ = storage.Keys(tc[0])
		expected := tc[1:]
		if len(l) != len(expected) {
			t.Fatal("Wrong len", tc[0], expected, l)
		}
		sort.Strings(l)
		sort.Strings(expected)
		isMatch := true
		for i := range expected {
			if expected[i] != l[i] {
				isMatch = false
				break
			}
		}
		if !isMatch {
			t.Fatal("Wrong keys list", tc[0], expected, l)
		}
	}
}
