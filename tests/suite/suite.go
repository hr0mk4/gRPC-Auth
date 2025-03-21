package suite

import (
	"context"
	"net"
	"strconv"
	"testing"

	"github.com/hr0mk4/grpc_auth/internal/config"
	authv1 "github.com/hr0mk4/protos_auth/gen/go/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	grpchost = "localhost"
)

type Suite struct {
	*testing.T
	Cfg        *config.Config
	AuthClient authv1.AuthClient
}

func New(t *testing.T) (context.Context, *Suite) {
	t.Helper()
	t.Parallel()

	cfg := config.MustLoadByPath("../config/local_tests.yaml")

	ctx, cancelCtx := context.WithTimeout(context.Background(), cfg.GRPC.Timeout)

	t.Cleanup(func() {
		t.Helper()
		cancelCtx()
	})
	cc, err := grpc.NewClient(grpcAdress(cfg),
		grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		t.Fatalf("grpc server connection error: %v", err)
	}

	return ctx, &Suite{t, cfg, authv1.NewAuthClient(cc)}
}

func grpcAdress(cfg *config.Config) string {
	return net.JoinHostPort(grpchost, strconv.Itoa(cfg.GRPC.Port))
}
