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

package core

import (
	"context"
	"reflect"
	"testing"

	"github.com/bilibili-base/powermock/pkg/interact"
)

func TestMatchRequestByJavascript(t *testing.T) {
	type args struct {
		ctx     context.Context
		request *interact.Request
		script  string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "test0",
			args: args{
				ctx: context.Background(),
				request: &interact.Request{
					Method: "POST",
					Header: map[string]string{
						"x-user-id": "320482",
					},
				},
				script: `
					(function(){
						if (parseInt(request.header["x-user-id"]) >= 320482) {
							return true
						}
						return false;
					})()
                `,
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "test0",
			args: args{
				ctx: context.Background(),
				request: &interact.Request{
					Method: "POST",
					Header: map[string]string{
						"x-user-id": "320481",
					},
				},
				script: `
					(function(){
						if (parseInt(request.header["x-user-id"]) == 320482) {
							return true
						}
						return false;
					})()
                `,
			},
			want:    false,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MatchRequestByJavascript(tt.args.ctx, tt.args.request, tt.args.script)
			if (err != nil) != tt.wantErr {
				t.Errorf("MatchRequestByJavascript() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("MatchRequestByJavascript() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMockResponseByJavascript(t *testing.T) {
	type args struct {
		ctx      context.Context
		request  *interact.Request
		response *interact.Response
		script   string
	}
	tests := []struct {
		name    string
		args    args
		want    *interact.Response
		wantErr bool
	}{
		{
			name: "test0",
			args: args{
				ctx: context.Background(),
				request: &interact.Request{
					Method: "POST",
					Header: map[string]string{
						"x-user-id": "320482",
					},
				},
				response: &interact.Response{
					Body: interact.NewBytesMessage(nil),
				},
				script: `
					(function(){
						return {
							code: 200,
							header: {
								"x-service-token": "micro-" + request.header["x-user-id"],
								"x-trace-id": "j92e210u90",
							},
							body: {message: "OK", code: 200},
						}
					})()
                `,
			},
			want: &interact.Response{
				Code: 200,
				Header: map[string]string{
					"x-service-token": "micro-320482",
					"x-trace-id":      "j92e210u90",
				},
				Body:    interact.NewBytesMessage([]byte(`{"message":"OK","code":200}`)),
				Trailer: nil,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := MockResponseByJavascript(tt.args.ctx, tt.args.request, tt.args.response, tt.args.script)
			if (err != nil) != tt.wantErr {
				t.Errorf("MockResponseByJavascript() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(tt.args.response, tt.want) {
				t.Errorf("MockResponseByJavascript() got = %v, want %v", tt.args.response, tt.want)
			}
		})
	}
}
