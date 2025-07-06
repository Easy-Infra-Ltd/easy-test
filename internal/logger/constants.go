package logger

import "log/slog"

const (
	timeFormat = "[18:43:02.000]"

	reset = "\033[0m"

	MaxInlineAttrsLength = 42
	JSONIndentSpaces     = " "

	black        = 30
	red          = 31
	green        = 32
	yellow       = 33
	blue         = 34
	magenta      = 35
	cyan         = 36
	lightGray    = 37
	darkGray     = 90
	lightRed     = 91
	lightGreen   = 92
	lightYellow  = 93
	lightBlue    = 94
	lightMagenta = 95
	lightCyan    = 96
	white        = 97
)

const LevelTrace = slog.LevelDebug - 4
const LevelFatal = slog.LevelError + 4
const ProcessKey = "process"
const AreaKey = "area"

const (
	OTELTraceIDKey        = "trace_id"
	OTELSpanIDKey         = "span_id"
	OTELTraceFlagsKey     = "trace_flags"
	OTELServiceNameKey    = "service.name"
	OTELServiceVersionKey = "service.version"
	OTELResourceKey       = "resource"
	OTELScopeNameKey      = "scope.name"
	OTELScopeVersionKey   = "scope.version"
	OTELTimestampKey      = "timestamp"
	OTELSeverityTextKey   = "severity_text"
	OTELSeverityNumberKey = "severity_number"
	OTELBodyKey           = "body"
	OTELAttributesKey     = "attributes"
)

var allColours = []int{
	31,
	32,
	33,
	34,
	35,
	36,
	37,
	90,
	91,
	92,
	93,
	94,
	95,
	96,
	97,
}

var colourMap = map[string]int{
	"black":        black,
	"red":          red,
	"green":        green,
	"yellow":       yellow,
	"blue":         blue,
	"magenta":      magenta,
	"cyan":         cyan,
	"lightGray":    lightGray,
	"darkGray":     darkGray,
	"lightRed":     lightRed,
	"lightGreen":   lightGreen,
	"lightYellow":  lightYellow,
	"lightBlue":    lightBlue,
	"lightMagenta": lightMagenta,
	"lightCyan":    lightCyan,
	"white":        white,
}

var handledKeysMap = map[string]struct{}{
	ProcessKey:            {},
	AreaKey:               {},
	slog.LevelKey:         {},
	slog.MessageKey:       {},
	slog.TimeKey:          {},
	OTELTraceIDKey:        {},
	OTELSpanIDKey:         {},
	OTELTraceFlagsKey:     {},
	OTELServiceNameKey:    {},
	OTELServiceVersionKey: {},
	OTELResourceKey:       {},
	OTELScopeNameKey:      {},
	OTELScopeVersionKey:   {},
	OTELTimestampKey:      {},
	OTELSeverityTextKey:   {},
	OTELSeverityNumberKey: {},
	OTELBodyKey:           {},
	OTELAttributesKey:     {},
}

func isHandledKey(key string) bool {
	_, ok := handledKeysMap[key]
	return ok
}
