package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/go-redis/redis"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	"github.com/rakyll/statik/fs"
	"github.com/techschool/simplebank/api"
	db "github.com/techschool/simplebank/db/sqlc"
	_ "github.com/techschool/simplebank/doc/statik"
	"github.com/techschool/simplebank/gapi"
	"github.com/techschool/simplebank/mail"
	"github.com/techschool/simplebank/pb"
	"github.com/techschool/simplebank/util"
	"github.com/techschool/simplebank/worker"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
)


func main() {

	config, err := util.LoadConfig(".")
	fmt.Println("the current db address is:",config.DBSource)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot load config:")
	}
	if config.Environment == "development" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	}

	connPool, err := pgxpool.New(context.Background(), config.DBSource)
	if err != nil{
		log.Fatal().Err(err)
	}

	// run db migration
	runDBMigration(config.MigrationURL,config.DBSource)

	store := db.NewStore(connPool)

	redisOpt := asynq.RedisClientOpt{
		Addr: config.RedisAddress,
	}

	taskDistributor := worker.NewRedisTaskDistributor(redisOpt)


	//asynq:1.create task by newTask
	//2.add to a queue
	//step 1, and 2, in one file
	//3.create server and servermux, add tasks to handleFunc and handle
	go runTaskProcessor(config,redisOpt, store)
	// go testRedis(redisOpt)
	go runGatewayServer(config, store, taskDistributor)
	runGrpcServer(config,store, taskDistributor)
}

func testRedis(redisOpt asynq.RedisClientOpt){
	log.Info().Msgf("here is the redisOptValue: %s", redisOpt.Addr)
	client := redis.NewClient(&redis.Options{
        Addr:     "localhost:6379",
        Password: "",
        DB:       0,
    })
	// ctx := context.TODO()
    if pong := client.Ping(); pong.String() != "ping: PONG" {
        log.Info().Msgf("-------------Error connection redis ----------:%s", pong)
    }else{
		log.Info().Msg("redis is successfully running")

	}

}

func runDBMigration(migrationURl string, dbSource string){
	 migration, err := migrate.New(migrationURl,dbSource)
	 if err != nil {
		log.Fatal().Err(err).Msg("cannot create new migrate instance:")
	 }

	 if err = migration.Up(); err != nil && err != migrate.ErrNoChange{
		log.Fatal().Err(err).Msg("failed to run migrate up:")
	 }

	 log.Info().Msg("db migrated successfully")
}

func runTaskProcessor(config util.Config,redisOpt asynq.RedisClientOpt, store db.Store){
	mailer := mail.NewGmailSender(config.EmailSenderName, config.EmailSenderAddress, config.EmailSenderPassword)
	taskProcessor := worker.NewRedisTaskProcessor(redisOpt, store, mailer)
	log.Info().Msg("start task processor")
	err := taskProcessor.Start()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to start task processor")
	}
}

func runGrpcServer(config util.Config, store db.Store,taskDistributor worker.TaskDistributor){
  	server, err := gapi.NewServer(config,store,taskDistributor)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create server:")
	}
	grpcLogger := grpc.UnaryInterceptor(gapi.GrpcLogger)
	grpcServer := grpc.NewServer(grpcLogger)
	pb.RegisterSimplebankServer(grpcServer, server)
	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp",config.GRPCServerAddress)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create listener:")
	}

	log.Info().Msgf("start gRPC server at %s", listener.Addr().String())
	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot start gRPC server:")
	}

}



func runGatewayServer(config util.Config, store db.Store, taskDistributor worker.TaskDistributor){
	server, err := gapi.NewServer(config,store,taskDistributor)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create server:")
	}


		jsonOption := runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
			MarshalOptions: protojson.MarshalOptions{
				UseProtoNames: true,
			},
			UnmarshalOptions: protojson.UnmarshalOptions{
				DiscardUnknown: true,
			},
		})
	
	

	grpcMux := runtime.NewServeMux(jsonOption)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err = pb.RegisterSimplebankHandlerServer(ctx,grpcMux,server)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot register handler server:")
	}

	//receive http request from client
	mux := http.NewServeMux()
	mux.Handle("/",grpcMux)

	// fs := http.FileServer(http.Dir("./doc/swagger"))
	statikFS, err := fs.New()
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create statik fs")
	}
	swaggerHandler := http.StripPrefix("/swagger/",http.FileServer(statikFS))
	mux.Handle("/swagger/",swaggerHandler)

	listener, err := net.Listen("tcp",config.HTTPServerAddress)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create listener:")
	}

	log.Info().Msgf("start HTTP gateway server at %s", listener.Addr().String())
	handler := gapi.HttpLogger(mux)
	err = http.Serve(listener,handler)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot start HTTP gateway server:")
	}

}

func runGinServer(config util.Config, store db.Store){
	server, err := api.NewServer(config,store)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create server:")
	}

	err = server.Start(config.HTTPServerAddress)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot start server")
	}
}


