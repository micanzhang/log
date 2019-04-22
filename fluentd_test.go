package log

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"reflect"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

func TestFormat(t *testing.T) {
	addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:8080")
	if err != nil {
		t.Fatal(err)
	}
	var nilPointer *User
	fields := logrus.Fields{
		"status":        404,
		"body":          []byte("Not Found"),
		"trace.spanid":  123456789,
		"trace.traceid": "42q54faf98745",
		"header": map[string][]string{
			"Content-Type": []string{"plain/text"},
		},
		"error": errors.New("not found"),
		"addr":  addr,
		"user":  &User{"gopher", 8, 1024},
		"nil": nilPointer,
	}
	tsFormat := time.RFC3339
	buf := bytes.NewBuffer([]byte{})
	l := logrus.New()
	l.Out = buf
	l.SetLevel(logrus.DebugLevel)
	l.Formatter = &FluentdFormatter{TimestampFormat: tsFormat}
	l.WithFields(fields).Debug(logrus.DebugLevel.String())
	l.WithFields(fields).Info(logrus.InfoLevel.String())
	l.WithFields(fields).Warn(logrus.WarnLevel.String())
	l.WithFields(fields).Error(logrus.ErrorLevel.String())
	for {
		data, err := buf.ReadBytes('\n')
		if err != nil && err != io.EOF {
			t.Error(err)
			break
		}
		if data == nil || err == io.EOF {
			break
		}
		m := make(map[string]interface{})
		if err := json.NewDecoder(bytes.NewBuffer(data)).Decode(&m); err != nil {
			t.Error(err)
			break
		}
		if got, expected := len(m), len(fields)+3; got != expected {
			t.Errorf("expected: %d, got: %d", expected, got)
		}
		level, msg, lt := m["severity"], m["message"], m["time"]
		if _, err := logrus.ParseLevel(level.(string)); err != nil {
			t.Error(err)
		}
		if level != msg {
			t.Errorf("%s should be equal to %s", msg, level)
		}
		if _, err := time.Parse(tsFormat, lt.(string)); err != nil {
			t.Error(err)
		}
	}
}

type User struct {
	Username string
	Age      int
	ID       uint64
}

func TestValue(t *testing.T) {
	var (
		s   = "hello, world"
		arr = [2]int{0, 1}
		sl  = []uint64{0, 1}
		c   = make(chan bool, 0)
		fn  = fmt.Print
		in  = (io.Reader)(nil)
		m   = map[string]interface{}{"string": "goher"}
		err = errors.New("not found")
		u   = User{"gopher", 8, 1024}
		now = time.Now()
	)
	cases := []struct {
		Original interface{}
		Value    interface{}
		Compare  func(a, b interface{}) bool
	}{
		{0, 0, reflect.DeepEqual},                         // int
		{3.14, 3.14, reflect.DeepEqual},                   // float
		{true, true, reflect.DeepEqual},                   // bool
		{s, s, reflect.DeepEqual},                         // string
		{arr, fmt.Sprintf("%+v", arr), reflect.DeepEqual}, // array
		{c, c, reflect.DeepEqual},                         // chan
		{fn, fn, fnCompare},                               // func
		{in, in, reflect.DeepEqual},                       // interface
		{m, fmt.Sprintf("%+v", m), reflect.DeepEqual},     // map
		{err, err.Error(), reflect.DeepEqual},             // error
		{sl, fmt.Sprintf("%+v", sl), reflect.DeepEqual},   // slice
		{u, fmt.Sprintf("%+v", u), reflect.DeepEqual},     // struct
		{&u, fmt.Sprintf("%+v", u), reflect.DeepEqual},    // pointer
		{now, now.String(), reflect.DeepEqual},            // fmt.Stringer
	}

	for i, tc := range cases {
		if got, expected := Value(tc.Original), tc.Value; !tc.Compare(got, expected) {
			gt, et := reflect.TypeOf(got), reflect.TypeOf(expected)
			t.Errorf("[%d] expected type: %s, got type: %s", i, et, gt)
			t.Errorf("[%d] expected: %#v, got: %#v", i, expected, got)
		}
	}
}

func fnCompare(fn1, fn2 interface{}) bool {
	v1, v2 := reflect.ValueOf(fn1), reflect.ValueOf(fn2)
	if v1.Type() != v2.Type() {
		return false
	}
	return reflect.ValueOf(fn1).Pointer() == reflect.ValueOf(fn2).Pointer()
}
