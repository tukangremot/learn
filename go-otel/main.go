package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func initTracer() (*sdktrace.TracerProvider, error) {
	ctx := context.Background()
	conn, err := grpc.NewClient("192.168.100.80:4317", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	exporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	// Mengonfigurasi resource untuk service
	res, err := resource.New(ctx,
		resource.WithAttributes(
			attribute.String("service.name", "learn-go-otel"),
			attribute.String("service.version", "1.0.0"),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Membuat tracer provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)

	return tp, nil
}

type Logger struct {
	*log.Logger
}

func NewLogger() *Logger {
	return &Logger{
		Logger: log.New(os.Stdout, "", log.LstdFlags),
	}
}

func (l *Logger) LogWithTrace(ctx context.Context, msg string) {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		traceID := span.SpanContext().TraceID().String()
		spanID := span.SpanContext().SpanID().String()
		l.Printf("[traceID=%s spanID=%s] %s", traceID, spanID, msg)
	} else {
		l.Println(msg)
	}
}

func main() {
	tp, err := initTracer()
	if err != nil {
		log.Fatalf("failed to initialize tracer: %v", err)
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Fatalf("failed to shutdown tracer provider: %v", err)
		}
	}()

	logger := NewLogger()

	// Mendapatkan tracer dari OpenTelemetry
	tracer := otel.Tracer("example-tracer")

	// Membuat span root
	ctx, span := tracer.Start(context.Background(), "main-operation")
	defer span.End()

	// Simulasi pemrosesan request
	time.Sleep(3 * time.Second)

	// Contoh pemanggilan fungsi dengan tracing
	processRequest(ctx, tracer)

	logger.LogWithTrace(ctx, "This is a log message with trace context")
}

func processRequest(ctx context.Context, tracer trace.Tracer) {
	// Membuat span baru dalam proses
	_, span := tracer.Start(ctx, "processRequest")
	defer span.End()

	// Simulasi pemrosesan request
	time.Sleep(3 * time.Second)

	// Menambahkan atribut tambahan ke span
	span.SetAttributes(attribute.String("example.attribute", "example-value"))

	subProcessRequest(ctx, tracer)
}

func subProcessRequest(ctx context.Context, tracer trace.Tracer) {
	// Membuat span baru dalam proses
	_, span := tracer.Start(ctx, "subProcessRequest")
	defer span.End()

	// Simulasi pemrosesan request
	time.Sleep(3 * time.Second)

	// Menambahkan atribut tambahan ke span
	span.SetAttributes(attribute.String("example.attribute", "example-value"))
}
