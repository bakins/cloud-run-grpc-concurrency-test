/*
 *
 * Copyright 2015 gRPC authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

//go:generate protoc -I ../helloworld --go_out=plugins=grpc:../helloworld ../helloworld/helloworld.proto

// Package main implements a server for Greeter service.
package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"time"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"
)

// server is used to implement helloworld.GreeterServer.
type server struct {
	pb.UnimplementedGreeterServer
}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	n := rand.Intn(90)
	time.Sleep(time.Duration(100+n) * time.Millisecond)
	return &pb.HelloReply{Message: "Hello " + in.GetName()}, nil
}

func handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := rand.Intn(90)
		time.Sleep(time.Duration(10+n) * time.Millisecond)
		_, _ = fmt.Fprintln(w, "OK")
	})
}

func main() {
	log.Println("starting container")

	rand.Seed(time.Now().UnixNano())
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	//l := &listener{lis}

	s := grpc.NewServer()

	http.Handle("/helloworld.Greeter/", s)
	pb.RegisterGreeterServer(s, &server{})

	http.Handle("/ping", handler())

	l := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("remote addr %s", r.RemoteAddr)
		http.DefaultServeMux.ServeHTTP(w, r)
	})

	h := h2c.NewHandler(l, &http2.Server{})

	if err := http.Serve(lis, h); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

type listener struct {
	net.Listener
}

func (l *listener) Accept() (net.Conn, error) {
	c, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}
	//log.Printf("accept %s %s", c.RemoteAddr(), c.LocalAddr())
	return c, nil
}
