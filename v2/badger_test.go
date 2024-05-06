package remember

import (
	"encoding/gob"
	"testing"
	"time"
)

func TestBadgerCache_Has(t *testing.T) {
	err := testBadgerCache.Forget("foo")
	if err != nil {
		t.Error(err)
	}

	inCache := testBadgerCache.Has("foo")
	if err != nil {
		t.Error(err)
	}

	if inCache {
		t.Error("foo found in cache, and it shouldn't be there")
	}

	_ = testBadgerCache.Set("foo", "bar")
	inCache = testBadgerCache.Has("foo")
	if err != nil {
		t.Error(err)
	}

	if !inCache {
		t.Error("foo not found in cache")
	}

	err = testBadgerCache.Forget("foo")
	if err != nil {
		t.Error(err)
	}
}

func TestBadgerCache_Get(t *testing.T) {
	err := testBadgerCache.Set("foo", "bar")
	if err != nil {
		t.Error(err)
	}

	x, err := testBadgerCache.Get("foo")
	if err != nil {
		t.Error(err)
	}

	if x != "bar" {
		t.Error("did not get correct value from cache")
	}

	err = testBadgerCache.Set("foo2", "bar", time.Second)
	if err != nil {
		t.Error(err)
	}

	time.Sleep(2 * time.Second)

	if testBadgerCache.Has("foo2") {
		t.Error("cache has foo2 and it should have expired")
	}

	testBadgerCache.Empty()
}

func TestBadgerCache_Forget(t *testing.T) {
	err := testBadgerCache.Set("foo", "foo")
	if err != nil {
		t.Error(err)
	}

	err = testBadgerCache.Forget("foo")
	if err != nil {
		t.Error(err)
	}

	inCache := testBadgerCache.Has("foo")
	if err != nil {
		t.Error(err)
	}

	if inCache {
		t.Error("foo found in cache, and it shouldn't be there")
	}

}

func TestBadgerCache_Empty(t *testing.T) {
	err := testBadgerCache.Set("alpha", "beta")
	if err != nil {
		t.Error(err)
	}

	err = testBadgerCache.Empty()
	if err != nil {
		t.Error(err)
	}

	inCache := testBadgerCache.Has("alpha")
	if err != nil {
		t.Error(err)
	}

	if inCache {
		t.Error("alpha found in cache, and it shouldn't be there")
	}
}

func TestBadgerCache_EmptyByMatch(t *testing.T) {
	err := testBadgerCache.Set("alpha", "beta")
	if err != nil {
		t.Error(err)
	}

	err = testBadgerCache.Set("alpha2", "beta2")
	if err != nil {
		t.Error(err)
	}

	err = testBadgerCache.Set("beta", "beta")
	if err != nil {
		t.Error(err)
	}

	err = testBadgerCache.EmptyByMatch("a")
	if err != nil {
		t.Error(err)
	}

	inCache := testBadgerCache.Has("alpha")
	if err != nil {
		t.Error(err)
	}

	if inCache {
		t.Error("alpha found in cache, and it shouldn't be there")
	}

	inCache = testBadgerCache.Has("alpha2")
	if err != nil {
		t.Error(err)
	}

	if inCache {
		t.Error("alpha2 found in cache, and it shouldn't be there")
	}

	inCache = testBadgerCache.Has("beta")
	if err != nil {
		t.Error(err)
	}

	if !inCache {
		t.Error("beta not found in cache, and it should be there")
	}
}

func TestBadgerCache_GetInt(t *testing.T) {

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
			testBadgerCache.Set(tt.key, tt.data)
		}

		x, err := testBadgerCache.GetInt(tt.key)
		if !tt.errorExpected && err != nil {
			t.Errorf("%s: received unexpected error: %s", tt.name, err.Error())
		}

		if !tt.errorExpected && x != tt.data {
			t.Errorf("%s: wrong value retrieved from cache; expected %d but got %d", tt.name, tt.data, x)
		}
	}
	testBadgerCache.Empty()
}

func TestBadgerCache_GetString(t *testing.T) {

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
			testBadgerCache.Set(tt.key, tt.data)
		}

		x, err := testBadgerCache.GetString(tt.key)
		if !tt.errorExpected && err != nil {
			t.Errorf("%s: received unexpected error: %s", tt.name, err.Error())
		}

		if !tt.errorExpected && x != tt.data {
			t.Errorf("%s: wrong value retrieved from cache; expected %s but got %s", tt.name, tt.data, x)
		}
	}
	testBadgerCache.Empty()
}

func TestBadgerCache_GetTime(t *testing.T) {
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
			err := testBadgerCache.Set(tt.key, tt.data)
			if err != nil {
				t.Error(err)
			}
		}

		x, err := testBadgerCache.GetTime(tt.key)
		if !tt.errorExpected && err != nil {
			t.Errorf("%s: received unexpected error: %s", tt.name, err.Error())
		}

		if !tt.errorExpected && !tt.data.Equal(x) {
			t.Errorf("%s: wrong value retrieved from cache; expected %s but got %s", tt.name, tt.data, x)
		}
	}
	testBadgerCache.Empty()
}

func TestBadgerCache_Close(t *testing.T) {
	err := testBadgerCache.Close()
	if err != nil {
		t.Error(err)
	}
}
