package observability

import (
	"bookadmin/global"
	"context"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

var tracerProvider *sdktrace.TracerProvider

// InitTracer 初始化 OpenTelemetry TracerProvider，向 Jaeger 等 OTLP 后端发送链路数据
func InitTracer(ctx context.Context) error {
	if global.GVA_CONFIG == nil || !global.GVA_CONFIG.Tracing.Enabled {
		return nil
	}

	endpoint := strings.TrimSpace(global.GVA_CONFIG.Tracing.Endpoint)
	if endpoint == "" {
		endpoint = "localhost:4317"
	}
	// OTLP 环境变量可能带 http:// 前缀，gRPC 需要 host:port
	endpoint = strings.TrimPrefix(endpoint, "http://")
	endpoint = strings.TrimPrefix(endpoint, "https://")

	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(endpoint),
	)
	if err != nil {
		return err
	}

	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(global.GVA_CONFIG.Tracing.Service),
		),
	)
	if err != nil {
		return err
	}

	tracerProvider = sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)
	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))
	return nil
}

// ShutdownTracer 关闭 TracerProvider，刷新未发送的 span
func ShutdownTracer(ctx context.Context) error {
	if tracerProvider == nil {
		return nil
	}
	deadline, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return tracerProvider.Shutdown(deadline)
}

// TracerEnabled 是否已启用分布式追踪
func TracerEnabled() bool {
	return tracerProvider != nil
}
