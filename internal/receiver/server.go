package receiver

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"connectrpc.com/vanguard"
	"google.golang.org/grpc"

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

type Server struct {
	srv  *http.Server
	addr string
}

func NewServer(addr string, maxRecvMsgSizeMiB int, svc *Service) (*Server, error) {
	grpcSrv := grpc.NewServer(grpc.MaxRecvMsgSize(maxRecvMsgSizeMiB * 1024 * 1024))
	coltracepb.RegisterTraceServiceServer(grpcSrv, &traceServer{svc: svc})
	collogspb.RegisterLogsServiceServer(grpcSrv, &logsServer{svc: svc})
	colmetricspb.RegisterMetricsServiceServer(grpcSrv, &metricsServer{svc: svc})

	services := []*vanguard.Service{
		vanguard.NewService(coltracepb.TraceService_ServiceDesc.ServiceName, grpcSrv, vanguard.WithTargetProtocols(vanguard.ProtocolGRPC)),
		vanguard.NewService(collogspb.LogsService_ServiceDesc.ServiceName, grpcSrv, vanguard.WithTargetProtocols(vanguard.ProtocolGRPC)),
		vanguard.NewService(colmetricspb.MetricsService_ServiceDesc.ServiceName, grpcSrv, vanguard.WithTargetProtocols(vanguard.ProtocolGRPC)),
	}
	transcoder, err := vanguard.NewTranscoder(services)
	if err != nil {
		return nil, fmt.Errorf("vanguard transcoder: %w", err)
	}

	protocols := new(http.Protocols)
	protocols.SetHTTP1(true)
	protocols.SetUnencryptedHTTP2(true)

	return &Server{
		addr: addr,
		srv: &http.Server{
			Addr:              addr,
			Handler:           transcoder,
			Protocols:         protocols,
			ReadHeaderTimeout: 10 * time.Second,
		},
	}, nil
}

func (s *Server) Start(logger *slog.Logger) error {
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("listen %s: %w", s.addr, err)
	}
	go func() {
		logger.Info("otlp listening", "addr", s.addr)
		if err := s.srv.Serve(ln); err != nil && err != http.ErrServerClosed {
			logger.Error("serve exited", "err", err)
		}
	}()
	return nil
}

func (s *Server) Stop(ctx context.Context) {
	_ = s.srv.Shutdown(ctx)
}
