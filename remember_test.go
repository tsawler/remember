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
	c1 := New(ops)

	err := c1.Set("foo", "bar")
	if err != nil {
		t.Error("error", err)
	}

	c2 := New()
	err = c2.Set("foo", "bar")
	if err != nil {
		t.Error("error", err)
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
		err := testCache.Set(tt.key, tt.data, tt.expires)
		if err != nil && !tt.errorExpected {
			t.Errorf("%s: received unexpected error: %s", tt.name, err.Error())
		}

		if tt.wait != 0 && tt.expires != 0 {
			time.Sleep(tt.wait)
		}

		retrieved, err := testCache.Get(tt.key)
		if !tt.errorExpected && retrieved != tt.data {
			t.Errorf("%s: incorrect value retrieved; got %s expected %s", tt.name, retrieved, tt.data)
		}

	}

	testCache.Empty()
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
			testCache.Set(tt.key, tt.data)
		}

		x, err := testCache.GetInt(tt.key)
		if !tt.errorExpected && err != nil {
			t.Errorf("%s: received unexpected error: %s", tt.name, err.Error())
		}

		if !tt.errorExpected && x != tt.data {
			t.Errorf("%s: wrong value retrieved from cache; expected %d but got %d", tt.name, tt.data, x)
		}
	}
	testCache.Empty()
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
			testCache.Set(tt.key, tt.data)
		}

		x, err := testCache.GetString(tt.key)
		if !tt.errorExpected && err != nil {
			t.Errorf("%s: received unexpected error: %s", tt.name, err.Error())
		}

		if !tt.errorExpected && x != tt.data {
			t.Errorf("%s: wrong value retrieved from cache; expected %s but got %s", tt.name, tt.data, x)
		}
	}
	testCache.Empty()
}

func TestForget(t *testing.T) {
	_ = testCache.Set("forgetme", "x")
	err := testCache.Forget("forgetme")
	if err != nil {
		t.Error(err)
	}

	_, err = testCache.Get("forgetme")
	if err == nil {
		t.Error("got something from the cache, and it should not be there")
	}
	testCache.Empty()
}

func TestHas(t *testing.T) {
	_ = testCache.Set("fish", "x")
	if !testCache.Has("fish") {
		t.Error("should have value in cache, but do not")
	}

	_ = testCache.Forget("fish")
	if testCache.Has("fish") {
		t.Error("value should not exist in cache, but it does")
	}

	testCache.Empty()
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
			err := testCache.Set(tt.key, tt.data)
			if err != nil {
				t.Error(err)
			}
		}

		x, err := testCache.GetTime(tt.key)
		if !tt.errorExpected && err != nil {
			t.Errorf("%s: received unexpected error: %s", tt.name, err.Error())
		}

		if !tt.errorExpected && !tt.data.Equal(x) {
			t.Errorf("%s: wrong value retrieved from cache; expected %s but got %s", tt.name, tt.data, x)
		}
	}
	testCache.Empty()
}

func TestEmptyByMatch(t *testing.T) {
	_ = testCache.Set("fooa", "bar")
	_ = testCache.Set("foob", "bar")
	_ = testCache.Set("fooc", "bar")

	err := testCache.EmptyByMatch("foo")
	if err != nil {
		t.Error(err)
	}

	if testCache.Has("fooa") ||
		testCache.Has("foob") ||
		testCache.Has("fooc") {
		t.Error("cache still has values beginning with foo")
	}

	testCache.Empty()
}

func TestEmpty(t *testing.T) {
	err := testCache.Set("x", "y")
	if err != nil {
		t.Error("error setting value in cache", err)
	}
	x, _ := testCache.Get("x")
	if x != "y" {
		t.Error("no value retrieved")
	}

	err = testCache.Empty()
	if err != nil {
		t.Error("error emptying cache", err)
	}

	x, _ = testCache.Get("x")
	if x != nil {
		t.Error("unexpected value retrieved", x)
	}
}
