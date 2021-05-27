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
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/bilibili-base/powermock/examples/advanced/apis"
	"github.com/bilibili-base/powermock/pkg/util/logger"
)

func RequestWithUid(log logger.Logger, client apis.GreeterClient, uid string) {
	log.LogInfo(map[string]interface{}{
		"uid": uid,
	}, "start to call mock server")
	var header, trailer metadata.MD
	ctx := metadata.AppendToOutgoingContext(context.TODO(), "uid", uid)
	startTime := time.Now()
	resp, err := client.Hello(ctx, &apis.HelloRequest{
		Timestamp: uint64(time.Now().Unix()),
		Message:   "hello",
	}, grpc.Header(&header), grpc.Trailer(&trailer))
	if err != nil {
		log.LogFatal(nil, "failed to call: %s", err)
	}
	log.LogInfo(nil, "[elapsed] %d ms", time.Since(startTime).Milliseconds())
	log.LogInfo(nil, "[headers] %+v", header)
	log.LogInfo(nil, "[trailer] %+v", trailer)
	log.LogInfo(nil, "[response] %+v", resp)
}

func main() {
	log := logger.NewDefault("main")
	log.LogInfo(nil, "starting call mock server")
	conn, err := grpc.Dial("127.0.0.1:30002", grpc.WithInsecure())
	if err != nil {
		log.LogFatal(nil, "failed to dial: %s", err)
	}
	client := apis.NewGreeterClient(conn)

	RequestWithUid(log, client, "20")
	fmt.Println(strings.Repeat("-", 20))
	RequestWithUid(log, client, "2233")
}
