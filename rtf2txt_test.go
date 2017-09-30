package rtf2txt

import (
	"bytes"
	"testing"
	"os"
)

func Test2Text(t *testing.T) {
	f, _ := os.Open(`testdata/np.new.rtf`)
	r, err := Text(f)
	if err != nil {
		t.Error(err)
	}
	s := r.(*bytes.Buffer).String()
	if s != "hello" {
		t.Error("doesn't match", s)
	}
}