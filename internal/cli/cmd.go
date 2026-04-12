package cli

import (
	"github.com/spf13/cobra"

	grpcSrv "config-analyzer/internal/grpc"
	"config-analyzer/internal/http"
	"config-analyzer/internal/service"
)

func NewRootCmd(svc *service.Service, httpServer *http.Server, grpcServer *grpcSrv.Server) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "config-analyzer",
		Short: "Analyzes YAML/JSON configuration files for security issues",
	}

	rootCmd.AddCommand(newAnalyzeCmd(svc))
	rootCmd.AddCommand(newServeHTTPCmd(httpServer))
	rootCmd.AddCommand(newServeGRPCCmd(grpcServer))

	rootCmd.SilenceUsage = true

	return rootCmd
}

func newAnalyzeCmd(svc *service.Service) *cobra.Command {
	var (
		format string
		silent bool
		stdin  bool
	)

	cmd := &cobra.Command{
		Use:   "analyze [path]",
		Short: "Analyze config file(s) for security issues",
		Long: `Analyze one or more configuration files for security issues.
If [path] is a directory, it is scanned recursively.

Examples:
  config-analyzer analyze config.json
  config-analyzer analyze configs/
  cat config.yaml | config-analyzer analyze --stdin
  printf '{"server":{"port":8080}}' | config-analyzer analyze --stdin --format json`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			app := NewApp(svc)
			return app.Run(stdin, args, format, silent)
		},
	}

	cmd.Flags().BoolVarP(&silent, "silent", "s", false, "Don't exit with error if issues found")
	cmd.Flags().BoolVar(&stdin, "stdin", false, "Read config from stdin instead of a file")
	cmd.Flags().StringVarP(&format, "format", "f", "", "Input format: json|yaml (auto-detected if omitted)")

	return cmd
}

func newServeHTTPCmd(httpServer *http.Server) *cobra.Command {
	var addr string

	cmd := &cobra.Command{
		Use:   "serve-http",
		Short: "Start REST API server for config analysis",
		Long: `Start an HTTP server that exposes a POST /analyze endpoint.
Send JSON or YAML config in the request body and receive issues as JSON.

Example:
  curl -X POST http://localhost:8080/analyze \
    -H "Content-Type: application/json" \
    -d '{"server":{"host":"0.0.0.0"},"tls":{"enabled":false}}'`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return httpServer.Run(addr)
		},
	}

	cmd.Flags().StringVarP(&addr, "addr", "a", ":8080", "HTTP listen address")

	return cmd
}

func newServeGRPCCmd(grpcServer *grpcSrv.Server) *cobra.Command {
	var addr string

	cmd := &cobra.Command{
		Use:   "serve-grpc",
		Short: "Start gRPC server for config analysis",
		Long: `Start a gRPC server that exposes the AnalyzerService.
Use any gRPC client (e.g., grpcurl) to send analysis requests.

Example:
  grpcurl -plaintext -d '{"config":{...}}' localhost:50051 analyzer.AnalyzerService/Analyze`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return grpcServer.Run(addr)
		},
	}

	cmd.Flags().StringVarP(&addr, "addr", "a", ":50051", "gRPC listen address")

	return cmd
}
