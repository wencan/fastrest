package restutils

import (
	"reflect"
	"testing"
)

func TestSubstringsWithTags(t *testing.T) {
	type args struct {
		str      string
		startTag string
		endTag   string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "empty",
			args: args{
				str:      "",
				startTag: "(",
				endTag:   ")",
			},
			want: nil,
		},
		{
			name: "nothing",
			args: args{
				str:      "测试这个一下",
				startTag: "(",
				endTag:   ")",
			},
		},
		{
			name: "one",
			args: args{
				str:      "测试(这个)一下",
				startTag: "(",
				endTag:   ")",
			},
			want: []string{"这个"},
		},
		{
			name: "two",
			args: args{
				str:      "测试(这个)一下，还有(this)",
				startTag: "(",
				endTag:   ")",
			},
			want: []string{"这个", "this"},
		},
		{
			name: "three",
			args: args{
				str:      "(now)测试(这个)一下，还有(this)",
				startTag: "(",
				endTag:   ")",
			},
			want: []string{"now", "这个", "this"},
		},
		{
			name: "strTag",
			args: args{
				str:      "测试{[(这个)]}一下",
				startTag: "{[(",
				endTag:   ")]}",
			},
			want: []string{"这个"},
		},
		{
			name: "repeat",
			args: args{
				str:      "测试(this)一下，还有(this)",
				startTag: "(",
				endTag:   ")",
			},
			want: []string{"this", "this"},
		},
		{
			name: "one_lost-endTag",
			args: args{
				str:      "测试(这个一下",
				startTag: "(",
				endTag:   ")",
			},
		},
		{
			name: "one_lost-startTag",
			args: args{
				str:      "测试这个)一下",
				startTag: "(",
				endTag:   ")",
			},
		},
		{
			name: "two_lost-endTag",
			args: args{
				str:      "测试(这个一下，还有(this)",
				startTag: "(",
				endTag:   ")",
			},
			want: []string{"this"},
		},
		{
			name: "two_lost-endTag-2",
			args: args{
				str:      "测试(这个)一下，还有(this",
				startTag: "(",
				endTag:   ")",
			},
			want: []string{"这个"},
		},
		{
			name: "two_lost_startTag",
			args: args{
				str:      "测试这个)一下，还有(this)",
				startTag: "(",
				endTag:   ")",
			},
			want: []string{"this"},
		},
		{
			name: "two_lost_startTag-2",
			args: args{
				str:      "测试(这个)一下，还有this)",
				startTag: "(",
				endTag:   ")",
			},
			want: []string{"这个"},
		},
		{
			name: "no_tag",
			args: args{
				str:      "测试这个一下",
				startTag: "(",
				endTag:   ")",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SubstringsWithTags(tt.args.str, tt.args.startTag, tt.args.endTag); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SubstringsWithTags() = %v, want %v", got, tt.want)
			}
		})
	}
}
