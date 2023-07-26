package main

import (
	"database/sql"
	"flag"
	"log"
	"net"

	"git.neds.sh/matty/entain/racing/db"
	"git.neds.sh/matty/entain/racing/proto/racing"
	"git.neds.sh/matty/entain/racing/service"
	"google.golang.org/grpc"
)

var (
	grpcEndpoint = flag.String("grpc-racing-endpoint", "localhost:9000", "gRPC racing server endpoint")
)

func main() {
	flag.Parse()

	if err := run(); err != nil {
		log.Fatalf("failed running grpc server: %s\n", err)
	}
}

func run() error {
	conn, err := net.Listen("tcp", ":9000")
	if err != nil {
		return err
	}

	racingDB, err := sql.Open("sqlite3", "./db/racing.db")
	if err != nil {
		return err
	}

	racesRepo := db.NewRacesRepo(racingDB)
	if err := racesRepo.Init(); err != nil {
		return err
	}

	grpcServer := grpc.NewServer()

	racing.RegisterRacingServer(
		grpcServer,
		service.NewRacingService(
			racesRepo,
		),
	)

	log.Printf("gRPC racing server listening on: %s\n", *grpcEndpoint)

	if err := grpcServer.Serve(conn); err != nil {
		return err
	}

	return nil
}
