package main

import (
	"bytes"
	"io"
	"os"
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

func TestDelete(t *testing.T) {
	opts := StoreOpts{
		PathTransformFunc: CASPathTransformFunc,
	}
	s := NewStore(opts)
	key := "momsspecialspicture"
	data := []byte("somejpegbytes")

	if err := s.writeStream(key, bytes.NewReader(data)); err != nil {
		t.Fatal(err)
	}

	if err := s.Delete(key); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(CASPathTransformFunc(key).FullPath()); !os.IsNotExist(err) {
		t.Fatal(err)
	}
}

func TestStore(t *testing.T) {
	opts := StoreOpts{
		PathTransformFunc: CASPathTransformFunc,
	}
	s := NewStore(opts)
	key := "momsspecialspicture"
	data := []byte("somejpegbytes")

	if err := s.writeStream(key, bytes.NewReader(data)); err != nil {
		t.Fatal(err)
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

}
