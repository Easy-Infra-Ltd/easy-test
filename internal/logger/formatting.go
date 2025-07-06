package logger

import (
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/Easy-Infra-Ltd/easy-test/internal/assert"
)

var processColour int64 = lightGray

func setProcessColour(colour string) {
	c, ok := colourMap[colour]
	assert.Assert(ok, "Process colour must exist in colour map")

	atomic.StoreInt64(&processColour, int64(c))
}

func getProcessColour(process string) int {
	return int(atomic.LoadInt64(&processColour))
}

var areaColours = map[string]int{}
var areaColoursIdx = 0
var acMutex sync.Mutex

func getAreaColour(area string) int {
	acMutex.Lock()
	defer acMutex.Unlock()

	colour, ok := areaColours[area]
	if !ok {
		colour = allColours[areaColoursIdx%len(allColours)]
		areaColours[area] = colour
		areaColoursIdx++
	}

	return colour
}

func stringifyAttrs(attrs map[string]any) string {
	str := strings.Builder{}
	keys := slices.Sorted(maps.Keys(attrs))

	for _, k := range keys {
		if isHandledKey(k) {
			continue
		}

		v := attrs[k]
		str.WriteString(k)
		str.WriteString("=")

		switch v.(type) {
		case string, int, float32, float64, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			str.WriteString(fmt.Sprintf("%v", v))
		default:
			str.WriteString(fmt.Sprintf("%+v", v))
		}
	}

	return strings.TrimSpace(str.String())
}

func Colouriser(colourCode int, v string) string {
	return fmt.Sprintf("\033[%sm%s%s", strconv.Itoa(colourCode), v, reset)
}

func PrettyLine(data map[string]any, colourise func(code int, value string) string) (string, error) {
	process, ok := data[ProcessKey]
	if !ok {
		return "", fmt.Errorf("process must be defined when calling PrettyLine")
	}
	area, ok := data[AreaKey]
	if !ok {
		return "", fmt.Errorf("area must be defined when calling PrettyLine")
	}

	levelRaw, ok := data["level"]
	if !ok {
		return "", fmt.Errorf("level must be defined in log data")
	}
	level, ok := levelRaw.(string)
	if !ok {
		return "", fmt.Errorf("level must be a string, got %T", levelRaw)
	}
	if level == "DEBUG-4" {
		level = "TRACE"
	}

	switch level {
	case "TRACE":
		fallthrough
	case "DEBUG":
		level = colourise(lightGray, level)
	case "INFO":
		level = colourise(cyan, level)
	case "WARN":
		level = colourise(lightYellow, level)
	case "ERROR":
		level = colourise(red, level)
	case "FATAL":
		level = colourise(magenta, level)
	default:
		assert.Never("unrecognised log level", "level", level)
	}

	msgRaw, ok := data["msg"]
	if !ok {
		return "", fmt.Errorf("msg must be defined in log data")
	}
	msg, ok := msgRaw.(string)
	if !ok {
		return "", fmt.Errorf("msg must be a string, got %T", msgRaw)
	}
	msg = colourise(white, msg)

	var attrsAsBytes []byte
	var err error
	attrString := stringifyAttrs(data)
	if len(attrString) > MaxInlineAttrsLength {
		attrsAsBytes, err = json.MarshalIndent(data, "", JSONIndentSpaces)
		if err != nil {
			return "", fmt.Errorf("error when marshalling attrs: %w", err)
		}
	} else {
		attrsAsBytes = []byte(attrString)
	}

	header := strings.Builder{}
	body := strings.Builder{}

	processStr, ok := process.(string)
	if !ok {
		return "", fmt.Errorf("process must be a string, got %T", process)
	}
	areaStr, ok := area.(string)
	if !ok {
		return "", fmt.Errorf("area must be a string, got %T", area)
	}

	header.WriteString(colourise(getProcessColour(processStr), processStr))
	header.WriteString(":")
	header.WriteString(colourise(getAreaColour(areaStr), areaStr))

	if traceID, exists := data[OTELTraceIDKey]; exists && traceID != "" {
		if traceIDStr, ok := traceID.(string); ok && len(traceIDStr) >= 8 {
			header.WriteString(" ")
			header.WriteString(colourise(darkGray, fmt.Sprintf("[%s]", traceIDStr[:8])))
		}
	}

	header.WriteString(" ")

	body.WriteString(level)
	body.WriteString(" ")
	body.WriteString(msg)

	if serviceName, exists := data[OTELServiceNameKey]; exists {
		body.WriteString(" ")
		body.WriteString(colourise(lightGray, fmt.Sprintf("service=%s", serviceName)))
	}

	if len(attrsAsBytes) > 0 {
		body.WriteString(" ")
		body.WriteString(colourise(lightGray, string(attrsAsBytes)))
	}

	return header.String() + body.String(), nil
}
