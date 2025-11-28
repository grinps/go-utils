package errext

import (
	"fmt"
	"strings"
	"testing"
)

func TestWithAttributes(t *testing.T) {
	t.Run("DefaultAttributes", func(t *testing.T) {
		ec := NewErrorCodeWithOptions(WithErrorCodeAndType(false, 1, DefaultErrorCodeType), WithAttributes("component", "database", "severity", "high"))

		err := ec.New("connection failed")
		msg := err.Error()

		expected := "connection failed [component=database severity=high]"
		if msg != expected {
			t.Errorf("Expected %q, got %q", expected, msg)
		}
	})

	t.Run("RuntimeAttributes", func(t *testing.T) {
		ec := NewErrorCode(2)

		err := ec.New("connection failed", "user_id", 123, "retry", true)
		msg := err.Error()

		expected := "connection failed [user_id=123 retry=true]"
		if msg != expected {
			t.Errorf("Expected %q, got %q", expected, msg)
		}
	})

	t.Run("MergedAttributes", func(t *testing.T) {
		ec := NewErrorCodeWithOptions(WithErrorCodeAndType(false, 3, DefaultErrorCodeType), WithAttributes("component", "api"))

		err := ec.New("request failed", "request_id", "req-1")
		msg := err.Error()

		expected := "request failed [component=api request_id=req-1]"
		if msg != expected {
			t.Errorf("Expected %q, got %q", expected, msg)
		}
	})

	t.Run("AttributesFormatting", func(t *testing.T) {
		args := []interface{}{"key1", "val1", "key2"}
		formatted := formatAttributes(args)
		expected := "[key1=val1 key2=null]" // odd number of args handling
		if formatted != expected {
			t.Errorf("Expected %q, got %q", expected, formatted)
		}
	})
}

func TestErrorFormatting(t *testing.T) {
	ec := NewErrorCode(10)
	err := ec.New("error msg", "k", "v")

	t.Run("String", func(t *testing.T) {
		if s := err.Error(); s != "error msg [k=v]" {
			t.Errorf("Unexpected string: %s", s)
		}
	})

	t.Run("Format+v", func(t *testing.T) {
		s := fmt.Sprintf("%+v", err)
		if !strings.Contains(s, "error msg [k=v]") {
			t.Errorf("Expected message and attributes in %+v", s)
		}
	})
}

func Example() {
	ec := NewErrorCodeWithOptions(WithErrorCode(1), WithAttributes("sys", "main"))
	err := ec.New("failed", "id", 10)
	fmt.Println(err)
	// Output:
	// failed [sys=main id=10]
}
