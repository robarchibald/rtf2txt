package rtf2txt

import (
	"bytes"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/EndFirstCorp/peekingReader"
)

const txt = `Of course, we frequently hear about larger brands pushing out a ton of amazing content, and they're often used as examples of how to do content right. `
const red = `Restore The Selling Balance. Ad Technology doesn't have to be faceless. Our platform is designed to connect media companies directly to advertisers.`

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

	mr := peekingReader.NewMemReader([]byte("\\\\hello\\\\there\\{\\}friends"))
	if r, err := Text(mr); err != nil || r.String() != "\\hello\\there{}friends" {
		t.Error("expected success", err, r.String())
	}
}

func TestReadControl(t *testing.T) {
	var s stack
	var text bytes.Buffer
	r := peekingReader.NewMemReader([]byte(""))
	if err := readControl(r, &s, &text); err != io.EOF {
		t.Error("expected error", err)
	}

	// no closing brace. Should error
	r = peekingReader.NewMemReader([]byte("*\rsidtbl \rs"))
	if err := readControl(r, &s, &text); err != io.EOF {
		t.Error("expected error", err)
	}

	// no parameters found, no previous control to get params for
	r = peekingReader.NewMemReader([]byte("*\rsidtbl \rs}"))
	if err := readControl(r, &s, &text); err != nil {
		t.Error("expected success", err)
	}

	// unicode
	r = peekingReader.NewMemReader([]byte("'A9 "))
	if err := readControl(r, &s, &text); err != nil || text.String() != "©" {
		t.Error("expected success", err, text.String())
	}
	text.Reset()

	r = peekingReader.NewMemReader([]byte("\\\\"))
	if err := readControl(r, &s, &text); err != nil || text.String() != "\\" {
		t.Error("expected success", err, text.String())
	}
	text.Reset()

	// carriage return
	r = peekingReader.NewMemReader([]byte(`
`))
	if err := readControl(r, &s, &text); err != nil || text.String() != "\n" {
		t.Error("expected success", err, text.String())
	}
	text.Reset()

	// binary data error
	r = peekingReader.NewMemReader([]byte(`bin412`))
	if err := readControl(r, &s, &text); err != io.EOF {
		t.Error("expected success", err, text.String())
	}

	// binary data success
	r = peekingReader.NewMemReader([]byte(`bin22 1234567890123456789012 hello}`))
	if err := readControl(r, &s, &text); err != nil || text.String() != "" {
		t.Error("expected success", err, text.String())
	}

	// truncated parameter
	r = peekingReader.NewMemReader([]byte(`f463 hi`))
	if err := readControl(r, &s, &text); err != io.EOF {
		t.Error("expected success", err, text.String())
	}
}

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

type errorAtReader struct {
	ErrorAfter int
}

func (e *errorAtReader) Peek(num int) ([]byte, error) {
	if e.ErrorAfter <= 0 {
		return nil, errors.New("failed")
	}
	e.ErrorAfter--
	return []byte("hi"), nil
}

func (e *errorAtReader) Read(p []byte) (int, error) {
	if e.ErrorAfter <= 0 {
		return -1, errors.New("failed")
	}
	e.ErrorAfter--
	return 0, nil
}

func (e *errorAtReader) ReadByte() (byte, error) {
	if e.ErrorAfter <= 0 {
		return '\x00', errors.New("failed")
	}
	e.ErrorAfter--
	return '\x00', nil
}

func (e *errorAtReader) ReadBytes(num int) ([]byte, error) {
	if e.ErrorAfter <= 0 {
		return nil, errors.New("failed")
	}
	e.ErrorAfter--
	return []byte("hi"), nil
}

func (e *errorAtReader) ReadRune() (rune, int, error) {
	if e.ErrorAfter <= 0 {
		return 0, 0, errors.New("failed")
	}
	e.ErrorAfter--
	return 0, 0, nil
}
