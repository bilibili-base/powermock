package util

import (
	"fmt"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	jsoniter "github.com/json-iterator/go"
)

// Dump is used to dump object into stdout
func Dump(message interface{}) error {
	if data, ok := message.(proto.Message); ok {
		encoder := &jsonpb.Marshaler{
			Indent: "    ",
		}
		data, err := encoder.MarshalToString(data)
		if err != nil {
			return err
		}
		fmt.Println(data)
		return nil
	}
	data, err := jsoniter.MarshalToString(message)
	if err != nil {
		return err
	}
	fmt.Println(data)
	return nil
}
