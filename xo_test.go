package xo

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/as/io/spaz"
)

// change to false to run all the rests
// currently will break
const skipRegressions = true

func driver(rie []string) (ac string, err error) {
	re, in, ex := rie[0], rie[1], rie[2]
	r, err := NewReaderString(bytes.NewReader([]byte(in)), "", re)
	if err != nil && err != io.EOF {
		return
	}
	buf, _, err := r.Structure()
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("fail: unexpected err: %s", err)
	}
	ac = string(buf)
	if ac != ex {
		return ac, fmt.Errorf("fail: %q != %q\n", ac, ex)
	}
	return ac, nil
}

func multi(rie []string) (ac string, err error) {
	re, in := rie[0], rie[1]
	r, err := NewReaderString(bytes.NewReader([]byte(in)), "", re)
	if err != nil && err != io.EOF {
		return
	}
	for i, ex := range rie[2:] {
		verb.Printf("\tmulti #%d: ex=%q\n", i, ex)
		buf, _, err := r.Structure()
		if buf == nil {
			return ac, fmt.Errorf("buf is nil")
		}
		if err != nil {
			return ac, fmt.Errorf("fail: %q != %q\n", ac, ex)
		}
		ac = string(buf)
		if ac != ex {
			return ac, fmt.Errorf("fail: %q != %q\n", ac, ex)
		}
	}
	return ac, nil
}

func nonuniform(rie []string) (ac string, err error) {
	re, in := rie[0], rie[1]
	r, err := NewReaderString(spaz.NewReader([]byte(in)), "", re)
	if err != nil && err != io.EOF {
		return
	}
	for i, ex := range rie[2:] {
		verb.Printf("\tmulti #%d: ex=%q\n", i, ex)
		buf, _, err := r.Structure()
		if buf == nil {
			return ac, fmt.Errorf("buf is nil")
		}
		if err != nil {
			return ac, fmt.Errorf("fail: %q != %q\n", ac, ex)
		}
		ac = string(buf)
		if ac != ex {
			return ac, fmt.Errorf("fail: %q != %q\n", ac, ex)
		}
	}
	return ac, nil
}

var ls = `./go/src/9fans.net/go/.git/config
./go/src/9fans.net/go/.hgignore
./go/src/9fans.net/go/LICENSE
./go/src/9fans.net/go/README
./go/src/9fans.net/go/acme/Dict/Dict.go
./go/src/9fans.net/go/acme/Makefile
./go/src/9fans.net/go/acme/Watch/main.go
./go/src/9fans.net/go/acme/acme.go`

func TestSub(t *testing.T) {
	var err error
	table := [][]string{
		// {`/abc/-/.../`, "abcdefg", ""}, // drained
		{"/aa/-/a/", "aaaa", "a"},
		{"/aaa/-/a/", "aaaa", "aa"},
		{"/aaaa/-/a/", "aaaaa", "aaa"},
		{"/baaaa/-/aa/", "baaaaaa", "baa"},
		{`/ /-/..../`, "the quick brown fox", "the"},
		{`/aa/-/./`, "aaa", "a"},
		{`/abc/-/.../`, "abcdefg", ""},
		{`/abc/-/./`, "abcdefg", "ab"},
		{`/abc/-/./`, "zabcdefg", "ab"},
		{`/abcd/-/../`, "abcdefg", "ab"},
	}
	for i, v := range table {
		_, err = driver(v)
		if err != nil && err != io.EOF {
			t.Logf("test: %d: %s", i, err)
			t.Fail()
			return
		}
		t.Log("test", i, "pass")
	}
}

func TestAdd(t *testing.T) {
	var err error
	table := [][]string{
		{"/(xz)+../", "xzxzzzzzzxzzzzzx", "xzxzzz"},
		{"/../", "ab", "ab"},
		{"/./", "a\n", "a"},
		{"/./", "xzxzzzzzzxzzzzzx", "x"},
		{"/\x00/", "\x00", "\x00"},
		{"/a./", "ab", "ab"},
		{"/a/", "a", "a"},
		{"/a/", "aa", "a"},
		{"/aa/", "aa", "aa"},
		{"/ab/", "ab", "ab"},
		{"/c//c/", "cc", "c"},
		{"/gab/+/we/", "gabenewell", "we"},
		{"/zx/", "xzzzzzzzzzzzzzzzzzzzzzzzx", "zx"},
		{"/zx/", "zzzzzzzzzzzzzzzzzzzzzzzzx", "zx"},
		{"/⁵/", "⁵", "⁵"},
	}
	for i, v := range table {
		_, err = driver(v)
		if err != nil && err != io.EOF {
			t.Logf("test: %d: %s", i, err)
			t.Fail()
			return
		}
		t.Log("test", i, "pass")
	}
}

func TestSem(t *testing.T) {
	var err error
	table := [][]string{
		{`/\n/;/..../`, "/dir/f1.png.bat.jpg\n", ".jpg\n"},
		{`/\n/;/[^.]+\./`, "/dir/f2.jpg.bat.png\n", ".png\n"},
	}
	for i, v := range table {
		_, err = driver(v)
		if err != nil && err != io.EOF {
			t.Logf("test: %d: %s", i, err)
			t.Fail()
			return
		}
		t.Log("test", i, "pass")
	}
}

func TestExcise(t *testing.T) {
	var err error
	table := [][]string{
		{"/@//./,/[a-z]+./", "@structure@", "structure@"},
		{"/@/,/@/", "@structure@", "@structure@"},
		{"/.+@/-/./", "structure@", "structure@"},
		{"/@//structur/-/./", "@structure@", "structu"},

		// Bug: If the end of the input is reached, backtracking
		// doesn't work. This test then fails. The fix is to suppress
		// io.EOF in Sub and Sem Ops.
		//
		// Uncomment the third line to make bug

		{"/@//structur/-/./", "@structure@", "structu"},
		{"/@//structure/-/./", "@structure@", "structur"},
		// {"/@//structure@/-/./", "@structure@",  "structure"},
	}
	for i, v := range table {
		_, err = driver(v)
		if err != nil && err != io.EOF {
			t.Logf("test: %d: %s", i, err)
			t.Fail()
			return
		}
		t.Log("test", i, "pass")
	}
}

func TestStuck(t *testing.T) {
	var err error
	in := "000111222333444555"

	table := [][]string{
		{`/.../`, in, "000", "111", "222", "333"},
		{`/[A-Z]+/,/\n\n/`, man,
			"NAME\n	xo - Search for patterns in arbitrary structures\n\n",
			"SYNOPSIS\n\txo [flags] [-x linedef] regexp [file ...]\n\n",
		},
	}

	for i, v := range table {
		_, err = multi(v)
		if err != nil {
			t.Logf("test: %d: %s", i, err)
			t.Fail()
			return
		}
		t.Log("test", i, "pass")
	}
}

func TestCom(t *testing.T) {
	var err error
	table := [][]string{
		//{"/..../,/z+/,/b/", "idontzebra", "idontzeb"},
		{"/@//./,/[a-z]+./", "@bullshit@", "bullshit@"},
		{"/@/,/@/", "@bullshit@", "@bullshit@"},
		{"/[a-z]+/,/sucks/", "xml sucks", "xml sucks"},
		{"/\x00\x00/,/\x01/", "\x00\x00\x11\x01", "\x00\x00\x11\x01"},
		{"/a/,/a/", "aa", "aa"},
		{"/a/,/b/", "aaabaab", "aaab"},
		{"/a/,/b/", "aaabaab", "aaab"},
		{"/a/,/b/", "aabaab", "aab"},
		{"/a/,/b/", "aabaab", "aab"},
		{"/a/,/b/", "ab", "ab"},
		{"/a/,/b/", "abab", "ab"},
		{"/a/,/b/", "abab", "ab"},
		{"/a/,/b/", "abc", "ab"},
		{"/a/,/b/", "abcd", "ab"},
		{"/a/,/c/", "abc", "abc"},
		{"/aa/,/b/", "aaabaab", "aaab"},
		{"/d/,/d/", "dd", "dd"},
		{"/z/,/./", "idontzzebraz", "zz"},
		{"/z/,/./", "idontzzebraz", "zz"},
		{`/ab/,/./`, "abcdefg", "abc"},
		{`/abc/,/./`, "abcdefg", "abcd"},
	}
	for i, v := range table {
		_, err = driver(v)
		if err != nil && err != io.EOF {
			t.Logf("test: %d: %s", i, err)
			t.Fail()
			return
		}
		t.Log("test", i, "pass")
	}
}

func TestRegress(t *testing.T) {
	if skipRegressions {
		return
	}
	var err error
	table := [][]string{
		{"/..../,/z+/,/b/", "idontzebra", "idontzeb"},
		{`,/(\n )/`, ls, "sys=xray ", "ip=1.1.1.1\n", "sys=present ", "ip=1.1.1.2\n", "sys=web "}, // crash
		{`/\n/;/(\/|\.)/`, ls, ".git/config", ".hgignore"},                                        // crash: fix the parser
		{`/abc/-/.../`, "abcdefg", ""},                                                            // drained
		{`/abcd/-/../`, "abcdefg", "ab"},
		{`,/$/-/^/+/./`, "this\nis\na\test", "\n"},
		{`,/./`, "dot test", "d", "o", "t", " ", "t", "e", "s", "t"},
		{`/abc/-/../`, "abcdefg", "a"},
	}
	for i, v := range table {
		_, err = driver(v)
		if err != nil && err != io.EOF {
			t.Logf("test: %d: %s", i, err)
			t.Fail()
			return
		}
		t.Log("test", i, "pass")
	}
}

func TestMultiLine(t *testing.T) {
	var err error
	ndb := "sys=xray ip=1.1.1.1\nsys=present ip=1.1.1.2\nsys=web ip=1.1.1.1.3\n"

	table := [][]string{
		{`/\n/;/[^.]+/`, ls, "git/config\n", "hgignore\n"},
		{`/\n/;/(\x2f|\.)/`, ls, "/config\n", ".hgignore\n"},
		{`/\n/`, "this\nis\na\ntest", "\n", "\n"},
		{`,/\n/`, "this\nis\na\ntest", "this\n", "is\n"},
		{`,/\n/`, ndb, "sys=xray ip=1.1.1.1\n", "sys=present ip=1.1.1.2\n"},
		{`,/ /`, ndb, "sys=xray ", "ip=1.1.1.1\nsys=present ", "ip=1.1.1.2\nsys=web "},
		{`,/[\n ]/`, ndb, "sys=xray ", "ip=1.1.1.1\n", "sys=present ", "ip=1.1.1.2\n", "sys=web "},
		{`,/(\n| )/`, ndb, "sys=xray ", "ip=1.1.1.1\n", "sys=present ", "ip=1.1.1.2\n", "sys=web "},
		{`,/(\n| |=)/`, ndb, "sys=", "xray ", "ip=", "1.1.1.1\n", "sys=", "present ", "ip=", "1.1.1.2\n", "sys=", "web "},
		{`/[A-Z]/,/\n[A-Z]/-/./                 `, man, "NAME\n	xo - Search for patterns in arbitrary structures\n\n"},
		{`/[A-Z]+/,/\n\n/`, man,
			"NAME\n	xo - Search for patterns in arbitrary structures\n\n",
			"SYNOPSIS\n\txo [flags] [-x linedef] regexp [file ...]\n\n",
		},
	}

	for i, v := range table {
		_, err = multi(v)
		if err != nil {
			t.Logf("test: %d: %s", i, err)
			t.Fail()
			return
		}
		t.Log("test", i, "pass")
	}
}

func TestNonUniform(t *testing.T) {
	in, err := ioutil.ReadFile("test/walkdist")
	re := `/[^\n]+/`
	ex := strings.Split(string(in), "\n")
	if err != nil {
		t.Logf("TestNonUniform: %s", err)
		t.Fail()
		return
	}

	r, err := NewReaderString(spaz.NewReader([]byte(in)), "", re)
	for i, exln := range ex {
		ln, _, err := r.Structure()
		fmt.Printf("%q\n",ln)
		if err != nil && err != io.EOF {
			t.Logf("TestNonUniform: %s", err)
			t.Fail()
			return
		}
		if err == io.EOF {
			t.Logf("TestNonUniform: unexpected EOF: read: %d/%d", i, len(ex))
			t.Fail()
			return
		}
		if exln != string(ln) {
			t.Logf("TestNonUniform: read: %d/%d: ac != ex: %q != %q", i, len(ex), ln, exln)
			t.Fail()
			return
		}
	}
}

var man = `
NAME
	xo - Search for patterns in arbitrary structures

SYNOPSIS
	xo [flags] [-x linedef] regexp [file ...]

DESCRIPTION
	Xo scans files for pattern using regexp. By default xo
	applies regexp to each line and prints matching lines found.
	This default behavior is similar to Plan 9 grep.

	However, the concept of a line is altered using -x by setting
	linedef to a structural regular expression set in the form:

	   -x /start/
	   -x /start/,
	   -x ,/stop/
	   -x /start/,/stop/

	Start, stop, and all the data between these two regular
	expressions, forms linedef, the operational definition of a line.

	The default linedef is simply: /\n/

	Xo reads lines from stdin unless a file list is given. If '-' is 
	present in the file list, xo reads a list of files from
	stdin instead of treating stdin as a file.

FLAGS
	Linedef:

	-x linedef	Redefine a line based on linedef
	-y linedef	The negation of linedef becomes linedef

	Regexp:

	-v regexp	Reverse. Print the lines not matching regexp
	-f file     File contains a list of regexps, one per line
				the newline is treated as an OR

	Tagging:

	-o  Preprend file:rune,rune offsets
	-l	Preprend file:line,line offsets
	-L  Print file names containing no matches
	-p  Print new line after every match

EXAMPLE
	# Examples operate on this help page, so
	xo -h > help.txt

	# Print the DESCRIPTION section from this help
	xo -p -o -x '/^[A-Z]/,/./' . help.txt

	# Print the Tagging sub-section
	xo -h | xo -x '/[A-Z][a-z]+:/,/\n\n/' Tagging

BUGS
	On a multi-line match, xo -l prints the offset
	of the final line in that match.
	
`
