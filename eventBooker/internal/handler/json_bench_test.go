package handlers

import (
	"bytes"
	"net/http"
	"testing"
	"time"

	"github.com/Egor-Pomidor-pdf/EventBooker/internal/models"
	"github.com/gin-gonic/gin/render"
)

type benchResponseWriter struct {
	header http.Header
	buf    *bytes.Buffer
	status int
}

func newBenchResponseWriter() *benchResponseWriter {
	return &benchResponseWriter{
		header: make(http.Header),
		buf:    &bytes.Buffer{},
	}
}

func (w *benchResponseWriter) Header() http.Header {
	return w.header
}

func (w *benchResponseWriter) Write(b []byte) (int, error) {
	return w.buf.Write(b)
}

func (w *benchResponseWriter) WriteHeader(statusCode int) {
	w.status = statusCode
}

func makeEvents(count int) []models.EventWithFreePlaces {
	events := make([]models.EventWithFreePlaces, count)
	base := time.Date(2026, 12, 31, 12, 0, 0, 0, time.UTC)
	for i := 0; i < count; i++ {
		ev := models.Event{
			ID:        int64(i + 1),
			Title:     "Event",
			StartTime: base.Add(time.Duration(i) * time.Minute),
			Capacity:  100,
			CreatedAt: base,
		}
		events[i] = models.EventWithFreePlaces{
			Event:      &ev,
			FreePlaces: 100,
		}
	}
	return events
}

func BenchmarkEventsJSON(b *testing.B) {
	payload := makeEvents(200)
	r := render.JSON{Data: payload}
	w := newBenchResponseWriter()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		w.buf.Reset()
		if err := r.Render(w); err != nil {
			b.Fatal(err)
		}
	}
}
