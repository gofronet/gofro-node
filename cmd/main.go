package main

import (
	"context"
	"gofronet-foundation/gofro-node/config"
	"gofronet-foundation/gofro-node/delivery"
	apiv1 "gofronet-foundation/gofro-node/gen/api/v1"
	xraymanager "gofronet-foundation/gofro-node/xray_manager"
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
	manager.Start()

	handler := delivery.NewHandler(cfg, manager)

	grpcServer := grpc.NewServer()
	apiv1.RegisterXrayServiceServer(grpcServer, handler)

	if cfg.IsDevMode {
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
