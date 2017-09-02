package rtf2txt

import (
	"bytes"
	"io"
	"strings"
	"time"
)

// Text is used to convert an io.Reader containing RTF data into
// plain text
func Text(r io.Reader) (io.Reader, error) {
	buf := make([]byte, 512)

	var text bytes.Buffer
	var symbol bytes.Buffer
	var isSymbol bool
	var isControl bool
	var isText bool
	var hasCR bool
	var lastByte byte
	var lastSymbol string
	var symbolStack stack
	var level int
	var controlLevel int
	for bytesRead, err := r.Read(buf); bytesRead != 0; bytesRead, err = r.Read(buf) {
		if err != nil {
			return nil, err
		}

		for i := range buf {
			c := buf[i]
			switch c {
			case '\\':
				isText = false
				isSymbol = true
				lastSymbol = symbol.String()
				symbol = bytes.Buffer{}
			case '{':
				isText = false
				lastSymbol = symbol.String()
				symbolStack.Push(lastSymbol)
				level++
			case '}':
				isText = false
				isSymbol = false
				if isControl && level == controlLevel {
					lastSymbol = symbolStack.Pop()
					isControl = false
				}
				level--
			case '\n':
				if isText || isTextSymbol(lastSymbol) {
					hasCR = true
					continue // skip the hasCR = false at end
				}
			case '\r': // noop
			default:
				isText = false
				if c == '*' && lastByte == '\\' { // this is a control sequence
					isControl = true
					controlLevel = level
				} else if (isSymbol || isControl) && c == ' ' {
					isSymbol = false
					lastSymbol = symbol.String()
					symbol = bytes.Buffer{}
					insertSpecialCharacter(lastSymbol, text)
				} else if !isSymbol && !isControl && !isSkippedSymbol(lastSymbol) {
					if hasCR { // CR followed by more text. Don't write out CR
						hasCR = false
					}
					isText = true
					text.WriteByte(c)
				} else if isSymbol {
					symbol.WriteByte(c)
				}
			}
			if hasCR {
				text.WriteByte('\n')
				hasCR = false
			}
			lastByte = c
		}
	}
	return &text, nil
}

func unicode(symbol string) string {
	return ""
}

func tablestuff(symbol string) string {
	return ""
}

func headerAndFooter(symbol string) string {
	return ""
}

func insertSpecialCharacter(symbol string, text bytes.Buffer) {
	switch symbol {
	case "chdate":
		text.WriteString(time.Now().String())
	case "chldpl":
	case "chdpa":
	case "chtime":
	case "chpgn":
	case "sectnum":
	case "chftn":
	case "chatn":
	case "chftnsep":
	case "chftnsepc":
	case "cell":
	case "nestcell":
	case "row":
	case "nestrow":
	case "par":
	case "sect":
	case "page":
	case "column":
	case "line":
	case "lbrN":
	case "softpage":
	case "softcol":
	case "softline":
	case "softlheightN":
	case "tab":
	case "emdash", "endash":
		text.WriteByte('-')
	case "emspace", "enspace", "qmspace":
		text.WriteByte(' ')
	case "bullet":
		text.WriteByte('*')
	case "lquote", "rquote":
		text.WriteByte('\'')
	case "ldblquote", "rdblquote":
		text.WriteByte('"')
	case "|":
	case "~":
	case "-":
	case "_":
	case ":":
	case "*":
	case "'hh":
	case "ltrmark":
	case "rtlmark":
	case "zwbo":
	case "zwnbo":
	case "zwj":
	case "zwnj":

	}
}

func isTextSymbol(symbol string) bool {
	return strings.HasPrefix(symbol, "insrsid") || strings.HasPrefix(symbol, "charrsid")
}

func isSkippedSymbol(symbol string) bool {
	return strings.HasPrefix(symbol, "fprq") || strings.HasPrefix(symbol, "spriority") || symbol == "sqformat" ||
		strings.HasPrefix(symbol, "snext") || strings.HasPrefix(symbol, "styrs") || strings.HasPrefix(symbol, "slink") ||
		symbol == "title" || symbol == "author" || symbol == "operator"
}
