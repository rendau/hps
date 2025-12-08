package middleware

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
)

type LogKafka struct {
	writer      *kafka.Writer
	filterRules []FilterRule
}

func NewLogKafka(host, topic string, filterRules []string) *LogKafka {
	if host == "" || topic == "" {
		return &LogKafka{}
	}

	return &LogKafka{
		writer: &kafka.Writer{
			Addr:                   kafka.TCP(host),
			Topic:                  topic,
			AllowAutoTopicCreation: true,
			Async:                  true,
		},
		filterRules: parseFilterRules(filterRules),
	}
}

type kafkaMessage struct {
	Ts        time.Time       `json:"ts"`
	Method    string          `json:"method"`
	Path      string          `json:"path"`
	Query     string          `json:"query"`
	ReqBody   json.RawMessage `json:"req_body"`
	RepStatus int             `json:"rep_status"`
	RepBody   json.RawMessage `json:"rep_body"`
	SessionID string          `json:"session_id,omitempty"`
}

func (m *LogKafka) Middleware(next http.Handler) http.Handler {
	if m.writer == nil {
		return next
	}

	slog.Info("Kafka writer created", "host", m.writer.Addr.String(), "topic", m.writer.Topic)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !m.filter(r.Method, r.URL.Path) {
			// slog.Info("filter rejected", "method", r.Method, "path", r.URL.Path)
			next.ServeHTTP(w, r)
			return
		}

		var reqBody []byte
		if r.Body != nil {
			reqBody, _ = io.ReadAll(r.Body)
			r.Body = io.NopCloser(bytes.NewBuffer(reqBody))
		}

		rw := newResponseWriter(w)
		next.ServeHTTP(rw, r)

		normalizedReqBody, _ := normalizeJSON(reqBody, false)

		normalizedRepBody, ok := normalizeJSON(rw.body.Bytes(), rw.Header().Get("Content-Encoding") == "gzip")
		if !ok {
			// slog.Error("response body is not valid", "method", r.Method, "path", r.URL.Path, "status", rw.statusCode, "body", string(rw.body.Bytes()))
		}

		go m.sendToKafka(&kafkaMessage{
			Ts:        time.Now().UTC(),
			Method:    r.Method,
			Path:      r.URL.Path,
			Query:     r.URL.RawQuery,
			ReqBody:   normalizedReqBody,
			RepStatus: rw.statusCode,
			RepBody:   normalizedRepBody,
			SessionID: ContextSessionID(r.Context()),
		})
	})
}

func (m *LogKafka) sendToKafka(msg *kafkaMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		slog.Error("failed to marshal log message", "error", err, "msg", msg)
		return
	}

	key := msg.SessionID
	if key == "" {
		key = msg.Method + " " + msg.Path
	}

	err = m.writer.WriteMessages(context.Background(),
		kafka.Message{
			Key:   []byte(key),
			Value: data,
		},
	)
	if err != nil {
		slog.Error("failed to write message to kafka", "error", err, "msg", msg)
	}

	// slog.Info("message sent to kafka", "msg", msg)
}

func (m *LogKafka) Close() {
	if m.writer != nil {
		_ = m.writer.Close()
	}
}

// responseWriter обёртка для захвата ответа
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	body       *bytes.Buffer
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
		body:           &bytes.Buffer{},
	}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	rw.body.Write(b)
	return rw.ResponseWriter.Write(b)
}

func normalizeJSON(data []byte, isGzip bool) (json.RawMessage, bool) {
	if isGzip {
		// slog.Info("gzip response body detected")
		gzipReader, err := gzip.NewReader(bytes.NewReader(data))
		if err == nil {
			defer gzipReader.Close()
			data, err = io.ReadAll(gzipReader)
			if err != nil {
				data = nil
			}
		} else {
			data = nil
		}
	}
	if len(data) == 0 {
		return json.RawMessage("null"), true
	}
	if json.Valid(data) {
		return data, true
	}
	return json.RawMessage("null"), false
}

type FilterRule struct {
	Method  string // пустой = любой метод
	Pattern string
}

func (r *FilterRule) String() string {
	if r.Method != "" {
		return "{" + r.Method + " " + r.Pattern + "}"
	}
	return r.Pattern
}

func parseFilterRules(src []string) []FilterRule {
	result := make([]FilterRule, 0, len(src))

	for _, r := range src {
		r = strings.TrimSpace(r)
		method := ""
		pattern := ""
		parts := strings.SplitN(r, ":", 2)
		switch len(parts) {
		case 1:
			pattern = parts[0]
		case 2:
			method = parts[0]
			pattern = parts[1]
		default:
			slog.Error("invalid filter rule", "rule", r)
			continue
		}
		if pattern == "" {
			slog.Error("empty filter rule", "rule", r)
			continue
		}
		result = append(result, FilterRule{
			Method:  strings.ToUpper(method),
			Pattern: strings.ToLower("/" + strings.Trim(pattern, "/")),
		})
	}

	// print parsed rules
	slog.Info("Applied filter rules:")
	for _, r := range result {
		slog.Info(r.String())
	}

	return result
}

func (m *LogKafka) filter(method, pathStr string) bool {
	if len(m.filterRules) == 0 {
		return true
	}

	pathStr = strings.ToLower("/" + strings.Trim(pathStr, "/"))
	method = strings.ToUpper(method)

	for _, rule := range m.filterRules {
		if rule.Method != "" && rule.Method != method {
			continue
		}

		if matched, _ := path.Match(rule.Pattern, pathStr); matched {
			return true
		}
	}

	return false
}
