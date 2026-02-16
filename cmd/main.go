package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	xraymanagmentv1 "github.com/gofronet/gofro-node/gen/go/api/xray_managment/v1"
	"github.com/gofronet/gofro-node/internal/config"
	grpcinterceptors "github.com/gofronet/gofro-node/internal/grpc_interceptors"
	xraymanager "github.com/gofronet/gofro-node/internal/xray_manager"

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

	// xrayConn, err := xrayconn.NewXrayConn(cfg.XrayApiAddress)
	// if err != nil {
	// 	log.Fatalln(err)
	// }

	xrayManagmentGrpcService := xraymanager.NewXrayManagmentService(cfg, manager)

	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(grpcinterceptors.UnaryLogging()))
	xraymanagmentv1.RegisterXrayManagmentServiceServer(grpcServer, xrayManagmentGrpcService)

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
