package xraymanager

import (
	"context"
	"fmt"

	xraymanagmentv1 "gofronet-foundation/gofro-node/gen/go/api/xray_managment/v1"
	"gofronet-foundation/gofro-node/internal/config"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GrpcXrayManagmentService struct {
	config      *config.Config
	xrayManager *XrayManager
	xraymanagmentv1.UnimplementedXrayManagmentServiceServer
}

func NewXrayManagmentService(config *config.Config, xrayManager *XrayManager) *GrpcXrayManagmentService {
	return &GrpcXrayManagmentService{
		config:      config,
		xrayManager: xrayManager,
	}
}

func (h *GrpcXrayManagmentService) StartXray(context.Context, *xraymanagmentv1.StartXrayRequest) (*xraymanagmentv1.StartXrayResponse, error) {
	err := h.xrayManager.Start()
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("cannnot start xray: %s", err))
	}
	return &xraymanagmentv1.StartXrayResponse{}, nil
}

func (h *GrpcXrayManagmentService) UpdateXrayConfig(ctx context.Context, req *xraymanagmentv1.UpdateXrayConfigRequest) (*xraymanagmentv1.UpdateXrayConfigResponse, error) {

	if err := config.WriteConfigToFile(h.config.XrayConfigFile, req.NewConfig); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("cannot write config to file: %s", err))
	}

	if err := h.xrayManager.UpdateConfig(req.NewConfig); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("cannot update config in xray manager: %s", err))
	}

	return &xraymanagmentv1.UpdateXrayConfigResponse{}, nil
}

func (h *GrpcXrayManagmentService) RestartXray(ctx context.Context, req *xraymanagmentv1.RestartXrayRequest) (*xraymanagmentv1.RestartXrayResponse, error) {
	if err := h.xrayManager.Restart(ctx); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("cannot restart xray: %s", err))
	}
	return &xraymanagmentv1.RestartXrayResponse{}, nil
}
func (h *GrpcXrayManagmentService) StopXray(ctx context.Context, req *xraymanagmentv1.StopXrayRequest) (*xraymanagmentv1.StopXrayResponse, error) {
	if err := h.xrayManager.Stop(ctx); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("cannot stop xray: %s", err))
	}
	return &xraymanagmentv1.StopXrayResponse{}, nil
}

func (h *GrpcXrayManagmentService) GetNodeInfo(context.Context, *xraymanagmentv1.GetNodeInfoRequest) (*xraymanagmentv1.GetNodeInfoResponse, error) {
	return &xraymanagmentv1.GetNodeInfoResponse{
		XrayRunning: h.xrayManager.IsRunning(),
		NodeName:    h.config.NodeName,
	}, nil
}

func (h *GrpcXrayManagmentService) GetCurrentConfig(context.Context, *xraymanagmentv1.GetCurrentConfigRequest) (*xraymanagmentv1.GetCurrentConfigResponse, error) {
	return &xraymanagmentv1.GetCurrentConfigResponse{
		CurrentConfig: h.xrayManager.GetCurrentConfig(),
	}, nil
}
