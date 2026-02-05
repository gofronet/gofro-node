package delivery

import (
	"context"
	"fmt"
	"gofronet-foundation/gofro-node/config"
	apiv1 "gofronet-foundation/gofro-node/gen/api/v1"
	xraymanager "gofronet-foundation/gofro-node/xray_manager"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Handler struct {
	config      *config.Config
	xrayManager *xraymanager.XrayManager
	apiv1.UnimplementedXrayServiceServer
}

func NewHandler(config *config.Config, xrayManager *xraymanager.XrayManager) *Handler {
	return &Handler{
		config:      config,
		xrayManager: xrayManager,
	}
}

func (h *Handler) StartXray(context.Context, *apiv1.StartXrayRequest) (*apiv1.StartXrayResponse, error) {
	err := h.xrayManager.Start()
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("cannnot start xray: %s", err))
	}
	return &apiv1.StartXrayResponse{}, nil
}
func (h *Handler) UpdateXrayConfig(ctx context.Context, req *apiv1.UpdateXrayConfigRequest) (*apiv1.UpdateXrayConfigResponse, error) {

	if err := config.WriteConfigToFile(h.config.XrayConfigFile, req.NewConfig); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("cannot write config to file: %s", err))
	}

	if err := h.xrayManager.UpdateConfig(req.NewConfig); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("cannot update config in xray manager: %s", err))
	}

	if err := h.xrayManager.Restart(ctx); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("cannot restart xray to apply new config: %s", err))
	}

	return &apiv1.UpdateXrayConfigResponse{}, nil
}
func (h *Handler) RestartXray(ctx context.Context, req *apiv1.RestartXrayRequest) (*apiv1.RestartXrayResponse, error) {
	if err := h.xrayManager.Restart(ctx); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("cannot restart xray: %s", err))
	}
	return &apiv1.RestartXrayResponse{}, nil
}
func (h *Handler) StopXray(ctx context.Context, req *apiv1.StopXrayRequest) (*apiv1.StopXrayResponse, error) {
	if err := h.xrayManager.Stop(ctx); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("cannot stop xray: %s", err))
	}
	return &apiv1.StopXrayResponse{}, nil
}

func (h *Handler) GetXrayStatus(context.Context, *apiv1.GetXrayStatusRequest) (*apiv1.GetXrayStatusResponse, error) {
	return &apiv1.GetXrayStatusResponse{
		IsRunning: h.xrayManager.IsRunning(),
	}, nil
}
