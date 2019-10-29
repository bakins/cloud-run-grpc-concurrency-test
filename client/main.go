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
	"bytes"
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"
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
	_ = pb.NewGreeterClient(conn)

	c := &client{
		target: u.String(),
		Client: http.Client{},
	}

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

type client struct {
	target string
	http.Client
}

func (c *client) SayHello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloReply, error) {
	var buf bytes.Buffer
	if err := Write(&buf, req, false); err != nil {
		return nil, err
	}

	//fmt.Println(c.target)

	// path is /<package>.<service>/<method>
	r, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.target+"/helloworld.Greeter/SayHello",
		bytes.NewBuffer(buf.Bytes()),
	)
	if err != nil {
		return nil, err
	}

	r.Header.Set("Content-Type", "application/grpc+proto")

	resp, err := c.Client.Do(r)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			body = []byte("unknown")
		}
		return nil, fmt.Errorf("unexpected status: %d %s %s", resp.StatusCode, resp.Status, string(body))
	}

	var helloResponse pb.HelloReply

	fmt.Println(resp.Header)
	fmt.Println(resp.Trailer)

	if err := Read(resp.Body, &helloResponse); err != nil {
		fmt.Println(resp.Header)
		fmt.Println(resp.Trailer)
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	// must read until EOF to ensure trailers are read.
	// there should be no data left before the trailers.
	//if _, err = resp.Body.Read([]byte{}); err != io.EOF {
	//	return nil, fmt.Errorf("unexpected error: %v", err)
	//}

	status := 0
	grpcStatus := resp.Trailer.Get("Grpc-Status")
	if grpcStatus == "" {
		// try header
		grpcStatus = resp.Header.Get("Grpc-Status")
	}

	if grpcStatus != "" {
		s, err := strconv.Atoi(grpcStatus)
		if err != nil {
			return nil, fmt.Errorf("failed to parse grpc-status %s: %v", grpcStatus, err)
		}
		status = s
	}

	if status != 0 {
		return nil, fmt.Errorf("unexpected grpc status %d %s", status, resp.Trailer.Get("Grpc-Message"))
	}

	return &helloResponse, nil
}
