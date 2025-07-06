package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"maps"
	"os"
	"sync"
	"time"

	"github.com/Easy-Infra-Ltd/easy-test/internal/assert"
	"go.opentelemetry.io/otel/sdk/resource"
)

type Handler struct {
	handler              slog.Handler
	replaceAttr          func([]string, slog.Attr) slog.Attr
	buf                  *bytes.Buffer
	mutex                *sync.Mutex
	writer               io.Writer
	timestamp            bool
	colourise            bool
	outputEmptyAttrs     bool
	resource             *resource.Resource
	instrumentationScope *InstrumentationScope
	enableOTEL           bool
}

type InstrumentationScope struct {
	Name    string
	Version string
}

func (h *Handler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &Handler{
		handler:              h.handler.WithAttrs(attrs),
		buf:                  h.buf,
		replaceAttr:          h.replaceAttr,
		mutex:                h.mutex,
		writer:               h.writer,
		colourise:            h.colourise,
		resource:             h.resource,
		instrumentationScope: h.instrumentationScope,
		enableOTEL:           h.enableOTEL,
	}
}

func (h *Handler) WithGroup(name string) slog.Handler {
	return &Handler{
		handler:              h.handler.WithGroup(name),
		buf:                  h.buf,
		replaceAttr:          h.replaceAttr,
		mutex:                h.mutex,
		writer:               h.writer,
		colourise:            h.colourise,
		resource:             h.resource,
		instrumentationScope: h.instrumentationScope,
		enableOTEL:           h.enableOTEL,
	}
}

func (h *Handler) computeAttrs(ctx context.Context, r slog.Record) (map[string]any, error) {
	h.mutex.Lock()
	defer func() {
		h.buf.Reset()
		h.mutex.Unlock()
	}()

	if err := h.handler.Handle(ctx, r); err != nil {
		return nil, fmt.Errorf("error when calling inner handler's Handle %w", err)
	}

	var attrs map[string]any
	err := json.Unmarshal(h.buf.Bytes(), &attrs)
	if err != nil {
		return nil, fmt.Errorf("error when unmarshalling inner handler's Handle result: %w", err)
	}

	return attrs, nil
}

func (h *Handler) Handle(ctx context.Context, r slog.Record) error {
	colourise := func(code int, value string) string {
		return value
	}

	if h.colourise {
		colourise = Colouriser
	}

	toPretty := map[string]any{}

	levelAttr := slog.Attr{
		Key:   slog.LevelKey,
		Value: slog.AnyValue(r.Level),
	}

	if h.replaceAttr != nil {
		levelAttr = h.replaceAttr([]string{}, levelAttr)
	}

	msgAttr := slog.Attr{
		Key:   slog.MessageKey,
		Value: slog.StringValue(r.Message),
	}

	if h.replaceAttr != nil {
		msgAttr = h.replaceAttr([]string{}, msgAttr)
	}

	attrs, err := h.computeAttrs(ctx, r)
	if err != nil {
		return err
	}

	toPretty[slog.LevelKey] = levelAttr.Value.String()
	toPretty[slog.MessageKey] = msgAttr.Value.String()
	toPretty[ProcessKey] = attrs[ProcessKey]
	toPretty[AreaKey] = attrs[AreaKey]

	if h.enableOTEL {
		traceID, spanID, traceFlags := extractTraceContext(ctx)
		if traceID != "" {
			toPretty[OTELTraceIDKey] = traceID
		}
		if spanID != "" {
			toPretty[OTELSpanIDKey] = spanID
		}
		if traceFlags != "" {
			toPretty[OTELTraceFlagsKey] = traceFlags
		}

		toPretty[OTELTimestampKey] = r.Time.Format(time.RFC3339Nano)

		toPretty[OTELSeverityTextKey] = levelAttr.Value.String()
		toPretty[OTELSeverityNumberKey] = mapSlogLevelToOTELSeverity(r.Level)

		serviceName, serviceVersion := h.getServiceInfo()
		toPretty[OTELServiceNameKey] = serviceName
		toPretty[OTELServiceVersionKey] = serviceVersion

		if h.instrumentationScope != nil {
			toPretty[OTELScopeNameKey] = h.instrumentationScope.Name
			toPretty[OTELScopeVersionKey] = h.instrumentationScope.Version
		}

		toPretty[OTELBodyKey] = msgAttr.Value.String()
	}

	maps.Copy(toPretty, attrs)

	str, err := PrettyLine(toPretty, colourise)
	if err != nil {
		fallback := fmt.Sprintf("%s: %s %s\n",
			toPretty[ProcessKey],
			toPretty[slog.LevelKey],
			toPretty[slog.MessageKey])
		if h.writer != nil {
			io.WriteString(h.writer, fallback)
		}
		return fmt.Errorf("pretty line formatting failed, used fallback: %w", err)
	}

	if h.writer != nil {
		io.WriteString(h.writer, str)
		io.WriteString(h.writer, "\n")
	}

	return nil
}

func suppressDefaults(next func([]string, slog.Attr) slog.Attr) func([]string, slog.Attr) slog.Attr {
	return func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == slog.TimeKey || a.Key == slog.LevelKey || a.Key == slog.MessageKey {
			return slog.Attr{}
		}
		if next == nil {
			return a
		}

		return next(groups, a)
	}
}

func New(handlerOptions *slog.HandlerOptions, options ...Option) *Handler {
	if handlerOptions == nil {
		handlerOptions = &slog.HandlerOptions{}
	}

	buf := &bytes.Buffer{}
	handler := &Handler{
		buf:       buf,
		timestamp: false,
		handler: slog.NewJSONHandler(buf, &slog.HandlerOptions{
			Level:       handlerOptions.Level,
			AddSource:   handlerOptions.AddSource,
			ReplaceAttr: suppressDefaults(handlerOptions.ReplaceAttr),
		}),
		replaceAttr: handlerOptions.ReplaceAttr,
		mutex:       &sync.Mutex{},
	}

	for _, opt := range options {
		opt(handler)
	}

	return handler
}

func NewHandler(opts *slog.HandlerOptions, params PrettyLogParams, options ...Option) *Handler {
	options = append([]Option{
		WithDestinationWriter(params.Out),
		WithColour(),
		WithOutputEmptyAttrs(),
	}, options...)

	return New(opts, options...)
}

func NewOTELHandler(opts *slog.HandlerOptions, params PrettyLogParams, options ...Option) *Handler {
	options = append([]Option{
		WithDestinationWriter(params.Out),
		WithColour(),
		WithOutputEmptyAttrs(),
		WithOpenTelemetry(),
		WithResource(NewDefaultResource()),
		WithInstrumentationScope("easy-test/logger", "1.0.0"),
	}, options...)

	return New(opts, options...)
}

type Option func(h *Handler)

func WithTimestamp() Option {
	return func(h *Handler) {
		h.timestamp = true
	}
}

func WithDestinationWriter(writer io.Writer) Option {
	return func(h *Handler) {
		h.writer = writer
	}
}

func WithColour() Option {
	return func(h *Handler) {
		h.colourise = true
	}
}

func WithOutputEmptyAttrs() Option {
	return func(h *Handler) {
		h.outputEmptyAttrs = true
	}
}

func WithOpenTelemetry() Option {
	return func(h *Handler) {
		h.enableOTEL = true
	}
}

func WithResource(res *resource.Resource) Option {
	return func(h *Handler) {
		h.resource = res
	}
}

func WithInstrumentationScope(name, version string) Option {
	return func(h *Handler) {
		h.instrumentationScope = &InstrumentationScope{
			Name:    name,
			Version: version,
		}
	}
}

type PrettyLogParams struct {
	Out   io.Writer
	Level slog.Level
}

func NewParams(out io.Writer) PrettyLogParams {
	return PrettyLogParams{
		Level: LevelTrace,

		Out: out,
	}
}

func SetProgramLevelPrettyLogger(params PrettyLogParams) *slog.Logger {
	if os.Getenv("NO_PRETTY_LOGGER") != "" {
		return slog.Default()
	}

	prettyHandler := NewHandler(&slog.HandlerOptions{
		Level:       params.Level,
		AddSource:   false,
		ReplaceAttr: nil,
	}, params)
	logger := slog.New(prettyHandler)
	slog.SetDefault(logger)

	return logger
}

func CreateLoggerSink() *os.File {
	var f *os.File
	var err error

	debugLog := os.Getenv("DEBUG_LOG")
	if debugLog == "" {
		f = os.Stderr
	} else {
		f, err = os.OpenFile(debugLog, os.O_RDWR|os.O_CREATE, 0644)
		assert.NoError(err, "unable to create temporary file")
	}

	return f
}

func CreateLoggerFromEnv(out *os.File, colour string) *slog.Logger {
	if out == nil {
		out = CreateLoggerSink()
	}

	setProcessColour(colour)

	if os.Getenv("DEBUG_TYPE") == "pretty" {
		return SetProgramLevelPrettyLogger(NewParams(out))
	}

	logger := slog.New(slog.NewJSONHandler(out, nil))
	slog.SetDefault(logger)
	return logger
}

func CreateOTELLoggerFromEnv(out *os.File, colour string) *slog.Logger {
	if out == nil {
		out = CreateLoggerSink()
	}

	setProcessColour(colour)

	if os.Getenv("DEBUG_TYPE") == "pretty" || os.Getenv("OTEL_LOGS_ENABLED") == "true" {
		prettyHandler := NewOTELHandler(&slog.HandlerOptions{
			Level:     LevelTrace,
			AddSource: false,
		}, NewParams(out))
		logger := slog.New(prettyHandler)
		slog.SetDefault(logger)
		return logger
	}

	logger := slog.New(slog.NewJSONHandler(out, nil))
	slog.SetDefault(logger)
	return logger
}

func Trace(log *slog.Logger, msg string, data ...any) {
	log.Log(context.Background(), LevelTrace, msg, data...)
}
