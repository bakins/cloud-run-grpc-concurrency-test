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

// Package main implements a client for Greeter service.
package main

import (
	"context"
	"crypto/tls"
	"flag"
	"log"
	"net"
	"net/url"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"
)

const (
	defaultName = "world"
)

func main() {
	var (
		address    string
		workers    int
		iterations int
		useHTTP    bool
	)
	flag.StringVar(&address, "url", "http://localhost:8080", "url endpoint - must include schema")
	flag.IntVar(&workers, "workers", 1, "concurrent workers")
	flag.IntVar(&iterations, "iterations", 1, "iterations per worker")
	flag.BoolVar(&useHTTP, "http", false, "non-grpc HTTP test")
	flag.Parse()

	u, err := url.Parse(address)
	if err != nil {
		log.Fatalf("failed to parse url: %v", err)
	}

	var options []grpc.DialOption

	if u.Scheme == "https" {
		options = append(options, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})))
	} else {
		options = append(options, grpc.WithInsecure())
	}

	port := u.Port()
	if port == "" {
		if u.Scheme == "https" {
			port = "443"
		} else {
			port = "80"
		}
	}
	addr := net.JoinHostPort(u.Hostname(), port)
	conn, err := grpc.Dial(addr, options...)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewGreeterClient(conn)

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()

			ctx := context.Background()
			for j := 0; j < iterations; j++ {
				j := j
				_, err := c.SayHello(ctx, &pb.HelloRequest{Name: defaultName})
				if err != nil {
					log.Printf("could not greet: %d %d %v", i, j, err)
				}
				//time.Sleep(time.Millisecond * 100)
				//log.Printf("%d %d %s", i, j, r.GetMessage())
			}
		}()
	}
	wg.Wait()
}
