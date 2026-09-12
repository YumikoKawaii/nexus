package receiver

import (
	"compress/flate"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/klauspost/compress/zstd"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	collogspb "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	colmetricspb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	coltracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
)

const maxBodyBytes = 16 << 20

type HTTPServer struct {
	srv *http.Server
	svc *Service
}

func NewHTTPServer(addr string, svc *Service) *HTTPServer {
	h := &HTTPServer{svc: svc}
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/traces", h.handleTraces)
	mux.HandleFunc("/v1/logs", h.handleLogs)
	mux.HandleFunc("/v1/metrics", h.handleMetrics)
	h.srv = &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}
	return h
}

func (h *HTTPServer) Start(logger *slog.Logger) error {
	go func() {
		logger.Info("otlp http listening", "addr", h.srv.Addr)
		if err := h.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("http serve exited", "err", err)
		}
	}()
	return nil
}

func (h *HTTPServer) Stop(ctx context.Context) {
	_ = h.srv.Shutdown(ctx)
}

func (h *HTTPServer) handleTraces(w http.ResponseWriter, r *http.Request) {
	body, isJSON, ok := readBody(w, r)
	if !ok {
		return
	}
	var req coltracepb.ExportTraceServiceRequest
	if err := unmarshal(body, isJSON, &req); err != nil {
		httpError(w, http.StatusBadRequest, err)
		return
	}
	h.svc.HandleTraces(r.Context(), req.GetResourceSpans())
	writeResponse(w, r, isJSON, &coltracepb.ExportTraceServiceResponse{})
}

func (h *HTTPServer) handleLogs(w http.ResponseWriter, r *http.Request) {
	body, isJSON, ok := readBody(w, r)
	if !ok {
		return
	}
	var req collogspb.ExportLogsServiceRequest
	if err := unmarshal(body, isJSON, &req); err != nil {
		httpError(w, http.StatusBadRequest, err)
		return
	}
	h.svc.HandleLogs(r.Context(), req.GetResourceLogs())
	writeResponse(w, r, isJSON, &collogspb.ExportLogsServiceResponse{})
}

func (h *HTTPServer) handleMetrics(w http.ResponseWriter, r *http.Request) {
	body, isJSON, ok := readBody(w, r)
	if !ok {
		return
	}
	var req colmetricspb.ExportMetricsServiceRequest
	if err := unmarshal(body, isJSON, &req); err != nil {
		httpError(w, http.StatusBadRequest, err)
		return
	}
	h.svc.HandleMetrics(r.Context(), req.GetResourceMetrics())
	writeResponse(w, r, isJSON, &colmetricspb.ExportMetricsServiceResponse{})
}

func readBody(w http.ResponseWriter, r *http.Request) (body []byte, isJSON, ok bool) {
	if r.Method != http.MethodPost {
		httpError(w, http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"))
		return nil, false, false
	}
	isJSON = strings.HasPrefix(r.Header.Get("Content-Type"), "application/json")
	reader, err := decompressReader(r.Header.Get("Content-Encoding"), http.MaxBytesReader(w, r.Body, maxBodyBytes))
	if err != nil {
		httpError(w, http.StatusBadRequest, err)
		return nil, false, false
	}
	defer reader.Close()
	b, err := io.ReadAll(reader)
	if err != nil {
		httpError(w, http.StatusBadRequest, err)
		return nil, false, false
	}
	return b, isJSON, true
}

func decompressReader(encoding string, body io.Reader) (io.ReadCloser, error) {
	switch strings.ToLower(strings.TrimSpace(encoding)) {
	case "", "identity":
		return io.NopCloser(body), nil
	case "gzip":
		return gzip.NewReader(body)
	case "deflate":
		return flate.NewReader(body), nil
	case "zstd":
		zr, err := zstd.NewReader(body)
		if err != nil {
			return nil, err
		}
		return zr.IOReadCloser(), nil
	default:
		return nil, fmt.Errorf("unsupported content-encoding %q", encoding)
	}
}

func unmarshal(body []byte, isJSON bool, msg proto.Message) error {
	if isJSON {
		return protojson.Unmarshal(body, msg)
	}
	return proto.Unmarshal(body, msg)
}

func writeResponse(w http.ResponseWriter, r *http.Request, isJSON bool, msg proto.Message) {
	var (
		b   []byte
		err error
	)
	if isJSON {
		w.Header().Set("Content-Type", "application/json")
		b, err = protojson.Marshal(msg)
	} else {
		w.Header().Set("Content-Type", "application/x-protobuf")
		b, err = proto.Marshal(msg)
	}
	if err != nil {
		httpError(w, http.StatusInternalServerError, err)
		return
	}
	if acceptsGzip(r) {
		w.Header().Set("Content-Encoding", "gzip")
		w.WriteHeader(http.StatusOK)
		gz := gzip.NewWriter(w)
		defer gz.Close()
		_, _ = gz.Write(b)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(b)
}

func acceptsGzip(r *http.Request) bool {
	for _, enc := range strings.Split(r.Header.Get("Accept-Encoding"), ",") {
		if strings.EqualFold(strings.TrimSpace(strings.SplitN(enc, ";", 2)[0]), "gzip") {
			return true
		}
	}
	return false
}

func httpError(w http.ResponseWriter, code int, err error) {
	http.Error(w, err.Error(), code)
}
