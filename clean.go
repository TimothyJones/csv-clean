package main

import (
	"bufio"
	"golang.org/x/text/transform"
	"io"
)

// Clean corrects the CSV provided by the reader to RFC-4180 format
func Clean(r io.Reader, w io.Writer) error {
	reader := transform.NewReader(bufio.NewReader(r), NewCleaner())
	_, err := io.Copy(w, reader)

	return err
}

type state int

const (
	uninitalised = iota
	betweenFields
	inUnquoted
	inQuoted
	inQuotedEnding
	inQuotedFoundSlash
	writingBuffer
)

// CSVCleaner is a transformer that cleans up CSV file into RFC-4180 format
type CSVCleaner struct {
	state                           state
	buf                             []byte
	bufpos, dstpos, keepWritingFrom int
	addQuote                        bool
}

// NewCleaner returns a new CSVCleaner
func NewCleaner() *CSVCleaner {
	return &CSVCleaner{}
}

func (c *CSVCleaner) startBuf() {
	if c.buf == nil {
		c.buf = make([]byte, 4096)
	}
	c.buf = c.buf[:0]
	c.bufpos = 0
	c.addQuote = false
	c.keepWritingFrom = 0
	c.state = betweenFields
}

func (c *CSVCleaner) finish(dst []byte) error {

	writeLen := c.bufpos - c.keepWritingFrom

	if c.addQuote {
		writeLen++
	}

	if c.dstpos+writeLen > len(dst) {
		writeLen = len(dst) - c.dstpos
	}

	if c.addQuote {
		dst[c.dstpos] = '"'
		c.dstpos++
		writeLen--
	}

	copy(dst[c.dstpos:], c.buf[c.keepWritingFrom:c.keepWritingFrom+writeLen])
	c.dstpos += writeLen
	if c.keepWritingFrom+writeLen == c.bufpos {
		// We wrote everything so we can start again
		c.startBuf()
		return nil
	}
	c.keepWritingFrom = c.keepWritingFrom + writeLen
	c.state = writingBuffer
	c.addQuote = false
	return transform.ErrShortDst
}

func (c *CSVCleaner) quotePrevious() {
	c.buf = append(c.buf, '"')
	c.buf[c.bufpos], c.buf[c.bufpos-1] = c.buf[c.bufpos-1], c.buf[c.bufpos]
	c.bufpos++
}

func (c *CSVCleaner) add(b byte) {
	c.buf = append(c.buf, b)
	c.bufpos++
}

func (c *CSVCleaner) correct(b byte) {
	c.buf[c.bufpos-1] = b
}

// Transform transforms the incoming byte slice into the output slice.
// It fulfils the contract described on transform.Transformer.
func (c *CSVCleaner) Transform(dst, src []byte, atEOF bool) (int, int, error) {
	c.dstpos = 0

	//log.Println("[ ] At transform ", len(dst), len(src), atEOF)
	if c.state == uninitalised {
		c.startBuf()
		c.state = betweenFields
	}
	if c.state == writingBuffer {
		err := c.finish(dst)
		if err != nil {
			//log.Println("[ ] We're returning ", c.dstpos, 0, err)
			return c.dstpos, 0, err
		}
	}

	for i, b := range src {
		switch c.state {
		case betweenFields:
			if b == '"' {
				c.state = inQuoted
			} else {
				c.state = inUnquoted
			}
		case inQuoted:
			switch b {
			case '"':
				c.state = inQuotedEnding
			case '\\':
				c.state = inQuotedFoundSlash
			case ',':
				if c.addQuote {
					// we saw a terminating comma when we added a quote due to unescaped quoting
					c.add('"')

					c.state = betweenFields
				}
			case '\n':
				if c.addQuote {
					// Newline should terminate the input if we were originally unquoted
					c.add('"')
					c.state = betweenFields
				}
			}
		case inUnquoted:
			switch b {
			case '"':
				c.addQuote = true
				c.state = inQuotedEnding
			case ',', '\n':
				// Saw the terminating comma or newline
				c.state = betweenFields
			}
		case inQuotedFoundSlash:
			switch b {
			case '"':
				// saw a badly escaped double quote
				c.correct('"')
				c.state = inQuoted
			}
		case inQuotedEnding:
			switch b {
			case '"':
				// saw a correctly escaped double quote
				c.state = inQuoted
			case ',', '\n':
				// saw the terminating double quote
				if c.addQuote {
					c.add('"')
					c.add('"')
				}

				c.state = betweenFields
			default:
				// we saw a double quote but it wasn't escaped
				c.add('"')
				c.state = inQuoted
			}
		}
		c.add(b)

		if c.state == betweenFields {
			err := c.finish(dst)
			if err != nil {
				return c.dstpos, i + 1, err
			}
		}
	}
	var err error
	if atEOF {
		if c.addQuote {
			switch c.state {
			case inQuoted:
				c.add('"')
			case inQuotedEnding:
				c.add('"')
				c.add('"')
			}
		}
		err = c.finish(dst)
		if err != nil {
			return c.dstpos, len(src), err
		}
	}
	if c.dstpos == 0 && len(src) != 0 {
		return 0, len(src), transform.ErrShortSrc
	}
	return c.dstpos, len(src), nil
}

// Reset resets the state and allows a Transformer to be reused.
func (c *CSVCleaner) Reset() {
	c.state = betweenFields
	c.buf = nil
	c.bufpos, c.dstpos = 0, 0
	c.addQuote = false
}
