package interceptors

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

func UnaryLogging() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {

		start := time.Now()
		resp, err := handler(ctx, req)

		st, _ := status.FromError(err)
		log.Printf("method=%s dur=%s code=%s err=%v",
			info.FullMethod, time.Since(start), st.Code(), err,
		)
		return resp, err
	}
}
