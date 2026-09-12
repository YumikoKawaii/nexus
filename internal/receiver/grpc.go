package receiver

import (
	"context"
	"fmt"
	"log/slog"
	"net"

	"google.golang.org/grpc"
	_ "google.golang.org/grpc/encoding/gzip"

	collogspb "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	colmetricspb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	coltracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
)

type traceServer struct {
	coltracepb.UnimplementedTraceServiceServer
	svc *Service
}

func (t *traceServer) Export(ctx context.Context, req *coltracepb.ExportTraceServiceRequest) (*coltracepb.ExportTraceServiceResponse, error) {
	t.svc.HandleTraces(ctx, req.GetResourceSpans())
	return &coltracepb.ExportTraceServiceResponse{}, nil
}

type logsServer struct {
	collogspb.UnimplementedLogsServiceServer
	svc *Service
}

func (l *logsServer) Export(ctx context.Context, req *collogspb.ExportLogsServiceRequest) (*collogspb.ExportLogsServiceResponse, error) {
	l.svc.HandleLogs(ctx, req.GetResourceLogs())
	return &collogspb.ExportLogsServiceResponse{}, nil
}

type metricsServer struct {
	colmetricspb.UnimplementedMetricsServiceServer
	svc *Service
}

func (m *metricsServer) Export(ctx context.Context, req *colmetricspb.ExportMetricsServiceRequest) (*colmetricspb.ExportMetricsServiceResponse, error) {
	m.svc.HandleMetrics(ctx, req.GetResourceMetrics())
	return &colmetricspb.ExportMetricsServiceResponse{}, nil
}

type GRPCServer struct {
	srv  *grpc.Server
	addr string
	ln   net.Listener
}

func NewGRPCServer(addr string, maxRecvMsgSizeMiB int, svc *Service) *GRPCServer {
	srv := grpc.NewServer(
		grpc.MaxRecvMsgSize(maxRecvMsgSizeMiB * 1024 * 1024),
	)
	coltracepb.RegisterTraceServiceServer(srv, &traceServer{svc: svc})
	collogspb.RegisterLogsServiceServer(srv, &logsServer{svc: svc})
	colmetricspb.RegisterMetricsServiceServer(srv, &metricsServer{svc: svc})
	return &GRPCServer{srv: srv, addr: addr}
}

func (s *GRPCServer) Start(logger *slog.Logger) error {
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("grpc listen %s: %w", s.addr, err)
	}
	s.ln = ln
	go func() {
		logger.Info("otlp grpc listening", "addr", s.addr)
		if err := s.srv.Serve(ln); err != nil {
			logger.Error("grpc serve exited", "err", err)
		}
	}()
	return nil
}

func (s *GRPCServer) Stop() {
	s.srv.GracefulStop()
}
