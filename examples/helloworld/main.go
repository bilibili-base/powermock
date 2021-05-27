package main

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/bilibili-base/powermock/examples/helloworld/apis"
)

func main() {
	fmt.Println("starting call mock server")
	conn, err := grpc.Dial("127.0.0.1:30002", grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	client := apis.NewGreeterClient(conn)

	var header, trailer metadata.MD
	startTime := time.Now()
	resp, err := client.Hello(context.TODO(), &apis.HelloRequest{
		Message: "hi",
	}, grpc.Header(&header), grpc.Trailer(&trailer))
	if err != nil {
		panic(err)
	}
	fmt.Printf("[elapsed] %d ms \r\n", time.Since(startTime).Milliseconds())
	fmt.Printf("[headers] %+v \r\n", header)
	fmt.Printf("[trailer] %+v \r\n", trailer)
	fmt.Printf("[response] %+v \r\n", resp.String())
}
