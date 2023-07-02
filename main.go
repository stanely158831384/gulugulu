package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	_ "github.com/lib/pq"
	"github.com/rakyll/statik/fs"
	"github.com/techschool/simplebank/api"
	db "github.com/techschool/simplebank/db/sqlc"
	_ "github.com/techschool/simplebank/doc/statik"
	"github.com/techschool/simplebank/gapi"
	"github.com/techschool/simplebank/pb"
	"github.com/techschool/simplebank/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
)


func main() {
	config, err := util.LoadConfig(".")
	fmt.Println("the current db address is:",config.DBSource)
	if err != nil {
		log.Fatal("cannot load config:", err)
	}
	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil{
		log.Fatal(err)
	}

	// run db migration
	runDBMigration(config.MigrationURL,config.DBSource)

	store := db.NewStore(conn)
	go runGatewayServer(config, store)
	runGrpcServer(config,store)
}

func runDBMigration(migrationURl string, dbSource string){
	 migration, err := migrate.New(migrationURl,dbSource)
	 if err != nil {
		log.Fatal("cannot create new migrate instance:", err)
	 }

	 if err = migration.Up(); err != nil && err != migrate.ErrNoChange{
		log.Fatal("failed to run migrate up:",err)
	 }

	 log.Println("db migrated successfully")
}

func runGrpcServer(config util.Config, store db.Store){
  	server, err := gapi.NewServer(config,store)
	if err != nil {
		log.Fatal("cannot create server:", err)
	}
	grpcLogger := grpc.UnaryInterceptor(gapi.GrpcLogger)
	grpcServer := grpc.NewServer(grpcLogger)
	pb.RegisterSimplebankServer(grpcServer, server)
	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp",config.GRPCServerAddress)
	if err != nil {
		log.Fatal("cannot create listener:",err)
	}

	log.Printf("start gRPC server at %s", listener.Addr().String())
	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatal("cannot start gRPC server:",err)
	}

}



func runGatewayServer(config util.Config, store db.Store){
	server, err := gapi.NewServer(config,store)
	if err != nil {
		log.Fatal("cannot create server:", err)
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
		log.Fatal("cannot register handler server:",err)
	}

	//receive http request from client
	mux := http.NewServeMux()
	mux.Handle("/",grpcMux)

	// fs := http.FileServer(http.Dir("./doc/swagger"))
	statikFS, err := fs.New()
	if err != nil {
		log.Fatal("cannot create statik fs",err)
	}
	swaggerHandler := http.StripPrefix("/swagger/",http.FileServer(statikFS))
	mux.Handle("/swagger/",swaggerHandler)

	listener, err := net.Listen("tcp",config.HTTPServerAddress)
	if err != nil {
		log.Fatal("cannot create listener:",err)
	}

	log.Printf("start HTTP gateway server at %s", listener.Addr().String())
	err = http.Serve(listener,mux)
	if err != nil {
		log.Fatal("cannot start HTTP gateway server:",err)
	}

}

func runGinServer(config util.Config, store db.Store){
	server, err := api.NewServer(config,store)
	if err != nil {
		log.Fatal("cannot create server:", err)
	}

	err = server.Start(config.HTTPServerAddress)
	if err != nil {
		log.Fatal("cannot start server", err)
	}
}