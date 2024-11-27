package outtrace

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type flusher interface {
	Flush() error
}

var _ sdktrace.SpanExporter = &Exporter{}

func New(w io.Writer) *Exporter {
	return &Exporter{
		w: w,
	}
}

type Exporter struct {
	w io.Writer

	stoppedMu sync.RWMutex
	stopped   bool
}

func (e *Exporter) ExportSpans(ctx context.Context, spans []sdktrace.ReadOnlySpan) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	e.stoppedMu.RLock()
	stopped := e.stopped
	e.stoppedMu.RUnlock()
	if stopped {
		return nil
	}

	if len(spans) == 0 {
		return nil
	}

	for _, span := range spans {
		if span.Parent().IsValid() {
			continue
		}
		e.printSpan(span, 0, spans)
		_, _ = e.w.Write([]byte("\n\n"))
		if f, ok := e.w.(flusher); ok {
			_ = f.Flush()
		}
	}

	return nil
}

func (e *Exporter) printSpan(span sdktrace.ReadOnlySpan, depth int, spans []sdktrace.ReadOnlySpan) {
	_, _ = e.w.Write([]byte(formatSpan(span, depth) + "\n"))

	for _, childSpan := range spans {
		if childSpan.Parent().SpanID() == span.SpanContext().SpanID() {
			e.printSpan(childSpan, depth+1, spans)
		}
	}
}

func formatSpan(span sdktrace.ReadOnlySpan, depth int) string {
	indentation := strings.Repeat(" ", depth)

	content := fmt.Sprintf("%s%s - %s", indentation, span.Name(), span.EndTime().Sub(span.StartTime()))

	for _, attribute := range span.Attributes() {
		content += fmt.Sprintf("\n%s • %s: %s", indentation, attribute.Key, attribute.Value.Emit())
	}

	return content
}

func (e *Exporter) Shutdown(ctx context.Context) error {
	e.stoppedMu.Lock()
	e.stopped = true
	e.stoppedMu.Unlock()

	return nil
}
