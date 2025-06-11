package otel

import (
	"context"
	"log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/trace"
)

func NewOTelProvider(ctx context.Context) *trace.TracerProvider {
	provider, err := otlptracehttp.New(ctx,
		otlptracehttp.WithInsecure(),
		otlptracehttp.WithEndpoint("localhost:4318"))

	if err != nil {
		log.Fatalf("failed to create provider: %v", err)
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(provider))

	otel.SetTracerProvider(tp)
	return tp
}
