package main

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/storyicon/powermock/apis/v1alpha1"
	"github.com/storyicon/powermock/examples/greeter/apis"
	"github.com/storyicon/powermock/pkg/util/logger"
)

func addMockAPIBygRPC() error {
	conn, err := grpc.Dial("127.0.0.1:30000", grpc.WithInsecure())
	if err != nil {
		return err
	}
	client := v1alpha1.NewMockClient(conn)
	_, err = client.SaveMockAPI(context.TODO(), &v1alpha1.SaveMockAPIRequest{
		Data: &v1alpha1.MockAPI{
			UniqueKey: "1",
			Path:      "/examples.greeter.api.Greeter/Hello",
			Method:    "POST",
			Cases: []*v1alpha1.MockAPI_Case{
				{
					Condition: &v1alpha1.MockAPI_Condition{
						Condition: &v1alpha1.MockAPI_Condition_Simple{
							Simple: &v1alpha1.MockAPI_Condition_SimpleCondition{
								Items: []*v1alpha1.MockAPI_Condition_SimpleCondition_Item{
									{
										OperandX: "$request.header.uid",
										Operator: ">=",
										OperandY: "5",
										Opposite: false,
									},
								},
							},
						},
					},
					Response: &v1alpha1.MockAPI_Response{
						Response: &v1alpha1.MockAPI_Response_Simple{
							Simple: &v1alpha1.MockAPI_Response_SimpleResponse{
								Code: 0,
								Header: map[string]string{
									"username": "$mock.name",
									"grpc-ddd": "OK",
								},
								Trailer: map[string]string{
									"qqq": "111",
								},
								Body: `
								{"timestamp": "111", "message": "{{ $mock.url }}", "amount": {{ $mock.price }} }
                            `,
							},
						},
					},
				},
			},
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func main() {
	if err := addMockAPIBygRPC(); err != nil {
		panic(err)
	}

	log := logger.NewDefault("main")
	log.LogInfo(nil, "starting call mock server")
	conn, err := grpc.Dial("127.0.0.1:30002", grpc.WithInsecure())
	if err != nil {
		log.LogFatal(nil, "failed to dial: %s", err)
	}
	client := apis.NewGreeterClient(conn)
	var header, trailer metadata.MD

	ctx := metadata.AppendToOutgoingContext(context.TODO(), "uid", "20")
	resp, err := client.Hello(ctx, &apis.HelloRequest{
		Timestamp: uint64(time.Now().Unix()),
		Message:   "hello",
	}, grpc.Header(&header), grpc.Trailer(&trailer))
	if err != nil {
		log.LogFatal(nil, "failed to call: %s", err)
	}
	log.LogInfo(nil, "[headers] %+v", header)
	log.LogInfo(nil, "[trailer] %+v", trailer)
	log.LogInfo(nil, "[response] %+v", resp)
}
