package remember

import (
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
