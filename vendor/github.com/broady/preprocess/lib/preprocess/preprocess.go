package preprocess

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"
	"unicode"
)

func Process(in io.Reader, flags []string, prefix string) ([]byte, error) {
	flagMap := make(map[string]bool)
	for _, f := range flags {
		flagMap[f] = true
	}
	flagMap["true"] = true // always

	var buf bytes.Buffer

	s := &scanner{
		r:            bufio.NewReader(in),
		templates:    make(map[string][]byte),
		replacements: make(map[string][]byte),
		prefix:       []byte(prefix),
		flags:        flagMap,
		w:            &buf,
	}
	if err := s.start(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type scanner struct {
	templates    map[string][]byte
	replacements map[string][]byte
	flags        map[string]bool
	prefix       []byte

	state []state

	w io.Writer

	lineNumber int
	r          *bufio.Reader
	eof        bool
}

func (s *scanner) start() (err error) {
	defer func() {
		if r := recover(); r != nil {
			//debug.PrintStack()
			e, ok := r.(error)
			if ok {
				err = e
			} else {
				panic(r)
			}
		}
	}()

	s.pushState(s.neutral)

	for !s.eof {
		s.state[len(s.state)-1](s.w, s.line())
	}

	return nil
}

func (s *scanner) setState(t state) {
	s.state[len(s.state)-1] = t
}

func (s *scanner) pushState(t state) {
	s.state = append(s.state, t)
}

func (s *scanner) popState() {
	if len(s.state) == 1 {
		s.error("unexpected end directive")
	}
	s.state = s.state[:len(s.state)-1]
}

func (s *scanner) error(e string) {
	//debug.PrintStack()
	// TODO(cbro): properly handle line numbers for defs.
	panic(fmt.Errorf("%d: %s", s.lineNumber, e))
}

func (s *scanner) line() []byte {
	l, err := s.r.ReadBytes('\n')
	if err != nil && err != io.EOF {
		panic(err)
	}
	if err == io.EOF {
		s.eof = true
	}
	s.lineNumber++
	return l
}

func (s *scanner) applyReplacements(line []byte) []byte {
	// TODO(cbro): stable iteration order.
	for sentinel, replacement := range s.replacements {
		line = bytes.Replace(line, []byte(sentinel), replacement, -1)
	}
	return line
}

// State functions

type state func(io.Writer, []byte)

func (s *scanner) neutral(w io.Writer, line []byte) {
	if !bytes.Contains(line, s.prefix) {
		w.Write(s.applyReplacements(line))
		return
	}

	before, pragma := s.splitPrefix(line)
	switch pragma[0] {
	case "end":
		s.popState()
	case "def":
		s.pushState(s.def(pragma[1]))
	case "if":
		// TODO(cbro): support full boolean expressions.
		if s.flags[pragma[1]] {
			s.pushState(s.neutral)
		} else if strings.HasPrefix(pragma[1], "!") && !s.flags[pragma[1][1:]] {
			s.pushState(s.neutral)
		} else {
			s.pushState(s.consumeUntilEnd)
		}
	case "omit":
		if len(pragma) > 2 && pragma[1] == "if" {
			if s.flags[pragma[2]] {
				// omit
			} else if strings.HasPrefix(pragma[2], "!") && !s.flags[pragma[2][1:]] {
				// omit
			} else {
				left := bytes.TrimRightFunc(before, unicode.IsSpace) // omit the pragma, print everything before
				w.Write(s.applyReplacements(left))
				w.Write([]byte("\n"))
			}
		}
		// omit entire line
	case "include":
		if len(pragma) > 2 {
			if pragma[1] != "if" {
				s.error("expected 'if' for include directive")
			}
			if s.flags[pragma[2]] || (strings.HasPrefix(pragma[2], "!") && !s.flags[pragma[2][1:]]) {
				left := bytes.TrimRightFunc(before, unicode.IsSpace) // omit the pragma, print everything before
				w.Write(s.applyReplacements(left))
				w.Write([]byte("\n"))
			}
		}
		// omit entire line
	case "template":
		contents, ok := s.templates[pragma[1]]
		if !ok {
			s.error("unknown template - must be defined beforehand")
		}
		// TODO(cbro): properly handle line numbers for defs.
		s.r = bufio.NewReader(io.MultiReader(bytes.NewReader(contents), s.r)) // prepend
	case "replace":
		if len(pragma) < 3 {
			s.error("replace needs both sentinel and replacement")
		}
		var (
			sentinel    = pragma[1]
			replacement = pragma[2]
		)
		s.replacements[sentinel] = []byte(replacement)
	default:
		s.error(fmt.Sprintf("unknown directive %q", pragma[0]))
	}
}

func (s *scanner) splitPrefix(line []byte) (before []byte, pragmaTokens []string) {
	if !bytes.Contains(line, s.prefix) {
		s.error("PARANOID: bad use of splitPrefix")
	}

	split := bytes.SplitN(line, s.prefix, 2)
	return split[0], strings.Fields(string(split[1]))
}

func (s *scanner) consumeUntilEnd(w io.Writer, line []byte) {
	if !bytes.Contains(line, s.prefix) {
		return
	}
	_, pragma := s.splitPrefix(line)
	if pragma[0] == "end" {
		s.popState()
	}
}

func (s *scanner) def(name string) state {
	var content []byte
	return func(w io.Writer, line []byte) {
		if !bytes.Contains(line, s.prefix) {
			content = append(content, line...)
			return
		}
		_, pragma := s.splitPrefix(line)
		if pragma[0] == "enddef" {
			s.templates[name] = content
			s.popState()
			return
		}
		content = append(content, line...)
	}
}
