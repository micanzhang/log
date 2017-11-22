package log

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/sirupsen/logrus"
)

// FluentdFormatter is similar to logrus.JSONFormatter but with log level that are recongnized
// by kubernetes fluentd.
type FluentdFormatter struct {
	TimestampFormat string
}

// Format the log entry. Implements logrus.Formatter.
func (f *FluentdFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	data := make(logrus.Fields, len(entry.Data)+3)
	for k, v := range entry.Data {
		data[k] = Value(v)
	}
	prefixFieldClashes(data)

	timestampFormat := f.TimestampFormat
	if timestampFormat == "" {
		timestampFormat = time.RFC3339
	}

	data["time"] = entry.Time.Format(timestampFormat)
	data["message"] = entry.Message
	data["severity"] = entry.Level.String()

	serialized, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("Failed to marshal fields to JSON, %v", err)
	}
	return append(serialized, '\n'), nil
}

func Value(i interface{}) interface{} {
	v := reflect.ValueOf(i)
	kind := v.Kind()
	if v, ok := i.(error); ok {
		return v.Error()
	}
	if kind == reflect.Ptr {
		return Value(v.Elem().Interface())
	}
	// handle basic type
	if kind < reflect.Array {
		return i
	}
	// handle string type
	if kind == reflect.String {
		return i
	}
	// handle type implement fmt.Stringer
	if s, ok := i.(fmt.Stringer); ok {
		return s.String()
	}
	if kind == reflect.Array || kind == reflect.Map || kind == reflect.Slice || kind == reflect.Struct {
		return fmt.Sprintf("%+v", i)
	}
	return i
}

func prefixFieldClashes(data logrus.Fields) {
	if t, ok := data["time"]; ok {
		data["fields.time"] = t
	}

	if m, ok := data["msg"]; ok {
		data["fields.msg"] = m
	}

	if l, ok := data["level"]; ok {
		data["fields.level"] = l
	}
}
