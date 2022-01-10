package projparser

import (
	"io"
	"strings"
	"testing"
)

func Test_grabModuleNameFromGoMod(t *testing.T) {
	type args struct {
		r io.Reader
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "simple file",
			args: args{r: strings.NewReader("// test\n// dependencies analyser\nmodule test8\n\ngo 1.16\n\nrequire github.com/iv-menshenin/appctl v1.0.0\n")},
			want: "test8",
		},
		{
			name: "short file",
			args: args{r: strings.NewReader(`module github.com/somerepo/proj`)},
			want: "github.com/somerepo/proj",
		},
		{
			name: "empty file",
			args: args{r: strings.NewReader("")},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := grabModuleNameFromGoMod(tt.args.r); got != tt.want {
				t.Errorf("grabModuleNameFromGoMod() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getCurrentPackageName(t *testing.T) {
	t.Run("gomod", func(t *testing.T) {
		const ex = "somerepo.someserver.com/test/test"
		if got := getCurrentPackageName("./test/gomod-test/"); got != ex {
			t.Error("wrong module name\nexpected:", ex, "\ngot:", got)
		}
	})
	t.Run("not exists", func(t *testing.T) {
		if got := getCurrentPackageName("../unexists"); got != "" {
			t.Error("wrong module name\nexpected nothing", "\ngot:", got)
		}
	})
}
