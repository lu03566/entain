package main

import (
	"context"
	"flag"
	"log"
	"net/http"

	"git.neds.sh/matty/entain/api/proto/racing"
	"git.neds.sh/matty/entain/api/proto/sports"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

var (
	apiEndpoint  = flag.String("api-endpoint", "localhost:8000", "API endpoint")
	racingEndpoint = flag.String("racing-endpoint", "localhost:9000", "racing server endpoint")
	sportsEndpoint = flag.String("sports-endpoint", "localhost:10000", "sports server endpoint")

)

func main() {
	flag.Parse()

	if err := run(); err != nil {
		log.Printf("failed running api server: %s\n", err)
	}
}

func run() error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := runtime.NewServeMux()
	if racingErr := racing.RegisterRacingHandlerFromEndpoint(
		ctx,
		mux,
		*racingEndpoint,
		[]grpc.DialOption{grpc.WithInsecure()},
	); racingErr != nil {
		return racingErr
	}

	if sportsErr := sports.RegisterSportsHandlerFromEndpoint(
		ctx,
		mux,
		*sportsEndpoint,
		[]grpc.DialOption{grpc.WithInsecure()},
	); sportsErr != nil {
		return sportsErr
	}

	log.Printf("API server listening on: %s\n", *apiEndpoint)

	return http.ListenAndServe(*apiEndpoint, mux)
}
