package xrayconn

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewXrayConn(addr string) (*grpc.ClientConn, error) {
	xrayConn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return xrayConn, nil

}
