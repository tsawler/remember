package remember

import (
	"encoding/gob"
	"testing"
	"time"
)

func TestBuntdbCache_Has(t *testing.T) {
	err := testBuntdbCache.Forget("foo")
	if err != nil {
		t.Error(err)
	}

	inCache := testBuntdbCache.Has("foo")
	if err != nil {
		t.Error(err)
	}

	if inCache {
		t.Error("foo found in cache, and it shouldn't be there")
	}

	_ = testBuntdbCache.Set("foo", "bar")
	inCache = testBuntdbCache.Has("foo")
	if err != nil {
		t.Error(err)
	}

	if !inCache {
		t.Error("foo not found in cache")
	}

	err = testBuntdbCache.Forget("foo")
	if err != nil {
		t.Error(err)
	}
}

func TestBuntdbCache_Get(t *testing.T) {
	err := testBuntdbCache.Set("foo", "bar")
	if err != nil {
		t.Error(err)
	}

	x, err := testBuntdbCache.Get("foo")
	if err != nil {
		t.Error(err)
	}

	if x != "bar" {
		t.Error("did not get correct value from cache")
	}

	err = testBuntdbCache.Set("foo2", "bar", time.Second)
	if err != nil {
		t.Error(err)
	}

	time.Sleep(2 * time.Second)

	if testBuntdbCache.Has("foo2") {
		t.Error("cache has foo2 and it should have expired")
	}

	testBuntdbCache.Empty()
}

func TestBuntdbCache_Forget(t *testing.T) {
	err := testBuntdbCache.Set("foo", "foo")
	if err != nil {
		t.Error(err)
	}

	err = testBuntdbCache.Forget("foo")
	if err != nil {
		t.Error(err)
	}

	inCache := testBuntdbCache.Has("foo")
	if err != nil {
		t.Error(err)
	}

	if inCache {
		t.Error("foo found in cache, and it shouldn't be there")
	}

}

func TestBuntdbCache_Empty(t *testing.T) {
	err := testBuntdbCache.Set("alpha", "beta")
	if err != nil {
		t.Error(err)
	}

	err = testBuntdbCache.Empty()
	if err != nil {
		t.Error(err)
	}

	inCache := testBuntdbCache.Has("alpha")
	if err != nil {
		t.Error(err)
	}

	if inCache {
		t.Error("alpha found in cache, and it shouldn't be there")
	}
}

func TestBuntdbCache_EmptyByMatch(t *testing.T) {
	err := testBuntdbCache.Set("alpha", "beta")
	if err != nil {
		t.Error(err)
	}

	err = testBuntdbCache.Set("alpha2", "beta2")
	if err != nil {
		t.Error(err)
	}

	err = testBuntdbCache.Set("beta", "beta")
	if err != nil {
		t.Error(err)
	}

	err = testBuntdbCache.EmptyByMatch("a")
	if err != nil {
		t.Error(err)
	}

	inCache := testBuntdbCache.Has("alpha")
	if err != nil {
		t.Error(err)
	}

	if inCache {
		t.Error("alpha found in cache, and it shouldn't be there")
	}

	inCache = testBuntdbCache.Has("alpha2")
	if err != nil {
		t.Error(err)
	}

	if inCache {
		t.Error("alpha2 found in cache, and it shouldn't be there")
	}

	inCache = testBuntdbCache.Has("beta")
	if err != nil {
		t.Error(err)
	}

	if !inCache {
		t.Error("beta not found in cache, and it should be there")
	}
}

func TestBuntdbCache_GetInt(t *testing.T) {

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
			testBuntdbCache.Set(tt.key, tt.data)
		}

		x, err := testBuntdbCache.GetInt(tt.key)
		if !tt.errorExpected && err != nil {
			t.Errorf("%s: received unexpected error: %s", tt.name, err.Error())
		}

		if !tt.errorExpected && x != tt.data {
			t.Errorf("%s: wrong value retrieved from cache; expected %d but got %d", tt.name, tt.data, x)
		}
	}
	testBuntdbCache.Empty()
}

func TestBuntdbCache_GetString(t *testing.T) {

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
			testBuntdbCache.Set(tt.key, tt.data)
		}

		x, err := testBuntdbCache.GetString(tt.key)
		if !tt.errorExpected && err != nil {
			t.Errorf("%s: received unexpected error: %s", tt.name, err.Error())
		}

		if !tt.errorExpected && x != tt.data {
			t.Errorf("%s: wrong value retrieved from cache; expected %s but got %s", tt.name, tt.data, x)
		}
	}
	testBuntdbCache.Empty()
}

func TestBuntdbCache_GetTime(t *testing.T) {
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
			err := testBuntdbCache.Set(tt.key, tt.data)
			if err != nil {
				t.Error(err)
			}
		}

		x, err := testBuntdbCache.GetTime(tt.key)
		if !tt.errorExpected && err != nil {
			t.Errorf("%s: received unexpected error: %s", tt.name, err.Error())
		}

		if !tt.errorExpected && !tt.data.Equal(x) {
			t.Errorf("%s: wrong value retrieved from cache; expected %s but got %s", tt.name, tt.data, x)
		}
	}
	testBuntdbCache.Empty()
}

func TestBuntdbCache_Close(t *testing.T) {
	err := testBuntdbCache.Close()
	if err != nil {
		t.Error(err)
	}
}
