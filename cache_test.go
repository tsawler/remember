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
			key:           "v1",
			data:          "bar",
			expires:       time.Microsecond,
			wait:          time.Millisecond,
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
}
