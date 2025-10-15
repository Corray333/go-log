package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

const (
	reset = "\033[0m"

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

func colorize(colorCode int, v string) string {
	return fmt.Sprintf("\033[%sm%s%s", strconv.Itoa(colorCode), v, reset)
}

type handler struct {
	h           slog.Handler
	b           *bytes.Buffer
	m           *sync.Mutex
	colorize    bool
	prettyPrint bool
}

func (h *handler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.h.Enabled(ctx, level)
}

func (h *handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &handler{h: h.h.WithAttrs(attrs), b: h.b, m: h.m, colorize: h.colorize, prettyPrint: h.prettyPrint}
}

func (h *handler) WithGroup(name string) slog.Handler {
	return &handler{h: h.h.WithGroup(name), b: h.b, m: h.m, colorize: h.colorize, prettyPrint: h.prettyPrint}
}

const (
	timeFormat = "[2006-01-02 15:04:05.000]"
)

func (h *handler) Handle(ctx context.Context, r slog.Record) error {

	level := r.Level.String() + ":"

	attrs, err := h.computeAttrs(ctx, r)
	if err != nil {
		return err
	}

	if r.Level == slog.LevelError {
		// Skip three levels of slog functions calls
		_, file, line, ok := runtime.Caller(3)
		if !ok {
			file = "unknown"
			line = 0
		}

		attrs["file"] = file
		attrs["line"] = line
	}

	var attrsBytes []byte
	if h.prettyPrint {
		attrsBytes, err = json.MarshalIndent(attrs, "", "  ")
	} else {
		attrsBytes, err = json.Marshal(attrs)
	}
	if err != nil {
		return fmt.Errorf("error when marshaling attrs: %w", err)
	}

	if h.colorize {
		switch r.Level {
		case slog.LevelDebug:
			level = colorize(darkGray, level)
		case slog.LevelInfo:
			level = colorize(cyan, level)
		case slog.LevelWarn:
			level = colorize(lightYellow, level)
		case slog.LevelError:
			level = colorize(lightRed, level)
		}

		fmt.Println(
			colorize(lightGray, r.Time.Format(timeFormat)),
			level,
			colorize(white, r.Message),
			colorize(darkGray, string(attrsBytes)),
		)
	} else {
		fmt.Println(
			r.Time.Format(timeFormat),
			level,
			r.Message,
			string(attrsBytes),
		)
	}

	return nil
}

func suppressDefaults(
	next func([]string, slog.Attr) slog.Attr,
) func([]string, slog.Attr) slog.Attr {
	return func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == slog.TimeKey ||
			a.Key == slog.LevelKey ||
			a.Key == slog.MessageKey {
			return slog.Attr{}
		}
		if next == nil {
			return a
		}
		return next(groups, a)
	}
}

type HandlerOptions struct {
	*slog.HandlerOptions
	Colorize    bool
	PrettyPrint bool
}

func NewHandler(opts *HandlerOptions) *handler {
	if opts == nil {
		opts = &HandlerOptions{}
	}
	if opts.HandlerOptions == nil {
		opts.HandlerOptions = &slog.HandlerOptions{}
	}
	b := &bytes.Buffer{}

	return &handler{
		b: b,
		h: slog.NewJSONHandler(b, &slog.HandlerOptions{
			Level:       opts.Level,
			AddSource:   opts.AddSource,
			ReplaceAttr: suppressDefaults(opts.ReplaceAttr),
		}),
		m:           &sync.Mutex{},
		colorize:    opts.Colorize,
		prettyPrint: opts.PrettyPrint,
	}
}

func (h *handler) computeAttrs(ctx context.Context, r slog.Record) (map[string]any, error) {
	h.m.Lock()
	defer func() {
		h.b.Reset()
		h.m.Unlock()
	}()
	if err := h.h.Handle(ctx, r); err != nil {
		return nil, fmt.Errorf("error when calling inner handler's Handle: %w", err)
	}

	var attrs map[string]any
	err := json.Unmarshal(h.b.Bytes(), &attrs)
	if err != nil {
		return nil, fmt.Errorf("error when unmarshaling inner handler's Handle result: %w", err)
	}
	return attrs, nil
}

func SetupLoggerWith(opts *HandlerOptions) {
	handler := NewHandler(opts)

	logger := slog.New(handler)

	slog.SetDefault(logger)
}

func new(log *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		log = log.With(
			slog.String("component", "middleware/logger"),
		)

		log.Info("logger middleware enabled")

		fn := func(w http.ResponseWriter, r *http.Request) {
			entry := log.With(
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("remote_addr", r.RemoteAddr),
				slog.String("user_agent", r.UserAgent()),
				slog.String("request_id", middleware.GetReqID(r.Context())),
			)

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			t1 := time.Now()
			defer func() {
				entry.Info("request completed",
					slog.Int("status", ww.Status()),
					slog.Int("size", ww.BytesWritten()),
					slog.Duration("duration", time.Since(t1)),
				)
			}()
			next.ServeHTTP(ww, r)
		}
		return http.HandlerFunc(fn)
	}
}
