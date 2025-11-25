package test

import (
	opentracing "github.com/opentracing/opentracing-go"
)

type MemTracer struct {
	opentracing.Tracer
	Spans []string
}

func (t *MemTracer) StartSpan(operationName string, opts ...opentracing.StartSpanOption) opentracing.Span {
	t.Spans = append(t.Spans, operationName)
	return t.Tracer.StartSpan(operationName, opts...)
}

func NewMemTracer() *MemTracer {
	return &MemTracer{
		Tracer: opentracing.NoopTracer{},
	}
}
