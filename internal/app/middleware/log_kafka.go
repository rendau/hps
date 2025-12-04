package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/segmentio/kafka-go"
)

type LogKafka struct {
	writer *kafka.Writer
}

func NewLogKafka(host, topic string) *LogKafka {
	if host == "" || topic == "" {
		return &LogKafka{}
	}

	writer := &kafka.Writer{
		Addr:                   kafka.TCP(host),
		Topic:                  topic,
		AllowAutoTopicCreation: true,
		Async:                  true,
	}

	slog.Info("kafka writer created", "host", host, "topic", topic)

	return &LogKafka{writer: writer}
}

type kafkaMessage struct {
	Ts        time.Time       `json:"ts"`
	Method    string          `json:"method"`
	Path      string          `json:"path"`
	ReqBody   json.RawMessage `json:"req_body"`
	RepStatus int             `json:"rep_status"`
	RepBody   json.RawMessage `json:"rep_body"`
}

func (m *LogKafka) Middleware(next http.Handler) http.Handler {
	if m.writer == nil {
		return next
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reqBody []byte
		if r.Body != nil {
			reqBody, _ = io.ReadAll(r.Body)
			r.Body = io.NopCloser(bytes.NewBuffer(reqBody))
		}

		rw := newResponseWriter(w)
		next.ServeHTTP(rw, r)

		normalizedReqBody, _ := normalizeJSON(reqBody)

		normalizedRepBody, ok := normalizeJSON(rw.body.Bytes())
		if !ok {
			slog.Error("response body is not valid", "method", r.Method, "path", r.URL.Path, "status", rw.statusCode, "body", string(rw.body.Bytes()))
		}

		go m.sendToKafka(&kafkaMessage{
			Ts:        time.Now().UTC(),
			Method:    r.Method,
			Path:      r.URL.Path,
			ReqBody:   normalizedReqBody,
			RepStatus: rw.statusCode,
			RepBody:   normalizedRepBody,
		})
	})
}

func (m *LogKafka) sendToKafka(msg *kafkaMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		slog.Error("failed to marshal log message", "error", err, "msg", msg)
		return
	}

	err = m.writer.WriteMessages(context.Background(),
		kafka.Message{
			Key:   []byte(msg.Method + " " + msg.Path),
			Value: data,
		},
	)
	if err != nil {
		slog.Error("failed to write message to kafka", "error", err, "msg", msg)
	}

	// slog.Info("message sent to kafka", "msg", msg)
}

func (m *LogKafka) Close() {
	_ = m.writer.Close()
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

func normalizeJSON(data []byte) (json.RawMessage, bool) {
	if len(data) == 0 {
		return json.RawMessage("null"), true
	}
	if json.Valid(data) {
		return data, true
	}
	return json.RawMessage("null"), false
}
