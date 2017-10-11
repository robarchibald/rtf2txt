package rtf2txt

import (
	"os"
	"testing"

	"github.com/EndFirstCorp/peekingReader"
)

func Test2Text(t *testing.T) {
	f, _ := os.Open(`testdata/np.new.rtf`)
	r, err := Text(f)
	if err != nil {
		t.Error(err)
	}
	f.Close()
	s := r.String()
	if s != txt {
		t.Error("doesn't match", s)
	}

	f, _ = os.Open(`testdata/ad.rtf`)
	r, err = Text(f)
	if err != nil {
		t.Error(err)
	}
	f.Close()
	s = r.String()
	if s != red {
		t.Error("doesn't match", s)
	}
}

const txt = `Of course, we frequently hear about larger brands pushing out a ton of amazing content, and they're often used as examples of how to do content right. `
const red = `Restore The Selling Balance. Ad Technology doesn't have to be faceless. Our platform is designed to connect media companies directly to advertisers.`

func TestTokenizeControl(t *testing.T) {
	r := peekingReader.NewMemReader([]byte("f463 re often"))
	control, num, _ := tokenizeControl(r)
	if control != "fN" || num != 463 {
		t.Error("expected valid control", control, num)
	}
	// check position of reader
	b, _ := r.ReadBytes(9)
	if string(b) != " re often" {
		t.Error("expected remaining string", string(b))
	}
}
