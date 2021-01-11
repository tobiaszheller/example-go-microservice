package main

import (
	"database/sql"
	"net"

	_ "github.com/go-sql-driver/mysql"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_logrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/kelseyhightower/envconfig"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	pb "github.com/tobiaszheller/example-go-microservice/service-users/proto"
	"github.com/tobiaszheller/example-go-microservice/service-users/pubsubmock"
	"github.com/tobiaszheller/example-go-microservice/service-users/rpc"
	"github.com/tobiaszheller/example-go-microservice/service-users/store"
	"github.com/tobiaszheller/example-go-microservice/service-users/telemetry"
)

type config struct {
	GRPCAddr      string `envconfig:"GRPC_ADDR" default:":18082"`
	TelemetryAddr string `envconfig:"TELEMETRY_ADDR" default:":18083"`
	DBDSN         string `envconfig:"DB_DSN" default:"user:password@tcp(127.0.0.1:23306)/test"`
}

func main() {
	var cfg config
	if err := envconfig.Process("", &cfg); err != nil {
		log.Fatal(err)
	}

	store := store.New(mustConnectDB(cfg))
	service := rpc.New(store, pubsubmock.New())

	grpcServer, lis := mustSetupGRPC(cfg, func(s *grpc.Server) {
		pb.RegisterUsersServer(s, service)
		grpc_prometheus.Register(s)
	})
	go func() {
		if err := telemetry.Serve(cfg.TelemetryAddr); err != nil {
			log.Fatal(err)
		}
	}()
	if err := runGRPC(grpcServer, lis); err != nil {
		log.Fatal(err)
	}
}

// TODO: grpc helpers should be moved into some helper lib.
func mustSetupGRPC(cfg config, registerFn func(*grpc.Server)) (*grpc.Server, net.Listener) {
	lis, err := net.Listen("tcp", cfg.GRPCAddr)
	if err != nil {
		log.Fatalf("Failed to start listener %v", err)
	}
	log.Infof("Will setup gRPC server at: %s", lis.Addr().String())
	// TODO: in real life implementation following options shoud be passed:
	// - TLS credentails
	// - greaceful shutdown
	// - interceptor for passing trace_id from incomming request
	// - interceptor for panic recovey
	// - interceptor for authorization
	grpcServer := grpc.NewServer(
		grpc_middleware.WithUnaryServerChain(
			grpc_logrus.UnaryServerInterceptor(log.NewEntry(log.New())),
		),
	)

	registerFn(grpcServer)
	return grpcServer, lis
}

func runGRPC(srv *grpc.Server, lis net.Listener) error {
	log.Info("Starting gRPC server")
	defer lis.Close()
	return srv.Serve(lis)
}

func mustConnectDB(cfg config) *sql.DB {
	db, err := sql.Open("mysql", cfg.DBDSN)
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping to DB: %v", err)
	}
	return db
}
