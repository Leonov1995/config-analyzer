package service

import (
	"config-analyzer/internal/model"
	"errors"

	pb "config-analyzer/internal/grpc/proto"
)

func ProtoToConfig(pbCfg *pb.Config) (*model.Config, error) {
	if pbCfg == nil {
		return nil, errors.New("config is nil")
	}

	cfg := &model.Config{}

	if pbCfg.Server != nil {
		cfg.Server.Host = pbCfg.Server.Host
		cfg.Server.Port = int(pbCfg.Server.Port)
	}

	if pbCfg.Auth != nil {
		cfg.Auth.Password = pbCfg.Auth.Password
	}

	if pbCfg.Log != nil {
		cfg.Log.Output = pbCfg.Log.Output
		cfg.Log.Level = pbCfg.Log.Level
	}

	if pbCfg.Tls != nil {
		cfg.TLS.Enabled = pbCfg.Tls.Enabled
	}

	if pbCfg.Storage != nil {
		cfg.Storage.DigestAlgorithm = pbCfg.Storage.DigestAlgorithm
	}

	return cfg, nil
}
