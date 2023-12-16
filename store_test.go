package main

import (
	"bytes"
	"fmt"
	"io"
	"testing"
)

func TestPathTransformFunc(t *testing.T) {
	key := "momsbestpicture"
	pathKey := CASPathTransformFunc(key)
	expectedOriginalKey := "6804429f74181a63c50c3d81d733a12f14a353ff"
	expectedPathName := "68044/29f74/181a6/3c50c/3d81d/733a1/2f14a/353ff"
	if pathKey.Pathname != expectedPathName {
		t.Errorf("Expected %s, got %s", pathKey.Pathname, expectedPathName)
	}
	if pathKey.Filename != expectedOriginalKey {
		t.Errorf("Expected %s, got %s", pathKey.Filename, expectedOriginalKey)
	}
}

func TestStore(t *testing.T) {
	s := newStore()
	defer teardown(t, s)

	for i := 0; i < 50; i++ {

		key := fmt.Sprintf("foo_%d", i)
		data := []byte("some jpg bytes")

		if err := s.writeStream(key, bytes.NewReader(data)); err != nil {
			t.Fatal(err)
		}

		if ok := s.Has(key); !ok {
			t.Errorf("Expected %s to exist", key)
		}

		r, err := s.Read(key)
		if err != nil {
			t.Fatal(err)
		}

		b, err := io.ReadAll(r)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(b, data) {
			t.Errorf("Expected %s, got %s", string(data), string(b))
		}

		if err := s.Delete(key); err != nil {
			t.Fatal(err)
		}

		if ok := s.Has(key); ok {
			t.Errorf("Expected %s to not exist", key)
		}
	}

}

func newStore() *Store {
	opts := StoreOpts{
		PathTransformFunc: CASPathTransformFunc,
	}
	return NewStore(opts)
}

func teardown(t *testing.T, s *Store) {
	if err := s.Clear(); err != nil {
		t.Error(err)
	}
}
