package grpc

import (
	"context"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"config-analyzer/internal/service"
	"config-analyzer/internal/server"
	pb "config-analyzer/internal/grpc/proto"
)

type Server struct {
	pb.UnimplementedAnalyzerServiceServer
	grpcServer *grpc.Server
	analyzer   service.Analyzer
}

func NewServer(a service.Analyzer) *Server {
	return &Server{
		grpcServer: grpc.NewServer(),
		analyzer:   a,
	}
}

func (s *Server) Run(addr string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	reflection.Register(s.grpcServer)
	pb.RegisterAnalyzerServiceServer(s.grpcServer, s)

	errCh := make(chan error, 1)
	go func() {
		log.Printf("gRPC server started on %s", addr)
		if err := s.grpcServer.Serve(lis); err != nil {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-server.WaitSignalChan():
	}

	return server.GracefulShutdown("gRPC", func(ctx context.Context) error {
		s.grpcServer.GracefulStop()
		return nil
	})
}

func (s *Server) Analyze(ctx context.Context, req *pb.ConfigRequest) (*pb.AnalyzeResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	cfg, err := service.ProtoToConfig(req.GetConfig())
	if err != nil {
		return &pb.AnalyzeResponse{
			Status:       pb.Status_error,
			ErrorMessage: err.Error(),
		}, nil
	}

	issues := s.analyzer.Analyze(cfg)

	resp := &pb.AnalyzeResponse{
		Status: pb.Status_clean,
	}

	if len(issues) > 0 {
		resp.Status = pb.Status_issues
		resp.Issues = make([]*pb.Issue, len(issues))
		for i, issue := range issues {
			resp.Issues[i] = &pb.Issue{
				Severity:       string(issue.Severity),
				Message:        issue.Message,
				Recommendation: issue.Recommendation,
				Path:           issue.Path,
			}
		}
	}

	return resp, nil
}
