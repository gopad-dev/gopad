package ls

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os/exec"

	"go.lsp.dev/jsonrpc2"
	"go.lsp.dev/protocol"
)

func newServerConn(ctx context.Context, rwc io.ReadWriteCloser, client protocol.Client, w io.Writer) (jsonrpc2.Conn, protocol.Server, error) {
	stream := jsonrpc2.NewStream(rwc)
	if w != io.Discard {
		stream = protocol.LoggingStream(stream, w)
	}

	logger := slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelInfo,
	}))

	_, conn, server := protocol.NewClient(ctx, client, stream, logger)
	return conn, server, nil
}

func newServerCmdStream(ctx context.Context, w io.Writer, name string, arg ...string) (*exec.Cmd, io.ReadWriteCloser, error) {
	logger := log.New(w, name, log.LstdFlags)
	logger.Println("newServerCmdStream", name, arg)

	if r := recover(); r != nil {
		logger.Println("panic while running lsp command", r)
	}

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
			logger.Printf("lsp client %s: %s", name, s.Text())
		}
	}()

	go func() {
		if r := recover(); r != nil {
			logger.Println("panic while running lsp command", r)
		}
		if err = cmd.Wait(); err != nil {
			logger.Println("error while running lsp command", err)
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
