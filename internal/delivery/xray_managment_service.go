package delivery

import (
	"context"
	"fmt"
	"gofronet-foundation/gofro-node/internal/config"
	apiv1 "gofronet-foundation/gofro-node/internal/gen/api/v1"
	xraymanager "gofronet-foundation/gofro-node/internal/xray_manager"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type XrayManagmentService struct {
	config      *config.Config
	xrayManager *xraymanager.XrayManager
	apiv1.UnimplementedXrayServiceServer
}

func NewXrayManagmentService(config *config.Config, xrayManager *xraymanager.XrayManager) *XrayManagmentService {
	return &XrayManagmentService{
		config:      config,
		xrayManager: xrayManager,
	}
}

func (h *XrayManagmentService) StartXray(context.Context, *apiv1.StartXrayRequest) (*apiv1.StartXrayResponse, error) {
	err := h.xrayManager.Start()
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("cannnot start xray: %s", err))
	}
	return &apiv1.StartXrayResponse{}, nil
}

func (h *XrayManagmentService) UpdateXrayConfig(ctx context.Context, req *apiv1.UpdateXrayConfigRequest) (*apiv1.UpdateXrayConfigResponse, error) {

	if err := config.WriteConfigToFile(h.config.XrayConfigFile, req.NewConfig); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("cannot write config to file: %s", err))
	}

	if err := h.xrayManager.UpdateConfig(req.NewConfig); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("cannot update config in xray manager: %s", err))
	}

	return &apiv1.UpdateXrayConfigResponse{}, nil
}

func (h *XrayManagmentService) RestartXray(ctx context.Context, req *apiv1.RestartXrayRequest) (*apiv1.RestartXrayResponse, error) {
	if err := h.xrayManager.Restart(ctx); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("cannot restart xray: %s", err))
	}
	return &apiv1.RestartXrayResponse{}, nil
}
func (h *XrayManagmentService) StopXray(ctx context.Context, req *apiv1.StopXrayRequest) (*apiv1.StopXrayResponse, error) {
	if err := h.xrayManager.Stop(ctx); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("cannot stop xray: %s", err))
	}
	return &apiv1.StopXrayResponse{}, nil
}

func (h *XrayManagmentService) GetNodeInfo(context.Context, *apiv1.GetNodeInfoRequest) (*apiv1.GetNodeInfoResponse, error) {
	return &apiv1.GetNodeInfoResponse{
		XrayRunning: h.xrayManager.IsRunning(),
		NodeName:    h.config.NodeName,
	}, nil
}

func (h *XrayManagmentService) GetCurrentConfig(context.Context, *apiv1.GetCurrentConfigRequest) (*apiv1.GetCurrentConfigResponse, error) {
	return &apiv1.GetCurrentConfigResponse{
		CurrentConfig: h.xrayManager.GetCurrentConfig(),
	}, nil
}
