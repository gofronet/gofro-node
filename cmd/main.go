package main

import (
	"context"
	"gofronet-foundation/gofro-node/internal/config"
	"gofronet-foundation/gofro-node/internal/delivery"
	"gofronet-foundation/gofro-node/internal/delivery/interceptors"
	apiv1 "gofronet-foundation/gofro-node/internal/gen/api/v1"
	xraymanager "gofronet-foundation/gofro-node/internal/xray_manager"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalln(err)
	}

	xrayConfig, err := config.LoadXrayConfigFromFile(cfg.XrayConfigFile)
	if err != nil {
		log.Fatalln(err)
	}

	manager := xraymanager.NewXrayManager(xrayConfig, cfg.XrayCorePath)
	if err := manager.Start(); err != nil {
		log.Fatalln(err)
	}

	handler := delivery.NewXrayManagmentService(cfg, manager)

	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(interceptors.UnaryLogging()))
	apiv1.RegisterXrayServiceServer(grpcServer, handler)

	if cfg.IsDevMode {
		log.Println("dev mode enabled, reflection registered")
		reflection.Register(grpcServer)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		lis, err := net.Listen("tcp", ":50051")
		if err != nil {
			return err
		}

		log.Printf("gRPC server listening on %s", lis.Addr().String())
		return grpcServer.Serve(lis)
	})

	g.Go(func() error {
		<-ctx.Done()
		log.Println("stopping gRPC server")
		grpcServer.GracefulStop()
		return nil
	})

	if err := g.Wait(); err != nil {
		log.Printf("server stopped with error: %v", err)
	} else {
		log.Println("server stopped gracefully")
	}

}
