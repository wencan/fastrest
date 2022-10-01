package httpserver

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/wencan/fastrest/restutils"
)

func TestReadRequest(t *testing.T) {
	type args struct {
		dest interface{}
		r    *http.Request
	}
	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			name: "get_query",
			args: args{
				dest: &struct {
					Greeting string `schema:"greeting"`
					Name     string `schema:"name"`
				}{},
				r: httptest.NewRequest(http.MethodGet, "/test?greeting=hi&name=Tom", nil),
			},
			want: &struct {
				Greeting string `schema:"greeting"`
				Name     string `schema:"name"`
			}{
				Greeting: "hi",
				Name:     "Tom",
			},
			wantErr: false,
		},
		{
			name: "post_json",
			args: args{
				dest: &struct {
					Greeting string `json:"greeting"`
					Name     string `json:"name"`
				}{},
				r: func() *http.Request {
					r := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBufferString(`{"greeting":"hi","name":"Tom"}`))
					r.Header.Set("Content-Type", "application/json")
					return r
				}(),
			},
			want: &struct {
				Greeting string `json:"greeting"`
				Name     string `json:"name"`
			}{
				Greeting: "hi",
				Name:     "Tom",
			},
			wantErr: false,
		},
		{
			name: "patch_form",
			args: args{
				dest: &struct {
					Greeting string `schema:"greeting"`
					Name     string `schema:"name"`
				}{},
				r: func() *http.Request {
					r := httptest.NewRequest(http.MethodPatch, "/test", bytes.NewBufferString(`greeting=hi&name=Tom`))
					r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
					return r
				}(),
			},
			want: &struct {
				Greeting string `schema:"greeting"`
				Name     string `schema:"name"`
			}{
				Greeting: "hi",
				Name:     "Tom",
			},
			wantErr: false,
		},
		{
			name: "get_chinese_query",
			args: args{
				dest: &struct {
					Greeting string `schema:"greeting"`
					Name     string `schema:"name"`
				}{},
				r: httptest.NewRequest(http.MethodGet, "/test?greeting=hi&name=测试员", nil),
			},
			want: &struct {
				Greeting string `schema:"greeting"`
				Name     string `schema:"name"`
			}{
				Greeting: "hi",
				Name:     "测试员",
			},
			wantErr: false,
		},
		{
			name: "post_chinese_json",
			args: args{
				dest: &struct {
					Greeting string `json:"greeting"`
					Name     string `json:"name"`
				}{},
				r: func() *http.Request {
					r := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBufferString(`{"greeting":"hi","name":"测试员"}`))
					r.Header.Set("Content-Type", "application/json")
					return r
				}(),
			},
			want: &struct {
				Greeting string `json:"greeting"`
				Name     string `json:"name"`
			}{
				Greeting: "hi",
				Name:     "测试员",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ReadRequest(context.TODO(), tt.args.dest, tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ReadRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(tt.want, tt.args.dest) {
				t.Errorf("want: %s, got: %s", restutils.JsonString(tt.want), restutils.JsonString(tt.args.dest))
			}
		})
	}
}

func TestReadValidateRequest(t *testing.T) {
	type args struct {
		dest interface{}
		r    *http.Request
	}
	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			name: "ok",
			args: args{
				dest: &struct {
					Greeting string `schema:"greeting" validate:"required"`
				}{},
				r: httptest.NewRequest(http.MethodGet, "/test?greeting=hi", nil),
			},
			want: &struct {
				Greeting string `schema:"greeting" validate:"required"`
			}{
				Greeting: "hi",
			},
			wantErr: false,
		},
		{
			name: "no_validate_tag",
			args: args{
				dest: &struct {
					Greeting string `schema:"greeting"`
				}{},
				r: httptest.NewRequest(http.MethodGet, "/test?greeting=hi", nil),
			},
			want: &struct {
				Greeting string `schema:"greeting"`
			}{
				Greeting: "hi",
			},
			wantErr: false,
		},
		{
			name: "required",
			args: args{
				dest: &struct {
					Greeting string `schema:"greeting" validate:"required"`
				}{},
				r: httptest.NewRequest(http.MethodGet, "/test", nil),
			},
			want: &struct {
				Greeting string `schema:"greeting" validate:"required"`
			}{
				Greeting: "",
			},
			wantErr: true,
		},
		{
			name: "valid_email",
			args: args{
				dest: &struct {
					Email string `schema:"email" validate:"required,email"`
				}{},
				r: httptest.NewRequest(http.MethodGet, "/test?email=abc@email.com", nil),
			},
			want: &struct {
				Email string `schema:"email" validate:"required,email"`
			}{
				Email: "abc@email.com",
			},
			wantErr: false,
		},
		{
			name: "invalid_email",
			args: args{
				dest: &struct {
					Email string `schema:"email" validate:"required,email"`
				}{},
				r: httptest.NewRequest(http.MethodGet, "/test?email=abcemail.com", nil),
			},
			want: &struct {
				Email string `schema:"email" validate:"required,email"`
			}{
				Email: "abcemail.com",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ReadValidateRequest(context.TODO(), tt.args.dest, tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadValidateRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(tt.want, tt.args.dest) {
				t.Errorf("want: %s, got: %s", restutils.JsonString(tt.want), restutils.JsonString(tt.args.dest))
			}
		})
	}
}
