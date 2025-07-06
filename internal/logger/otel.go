package logger

import (
	"context"
	"log/slog"
	"os"

	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

func extractTraceContext(ctx context.Context) (traceID, spanID, traceFlags string) {
	spanContext := trace.SpanContextFromContext(ctx)
	if spanContext.IsValid() {
		traceID = spanContext.TraceID().String()
		spanID = spanContext.SpanID().String()
		if spanContext.IsSampled() {
			traceFlags = "01"
		} else {
			traceFlags = "00"
		}
	}
	return
}

func mapSlogLevelToOTELSeverity(level slog.Level) int {
	switch {
	case level == LevelTrace:
		return 1
	case level == slog.LevelDebug:
		return 5
	case level == slog.LevelInfo:
		return 9
	case level == slog.LevelWarn:
		return 13
	case level == slog.LevelError:
		return 17
	case level == LevelFatal:
		return 21
	default:
		return 9
	}
}

func (h *Handler) getServiceInfo() (serviceName, serviceVersion string) {
	serviceName = os.Getenv("OTEL_SERVICE_NAME")
	if serviceName == "" {
		serviceName = "unknown_service"
	}

	serviceVersion = os.Getenv("OTEL_SERVICE_VERSION")
	if serviceVersion == "" {
		serviceVersion = "unknown"
	}

	if h.resource != nil {
		for _, attr := range h.resource.Attributes() {
			switch attr.Key {
			case semconv.ServiceNameKey:
				serviceName = attr.Value.AsString()
			case semconv.ServiceVersionKey:
				serviceVersion = attr.Value.AsString()
			}
		}
	}

	return
}

func NewDefaultResource() *resource.Resource {
	serviceName := os.Getenv("OTEL_SERVICE_NAME")
	if serviceName == "" {
		serviceName = "easy-test"
	}

	serviceVersion := os.Getenv("OTEL_SERVICE_VERSION")
	if serviceVersion == "" {
		serviceVersion = "1.0.0"
	}

	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(serviceVersion),
			semconv.DeploymentEnvironment(os.Getenv("DEPLOYMENT_ENVIRONMENT")),
		),
	)
	if err != nil {
		res = resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(serviceVersion),
		)
	}

	return res
}
