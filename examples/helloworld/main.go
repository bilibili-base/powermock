// Copyright 2021 bilibili-base
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
