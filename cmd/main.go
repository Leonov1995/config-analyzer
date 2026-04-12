package main

import (
	"os"
	"runtime"

	"config-analyzer/internal/analyzer"
	"config-analyzer/internal/cli"
	"config-analyzer/internal/config"
	grpcSrv "config-analyzer/internal/grpc"
	httpSrv "config-analyzer/internal/http"
	"config-analyzer/internal/rule"
	"config-analyzer/internal/service"
)

func main() {
	rules := rule.NewDefaultRules()
	analyz := analyzer.New(rules)
	cfgLoader := config.NewLoader()
	workers := runtime.NumCPU()

	svc := service.New(analyz, cfgLoader, workers)

	httpHandler := httpSrv.NewHandler(analyz)
	httpRouter := httpSrv.NewRouter(httpHandler)
	httpServer := httpSrv.NewServer(httpRouter)

	grpcServer := grpcSrv.NewServer(analyz)

	rootCmd := cli.NewRootCmd(svc, httpServer, grpcServer)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
