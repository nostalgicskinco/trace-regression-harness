// Copyright 2024 Nostalgic Skin Co.
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package trace reads OTLP JSON trace exports into a simplified span model
// suitable for policy evaluation.
package trace

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Span is a simplified representation of an OTel span for policy evaluation.
type Span struct {
	TraceID    string            `json:"traceId"`
	SpanID     string            `json:"spanId"`
	ParentID   string            `json:"parentSpanId,omitempty"`
	Name       string            `json:"name"`
	Kind       int               `json:"kind"`
	StartTime  time.Time         `json:"-"`
	EndTime    time.Time         `json:"-"`
	DurationMs float64           `json:"durationMs"`
	StatusCode int               `json:"statusCode"`
	StatusMsg  string            `json:"statusMessage,omitempty"`
	Attributes map[string]any    `json:"attributes,omitempty"`
	Events     []Event           `json:"events,omitempty"`
}

// Event is a simplified span event.
type Event struct {
	Name       string         `json:"name"`
	Timestamp  time.Time      `json:"-"`
	Attributes map[string]any `json:"attributes,omitempty"`
}

// TraceData holds all spans from one or more resource scopes.
type TraceData struct {
	Spans []Span `json:"spans"`
}

// otlpExport mirrors the OTLP JSON export format.
type otlpExport struct {
	ResourceSpans []resourceSpan `json:"resourceSpans"`
}

type resourceSpan struct {
	ScopeSpans []scopeSpan `json:"scopeSpans"`
}

type scopeSpan struct {
	Spans []rawSpan `json:"spans"`
}

type rawSpan struct {
	TraceID      string       `json:"traceId"`
	SpanID       string       `json:"spanId"`
	ParentSpanID string       `json:"parentSpanId,omitempty"`
	Name         string       `json:"name"`
	Kind         int          `json:"kind"`
	StartTimeNs  string       `json:"startTimeUnixNano"`
	EndTimeNs    string       `json:"endTimeUnixNano"`
	Status       *rawStatus   `json:"status,omitempty"`
	Attributes   []rawKV      `json:"attributes,omitempty"`
	Events       []rawEvent   `json:"events,omitempty"`
}

type rawStatus struct {
	Code    int    `json:"code"`
	Message string `json:"message,omitempty"`
}

type rawEvent struct {
	Name         string  `json:"name"`
	TimeNs       string  `json:"timeUnixNano,omitempty"`
	Attributes   []rawKV `json:"attributes,omitempty"`
}

type rawKV struct {
	Key   string   `json:"key"`
	Value rawValue `json:"value"`
}

type rawValue struct {
	StringValue *string  `json:"stringValue,omitempty"`
	IntValue    *string  `json:"intValue,omitempty"`
	DoubleValue *float64 `json:"doubleValue,omitempty"`
	BoolValue   *bool    `json:"boolValue,omitempty"`
}

func parseValue(v rawValue) any {
	if v.StringValue != nil {
		return *v.StringValue
	}
	if v.IntValue != nil {
		return *v.IntValue
	}
	if v.DoubleValue != nil {
		return *v.DoubleValue
	}
	if v.BoolValue != nil {
		return *v.BoolValue
	}
	return nil
}

func parseAttrs(kvs []rawKV) map[string]any {
	if len(kvs) == 0 {
		return nil
	}
	m := make(map[string]any, len(kvs))
	for _, kv := range kvs {
		m[kv.Key] = parseValue(kv.Value)
	}
	return m
}

func parseNanos(s string) time.Time {
	var ns int64
	fmt.Sscanf(s, "%d", &ns)
	if ns == 0 {
		return time.Time{}
	}
	return time.Unix(0, ns)
}

// LoadFile reads an OTLP JSON trace export file and returns TraceData.
func LoadFile(path string) (*TraceData, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read trace file: %w", err)
	}
	return Parse(data)
}

// Parse parses OTLP JSON bytes into TraceData.
func Parse(data []byte) (*TraceData, error) {
	var export otlpExport
	if err := json.Unmarshal(data, &export); err != nil {
		return nil, fmt.Errorf("parse OTLP JSON: %w", err)
	}

	var spans []Span
	for _, rs := range export.ResourceSpans {
		for _, ss := range rs.ScopeSpans {
			for _, raw := range ss.Spans {
				start := parseNanos(raw.StartTimeNs)
				end := parseNanos(raw.EndTimeNs)
				dur := end.Sub(start).Seconds() * 1000

				s := Span{
					TraceID:    raw.TraceID,
					SpanID:     raw.SpanID,
					ParentID:   raw.ParentSpanID,
					Name:       raw.Name,
					Kind:       raw.Kind,
					StartTime:  start,
					EndTime:    end,
					DurationMs: dur,
					Attributes: parseAttrs(raw.Attributes),
				}
				if raw.Status != nil {
					s.StatusCode = raw.Status.Code
					s.StatusMsg = raw.Status.Message
				}
				for _, re := range raw.Events {
					s.Events = append(s.Events, Event{
						Name:       re.Name,
						Timestamp:  parseNanos(re.TimeNs),
						Attributes: parseAttrs(re.Attributes),
					})
				}
				spans = append(spans, s)
			}
		}
	}
	return &TraceData{Spans: spans}, nil
}

// SpansWithName returns spans matching the given name.
func (td *TraceData) SpansWithName(name string) []Span {
	var out []Span
	for _, s := range td.Spans {
		if s.Name == name {
			out = append(out, s)
		}
	}
	return out
}

// SpansWithAttribute returns spans that have the given attribute key.
func (td *TraceData) SpansWithAttribute(key string) []Span {
	var out []Span
	for _, s := range td.Spans {
		if _, ok := s.Attributes[key]; ok {
			out = append(out, s)
		}
	}
	return out
}
