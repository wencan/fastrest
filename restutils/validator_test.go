package restutils

import (
	"context"
	"testing"
)

func TestValidateStruct(t *testing.T) {
	type Request struct {
		// Page 页码。从第一页开始。
		Page int `validate:"required,gte=1"`

		// Limit 每页数量上限。
		Limit int `validate:"required"`
	}

	type ComplexRequest struct {
		Request Request `validate:"required"`
		Name    string  `validate:"required,max=10"`
	}

	type ComplexSlice struct {
		Requests []*Request `validate:"dive,required"`
		Name     string     `validate:"required"`
	}

	tests := []struct {
		name    string
		args    interface{}
		wantErr bool
	}{
		{
			name: "struct_ok",
			args: Request{
				Page:  1,
				Limit: 10,
			},
		},
		{
			name: "struct_fail",
			args: Request{
				Page:  0,
				Limit: 10,
			},
			wantErr: true,
		},
		{
			name: "struct_ptr_ok",
			args: &Request{
				Page:  1,
				Limit: 10,
			},
		},
		{
			name: "struct_ptr_fail",
			args: &Request{
				Page:  0,
				Limit: 10,
			},
			wantErr: true,
		},
		{
			name: "struct_slice_ok",
			args: []Request{
				{
					Page:  1,
					Limit: 10,
				}},
		},
		{
			name: "struct_slice_fail",
			args: []Request{
				{
					Page:  0,
					Limit: 10,
				},
			},
			wantErr: true,
		},
		{
			name: "struct_ptr_slice_ok",
			args: []*Request{
				{
					Page:  1,
					Limit: 10,
				}},
		},
		{
			name: "struct_ptr_slice_fail",
			args: []*Request{
				{
					Page:  0,
					Limit: 10,
				},
			},
			wantErr: true,
		},
		{
			name: "complex_ok",
			args: ComplexRequest{
				Request: Request{
					Page:  1,
					Limit: 10,
				},
				Name: "测试",
			},
			wantErr: false,
		},
		{
			name: "complex_fail",
			args: ComplexRequest{
				Request: Request{
					Page:  0,
					Limit: 10,
				},
				Name: "测试",
			},
			wantErr: true,
		},
		{
			name: "complex_fail_2",
			args: ComplexRequest{
				Request: Request{
					Page:  1,
					Limit: 10,
				},
				Name: "",
			},
			wantErr: true,
		},
		{
			name: "complex_slice_ok",
			args: ComplexSlice{
				Requests: []*Request{
					{
						Page:  1,
						Limit: 10,
					},
					{
						Page:  2,
						Limit: 10,
					},
				},
				Name: "测试",
			},
			wantErr: false,
		},
		{
			name: "complex_slice_fail",
			args: ComplexSlice{
				Requests: []*Request{
					{
						Page:  1,
						Limit: 10,
					},
					{
						Page:  0,
						Limit: 10,
					},
				},
				Name: "测试",
			},
			wantErr: true,
		},
		{
			name: "complex_slice_fail_2",
			args: ComplexSlice{
				Requests: []*Request{
					{
						Page:  1,
						Limit: 10,
					},
					{
						Page:  2,
						Limit: 10,
					},
				},
				Name: "",
			},
			wantErr: true,
		},
		{
			name: "complex_maxlimit_ok",
			args: ComplexRequest{
				Request: Request{
					Page:  1,
					Limit: 10,
				},
				Name: "测试测试测试测试测试",
			},
			wantErr: false,
		},
		{
			name: "complex_maxlimit_fail",
			args: ComplexRequest{
				Request: Request{
					Page:  1,
					Limit: 10,
				},
				Name: "测试测试测试测试测试0",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateStruct(context.TODO(), tt.args); (err != nil) != tt.wantErr {
				t.Errorf("ValidateStruct() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
