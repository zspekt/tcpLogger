package setup

import (
	"testing"
)

func Test_getEnvOrDefault(t *testing.T) {
	type args struct {
		key string
		def string
	}
	tests := []struct {
		name string
		args args

		want string

		wantErr       bool
		wantErrTarget error

		wantSetEnv bool
		envVal     string
	}{
		{
			name: "calling with empty key arg",
			args: args{key: "", def: "NOTEMPTY"},
			want: "",

			wantErr: true,
			wantErrTarget: &ArgError{
				Err:   "empty string passed as argument",
				Param: []string{"key"},
			},

			wantSetEnv: false,
			envVal:     "",
		},
		{
			name: "calling with empty def arg",
			args: args{key: "NOTEMPTY", def: ""},
			want: "",

			wantErr: true,
			wantErrTarget: &ArgError{
				Err:   "empty string passed as argument",
				Param: []string{"def"},
			},

			wantSetEnv: false,
			envVal:     "",
		},
		{
			name: "calling with empty strings for both args",
			args: args{key: "", def: ""},
			want: "",

			wantErr: true,
			wantErrTarget: &ArgError{
				Err:   "empty string passed as argument",
				Param: []string{"key", "def"},
			},

			wantSetEnv: false,
			envVal:     "",
		},
		{
			name: "calling with unset key arg",
			args: args{key: "THISENVVARISNOTSET", def: "THISISWHATWEWANT"},
			want: "THISISWHATWEWANT",

			wantErr:       false,
			wantErrTarget: nil,

			wantSetEnv: false,
			envVal:     "",
		},
		{
			name: "calling with a set but empty key arg",
			args: args{key: "THISENVVARISEMPTY", def: "THISISWHATWEWANT"},
			want: "THISISWHATWEWANT",

			wantErr:       false,
			wantErrTarget: nil,

			wantSetEnv: true,
			envVal:     "",
		},
		{
			name: "SAPBEE!!",
			args: args{key: "SAPBE", def: "NOTEMPTY"},
			want: "SAPBE",

			wantErr:       false,
			wantErrTarget: nil,

			wantSetEnv: true,
			envVal:     "SAPBE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantSetEnv { // do we have to set an env var for this case?
				t.Setenv(tt.args.key, tt.envVal)
			}

			got, err := getEnvOrDefaultString(tt.args.key, tt.args.def)

			if tt.wantErr { // if this is a fail case

				switch {
				case tt.wantErrTarget.Error() != err.Error(): // we did get an error, but not the one we wanted
					t.Errorf(
						"getEnvOrDefault() got error <%v>, want error <%v>",
						err,
						tt.wantErrTarget.Error(),
					)

				case err == nil: // we wanted an error, but didn't get one
					t.Errorf(
						"getEnvOrDefault() returned no error, want error <%v>",
						tt.wantErrTarget.Error(),
					)
				}
			}

			if got != tt.want {
				t.Errorf("getEnvOrDefault() got <%v>, wanted <%v>", got, tt.want)
			}
		})
	}
}

func Test_setupConfig(t *testing.T) {
	tests := []struct {
		name string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Config()
		})
	}
}

func Test_setupLogger(t *testing.T) {
	tests := []struct {
		name string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger()
		})
	}
}
