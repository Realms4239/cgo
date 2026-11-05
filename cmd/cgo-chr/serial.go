package main

import (
	"bufio"
	"errors"
	"io"
	"regexp"
	"strings"
	"time"
)

var (
	loginRe  = regexp.MustCompile(`(?i)login:\s*$`)
	passRe   = regexp.MustCompile(`(?i)password:\s*$`)
	promptRe = regexp.MustCompile(`\[admin@[^\]]*\]\s*>\s*$`)
	errorRe  = regexp.MustCompile(`(?i)(failure:|syntax error|bad command name|expected end of command|input does not match)`)
)

var errTimeout = errors.New("chr: délai dépassé")

// drive logs in, sends the script line by line, then captures the export.
// The transcript keeps every byte the device sent, including refusals.
func drive(tr io.ReadWriteCloser, script string, timeout time.Duration) (transcript, export, importErr string, err error) {
	r := bufio.NewReader(tr)
	var all strings.Builder
	read := func(re *regexp.Regexp) (string, error) {
		chunk, err := expect(r, re, timeout)
		all.WriteString(chunk)
		return chunk, err
	}
	if _, err := read(loginRe); err != nil {
		return all.String(), "", "", err
	}
	if _, err := tr.Write([]byte("admin\n")); err != nil {
		return all.String(), "", "", err
	}
	if _, err := read(passRe); err != nil {
		return all.String(), "", "", err
	}
	if _, err := tr.Write([]byte("\n")); err != nil {
		return all.String(), "", "", err
	}
	if _, err := read(promptRe); err != nil {
		return all.String(), "", "", err
	}
	for _, line := range strings.Split(script, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if _, err := tr.Write([]byte(line + "\n")); err != nil {
			return all.String(), "", "", err
		}
		chunk, err := read(promptRe)
		if err != nil {
			return all.String(), "", "", err
		}
		if errorRe.MatchString(chunk) {
			return all.String(), "", lastErrorLine(chunk), nil
		}
	}
	if _, err := tr.Write([]byte("/export terse\n")); err != nil {
		return all.String(), "", "", err
	}
	chunk, err := read(promptRe)
	if err != nil {
		return all.String(), "", "", err
	}
	return all.String(), cleanupExport(chunk), "", nil
}

func expect(r *bufio.Reader, re *regexp.Regexp, d time.Duration) (string, error) {
	type res struct {
		s   string
		err error
	}
	ch := make(chan res, 1)
	go func() {
		var b strings.Builder
		for {
			c, err := r.ReadByte()
			if err != nil {
				ch <- res{b.String(), err}
				return
			}
			b.WriteByte(c)
			if re.MatchString(b.String()) {
				ch <- res{b.String(), nil}
				return
			}
		}
	}()
	select {
	case v := <-ch:
		return v.s, v.err
	case <-time.After(d):
		return "", errTimeout
	}
}

// cleanupExport drops the echoed command and the trailing prompt, keeping
// only what RouterOS printed for /export.
func cleanupExport(chunk string) string {
	if i := strings.Index(chunk, "\n"); i >= 0 {
		chunk = chunk[i+1:]
	}
	if i := strings.LastIndex(chunk, "[admin@"); i >= 0 {
		chunk = chunk[:i]
	}
	return strings.TrimSpace(strings.ReplaceAll(chunk, "\r", ""))
}

func lastErrorLine(chunk string) string {
	lines := strings.Split(strings.ReplaceAll(chunk, "\r", ""), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		if errorRe.MatchString(lines[i]) {
			return lines[i]
		}
	}
	return ""
}
