package main

import (
	"bytes"
	"testing"
)

func TestPathTransformFunc(t *testing.T) {
	key := "momsbestpicture"
	pathname := CASPathTransformFunc(key)
	expected := "/68044/29f74/181a6/3c50c/3d81d/733a1/2f14a/353ff"
	if pathname != expected {
		t.Errorf("Expected %s, got %s", expected, pathname)
	}
}

func TestStore(t *testing.T) {
	opts := StoreOpts{
		PathTransformFunc: DefaultPathTransformFunc,
	}
	s := NewStore(opts)

	data := bytes.NewReader([]byte("somejpegbytes"))
	if err := s.writeStream("somespecialpicture", data); err != nil {
		t.Fatal(err)
	}
}
