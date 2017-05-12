package main

import (
	"bytes"
	"golang.org/x/text/transform"
	"os"
	"strings"
	"testing"
)

var csvtests = []struct {
	in  string
	out string
}{
	{"1,2,3", "1,2,3"},                                                          // Complete pass through
	{`1," commas, in, the, middle,",3`, `1," commas, in, the, middle,",3`},      // Passthrough with enclosed commas
	{`1," commas, " in, the, middle,",3`, `1," commas, "" in, the, middle,",3`}, // Unescaped double quotes
	{`1,string "  end,3`, `1,"string ""  end",3`},                               // escaping found in the middle
	{`1,"string ,\"  end",3`, `1,"string ,""  end",3`},                          // correcting slash escaping
	{`1,"string \",  end",3`, `1,"string "",  end",3`},                          // correcting slash followed by ignorable comma
	{`1,2",3`, `1,"2""",3`},                                                     // Missing start quote
	{`1,2,3"`, `1,2,"3"""`},                                                     // Missing quote at end
	{`1,2,"
			3"`, `1,2,"
			3"`}, // Newlines should pass through
	{`1,2,this newline should "have a quote" terminator
			4,5,6`, `1,2,"this newline should ""have a quote"" terminator"
			4,5,6`}, // Newlines should terminate quoted lines
	{`1,"2"
			3,"4"`, `1,"2"
			3,"4"`}, // double quote then newline should trigger end of quoted entry
	{`6,"work",o "a". `, `6,"work","o ""a"". "`}, // end of file should also have a quote if missing
	{`626,"Mon",monopoly
			629,"Train",Can 'test" tiny?
			656,"Clean."`,
		`626,"Mon",monopoly
			629,"Train","Can 'test"" tiny?"
			656,"Clean."`},
}

func TestCsvCorrector(t *testing.T) {
	for _, tt := range csvtests {
		var buf bytes.Buffer
		err := Clean(strings.NewReader(tt.in), &buf)
		result := buf.String()
		if err != nil {
			t.Error("Unexpected error: ", err)
		}
		if result != tt.out {
			t.Errorf("Expected '%s' but got '%s'", tt.out, result)
		}
	}
}

func TestTransformSmallBuffer(t *testing.T) {
	cleaner := NewCleaner()

	for _, size := range []int{1, 2} {
		for _, tt := range csvtests {
			in := []byte(tt.in)
			out := make([]byte, len(in)*2)
			outPos := 0

			var lastErr error
			var i, outMove, inRead int
			for {
				if i >= len(in) {
					break
				}
				end := false
				if i == len(in)-1 {
					end = true
				}
				inEnd := i + size
				if inEnd > len(in) {
					inEnd = len(in)
				}
				outMove, inRead, lastErr = cleaner.Transform(out[outPos:outPos+size], in[i:inEnd], end)
				outPos += outMove
				i += inRead
			}
			if lastErr == transform.ErrShortSrc {
				outMove, inRead, lastErr = cleaner.Transform(out[outPos:outPos+size], nil, true)
				if outMove == 0 {
					t.Error("Didn't print into the buffer")
				}

				outPos += outMove
			}

			for lastErr == transform.ErrShortDst {
				outMove, inRead, lastErr = cleaner.Transform(out[outPos:outPos+size], nil, true)
				if outMove == 0 {
					t.Error("Didn't print into the buffer")
				}
				outPos += outMove
			}

			result := string(out[:outPos])
			if result != tt.out {
				t.Errorf("Expected '%s' but got '%s' with buffer size %d (pos:%d, len:%d, expectedLen:%d)", tt.out, result, size, outPos, len(result), len(tt.out))
			}
		}
		cleaner.Reset()
	}
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
