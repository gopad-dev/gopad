package lsp

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os/exec"

	"go.lsp.dev/jsonrpc2"
	"go.lsp.dev/protocol"
	"go.uber.org/zap"
)

func newServer(ctx context.Context, rwc io.ReadWriteCloser, client protocol.Client, w io.Writer) (jsonrpc2.Conn, protocol.Server, error) {
	stream := jsonrpc2.NewStream(rwc)
	if w != nil {
		stream = protocol.LoggingStream(stream, w)
	}
	_, conn, server := protocol.NewClient(ctx, client, stream, zap.NewNop())
	return conn, server, nil
}

func newCmdStream(ctx context.Context, name string, arg ...string) (*exec.Cmd, io.ReadWriteCloser, error) {
	log.Println("newCmdStream", name, arg)
	cmd := exec.CommandContext(ctx, name, arg...)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, nil, err
	}

	if err = cmd.Start(); err != nil {
		return nil, nil, err
	}

	go func() {
		s := bufio.NewScanner(stderr)
		for s.Scan() {
			log.Printf("lsp client %s: %s", name, s.Text())
		}
	}()

	go func() {
		if err = cmd.Wait(); err != nil {
			log.Println("error while running lsp command", err)
		}
	}()

	return cmd, &processReadWriter{
		in:  stdin,
		out: stdout,
	}, nil
}

type processReadWriter struct {
	in  io.WriteCloser
	out io.ReadCloser
}

func (prw *processReadWriter) Read(p []byte) (n int, err error) {
	return prw.out.Read(p)
}

func (prw *processReadWriter) Write(p []byte) (n int, err error) {
	return prw.in.Write(p)
}

func (prw *processReadWriter) Close() error {
	errInClose := prw.in.Close()
	errOutClose := prw.out.Close()
	if errInClose != nil || errOutClose != nil {
		return fmt.Errorf("error closing process - in: %w, out: %w", errInClose, errOutClose)
	}
	return nil
}
