package restmime

import (
	"bytes"
	"testing"

	"github.com/wencan/fastrest/restutils"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"
	"google.golang.org/protobuf/proto"
)

func TestMarshal(t *testing.T) {
	type args struct {
		v           interface{}
		contentType string
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "json",
			args: args{
				v: &struct {
					Greeting string `json:"greeting"`
					Name     string `json:"name"`
				}{
					Greeting: "hi",
					Name:     "Tom",
				},
				contentType: "application/json",
			},
			want:    []byte(`{"greeting":"hi","name":"Tom"}`),
			wantErr: false,
		},
		{
			name: "form",
			args: args{
				v: &struct {
					Greeting string `schema:"greeting"`
					Name     string `schema:"name"`
				}{
					Greeting: "hi",
					Name:     "Tom",
				},
				contentType: "application/x-www-form-urlencoded",
			},
			want:    []byte(`greeting=hi&name=Tom`),
			wantErr: false,
		},
		{
			name: "protobuf",
			args: args{
				v: &pb.HelloRequest{
					Name: "Hi",
				},
				contentType: "application/x-protobuf",
			},
			want: func() []byte {
				data, _ := proto.Marshal(&pb.HelloRequest{
					Name: "Hi",
				})
				return data
			}(),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writer := &bytes.Buffer{}
			if err := Marshal(tt.args.v, tt.args.contentType, writer); (err != nil) != tt.wantErr {
				t.Errorf("Marshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotBody := writer.String(); restutils.CompareHumanizeString(gotBody, string(tt.want)) != 0 {
				t.Errorf("Marshal() = %v, want %v", gotBody, string(tt.want))
			}
		})
	}
}
