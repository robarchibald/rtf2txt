package rtf2txt

import (
	"bytes"
	"os"
	"testing"
)

func Test2Text(t *testing.T) {
	f, _ := os.Open(`testdata/np.new.rtf`)
	r, err := Text(f)
	if err != nil {
		t.Error(err)
	}
	s := r.(*bytes.Buffer).String()
	if s != txt {
		t.Error("doesn't match", s)
	}
}

const txt = `Of course, we frequently hear about larger brands pushing out a ton of amazing content, and they're often used as examples of how to do content right.`
