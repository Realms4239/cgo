package main

import (
	"io"
	"os/exec"
)

type Transport interface {
	io.ReadWriteCloser
}

type qemuTransport struct {
	cmd *exec.Cmd
	in  io.WriteCloser
	out io.Reader
}

func (q *qemuTransport) Read(p []byte) (int, error)  { return q.out.Read(p) }
func (q *qemuTransport) Write(p []byte) (int, error) { return q.in.Write(p) }
func (q *qemuTransport) Close() error {
	_ = q.in.Close()
	return q.cmd.Process.Kill()
}
