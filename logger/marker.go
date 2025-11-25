package logger

import (
	"fmt"
	"log/slog"
	"strings"
)

// Marker represents a hierarchical execution path.
// It implements slog.LogValuer to be used directly in structured logs.
//
// It uses a linked-list structure to ensure O(1) allocation during
// creation and modification, deferring string generation until the
// log is actually written.
type Marker struct {
	node *markerNode
}

type markerNode struct {
	parent   *markerNode
	name     string
	args     []any
	isMethod bool // distinguishes between simple Add (namespace) and AddMethod
}

// NewMarker creates a new marker with the given initial path.
func NewMarker(name string) Marker {
	return Marker{
		node: &markerNode{
			name: name,
		},
	}
}

// String returns the string representation of the marker.
// This constructs the full path on demand.
func (m Marker) String() string {
	if m.node == nil {
		return ""
	}
	return m.node.buildString()
}

// LogValue implements slog.LogValuer.
// It constructs the string representation only when the log is processed.
func (m Marker) LogValue() slog.Value {
	return slog.StringValue(m.String())
}

// Add appends components to the marker path.
// This is an O(1) operation.
func (m Marker) Add(components ...string) Marker {
	if len(components) == 0 {
		return m
	}
	// We join the components here because they are likely static strings,
	// but we link to the parent instead of concatenating the full history.
	name := strings.Join(components, ".")

	return Marker{
		node: &markerNode{
			parent:   m.node,
			name:     name,
			isMethod: false,
		},
	}
}

// AddMethod adds a method call to the marker path.
// This is an O(1) operation. Arguments are not formatted until logging.
func (m Marker) AddMethod(method string, args ...any) Marker {
	return Marker{
		node: &markerNode{
			parent:   m.node,
			name:     method,
			args:     args,
			isMethod: true,
		},
	}
}

// buildString traverses the linked list recursively to build the path.
func (n *markerNode) buildString() string {
	var sb strings.Builder
	// Estimate length to reduce re-allocations
	// A simple heuristic: 10 chars per node is better than 0
	sb.Grow(32)
	n.writeTo(&sb)
	return sb.String()
}

func (n *markerNode) writeTo(sb *strings.Builder) {
	if n.parent != nil {
		n.parent.writeTo(sb)
		sb.WriteString(".")
	}

	sb.WriteString(n.name)

	if n.isMethod {
		sb.WriteString("(")
		if len(n.args) > 0 {
			// We format args here, just in time.
			argsStr := fmt.Sprintln(n.args...)
			sb.WriteString(strings.TrimRight(argsStr, "\r\n"))
		}
		sb.WriteString(")")
	}
}
