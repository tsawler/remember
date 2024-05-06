package remember

import (
	"encoding/gob"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	ops := Options{
		Server: testRedis.Host(),
		Port:   testRedis.Port(),
		Prefix: "test_cache",
	}
	c1, _ := New("redis", ops)

	err := c1.Set("foo", "bar")
	if err != nil {
		t.Error("error", err)
	}

	c2, _ := New("redis")
	err = c2.Set("foo", "bar")
	if err != nil {
		t.Error("error", err)
	}

	_, err = New("fish")
	if err == nil {
		t.Error("expected error but did not get one")
	}
}

func TestSet(t *testing.T) {
	var tests = []struct {
		name          string
		key           string
		data          string
		expires       time.Duration
		wait          time.Duration
		errorExpected bool
	}{
		{
			name:          "valid",
			key:           "v1",
			data:          "bar",
			expires:       0,
			wait:          0,
			errorExpected: false,
		},

		{
			name:          "expired",
			key:           "v2",
			data:          "bar",
			expires:       time.Millisecond,
			wait:          2 * time.Millisecond,
			errorExpected: true,
		},
	}

	for _, tt := range tests {
		err := testRedisCache.Set(tt.key, tt.data, tt.expires)
		if err != nil && !tt.errorExpected {
			t.Errorf("%s: received unexpected error: %s", tt.name, err.Error())
		}

		if tt.wait != 0 && tt.expires != 0 {
			time.Sleep(tt.wait)
		}

		retrieved, _ := testRedisCache.Get(tt.key)
		if !tt.errorExpected && retrieved != tt.data {
			t.Errorf("%s: incorrect value retrieved; got %s expected %s", tt.name, retrieved, tt.data)
		}

	}

	testRedisCache.Empty()
}

func TestGetInt(t *testing.T) {

	var tests = []struct {
		name          string
		key           string
		data          int
		setVal        bool
		errorExpected bool
	}{
		{
			name:          "valid",
			key:           "v1",
			data:          10,
			setVal:        true,
			errorExpected: false,
		},

		{
			name:          "no key",
			key:           "non_existent",
			data:          11,
			setVal:        false,
			errorExpected: true,
		},
	}

	for _, tt := range tests {
		if tt.setVal {
			testRedisCache.Set(tt.key, tt.data)
		}

		x, err := testRedisCache.GetInt(tt.key)
		if !tt.errorExpected && err != nil {
			t.Errorf("%s: received unexpected error: %s", tt.name, err.Error())
		}

		if !tt.errorExpected && x != tt.data {
			t.Errorf("%s: wrong value retrieved from cache; expected %d but got %d", tt.name, tt.data, x)
		}
	}
	testRedisCache.Empty()
}

func TestGetString(t *testing.T) {

	var tests = []struct {
		name          string
		key           string
		data          string
		setVal        bool
		errorExpected bool
	}{
		{
			name:          "valid",
			key:           "v1",
			data:          "alpha",
			setVal:        true,
			errorExpected: false,
		},

		{
			name:          "no key",
			key:           "non_existent",
			data:          "beta",
			setVal:        false,
			errorExpected: true,
		},
	}

	for _, tt := range tests {
		if tt.setVal {
			testRedisCache.Set(tt.key, tt.data)
		}

		x, err := testRedisCache.GetString(tt.key)
		if !tt.errorExpected && err != nil {
			t.Errorf("%s: received unexpected error: %s", tt.name, err.Error())
		}

		if !tt.errorExpected && x != tt.data {
			t.Errorf("%s: wrong value retrieved from cache; expected %s but got %s", tt.name, tt.data, x)
		}
	}
	testRedisCache.Empty()
}

func TestForget(t *testing.T) {
	_ = testRedisCache.Set("forgetme", "x")
	err := testRedisCache.Forget("forgetme")
	if err != nil {
		t.Error(err)
	}

	_, err = testRedisCache.Get("forgetme")
	if err == nil {
		t.Error("got something from the cache, and it should not be there")
	}
	testRedisCache.Empty()
}

func TestHas(t *testing.T) {
	_ = testRedisCache.Set("fish", "x")
	if !testRedisCache.Has("fish") {
		t.Error("should have value in cache, but do not")
	}

	_ = testRedisCache.Forget("fish")
	if testRedisCache.Has("fish") {
		t.Error("value should not exist in cache, but it does")
	}

	testRedisCache.Empty()
}

func TestGetTime(t *testing.T) {
	gob.Register(time.Time{})
	testTime := time.Now()

	var tests = []struct {
		name          string
		key           string
		data          time.Time
		setVal        bool
		errorExpected bool
	}{
		{
			name:          "valid",
			key:           "v1",
			data:          testTime,
			setVal:        true,
			errorExpected: false,
		},

		{
			name:          "no key",
			key:           "non_existent",
			data:          testTime,
			setVal:        false,
			errorExpected: true,
		},
	}

	for _, tt := range tests {
		if tt.setVal {
			err := testRedisCache.Set(tt.key, tt.data)
			if err != nil {
				t.Error(err)
			}
		}

		x, err := testRedisCache.GetTime(tt.key)
		if !tt.errorExpected && err != nil {
			t.Errorf("%s: received unexpected error: %s", tt.name, err.Error())
		}

		if !tt.errorExpected && !tt.data.Equal(x) {
			t.Errorf("%s: wrong value retrieved from cache; expected %s but got %s", tt.name, tt.data, x)
		}
	}
	testRedisCache.Empty()
}

func TestEmptyByMatch(t *testing.T) {
	_ = testRedisCache.Set("fooa", "bar")
	_ = testRedisCache.Set("foob", "bar")
	_ = testRedisCache.Set("fooc", "bar")

	err := testRedisCache.EmptyByMatch("foo")
	if err != nil {
		t.Error(err)
	}

	if testRedisCache.Has("fooa") ||
		testRedisCache.Has("foob") ||
		testRedisCache.Has("fooc") {
		t.Error("cache still has values beginning with foo")
	}

	testRedisCache.Empty()
}

func TestEmpty(t *testing.T) {
	err := testRedisCache.Set("x", "y")
	if err != nil {
		t.Error("error setting value in cache", err)
	}
	x, _ := testRedisCache.Get("x")
	if x != "y" {
		t.Error("no value retrieved")
	}

	err = testRedisCache.Empty()
	if err != nil {
		t.Error("error emptying cache", err)
	}

	x, _ = testRedisCache.Get("x")
	if x != nil {
		t.Error("unexpected value retrieved", x)
	}
}

func TestClose(t *testing.T) {
	err := testRedisCache.Close()
	if err != nil {
		t.Error(err)
	}
}
