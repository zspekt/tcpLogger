package main

import "testing"

func Test_slogFatal(t *testing.T) {
	type args struct {
		msg  string
		args []any
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			slogFatal(tt.args.msg, tt.args.args...)
		})
	}
}

func Test_getEnvOrDefault(t *testing.T) {
	type args struct {
		key string
		def string
	}
	tests := []struct {
		name string
		args args

		want string

		wantErr    bool
		wantErrStr string

		wantSetEnv bool
		envVal     string
	}{
		{
			name: "calling with empty key arg",
			args: args{
				key: "",
				def: "NOTEMPTY",
			},
			want:       "",
			wantErr:    true,
			wantErrStr: "caller passed an empty string as key arg",
			wantSetEnv: false,
			envVal:     "",
		},
		{
			name: "calling with empty def arg",
			args: args{
				key: "NOTEMPTY",
				def: "",
			},
			want:       "",
			wantErr:    true,
			wantErrStr: "caller passed an empty string as def arg",
			wantSetEnv: false,
			envVal:     "",
		},
		{
			name: "calling with empty strings for both args",
			args: args{
				key: "",
				def: "",
			},
			want:       "",
			wantErr:    true,
			wantErrStr: "caller passed empty strings for key and def arg",
			wantSetEnv: false,
			envVal:     "",
		},
		{
			name: "calling with unset key arg",
			args: args{
				key: "THISENVVARISNOTSET",
				def: "THISISWHATWEWANT",
			},
			want:       "THISISWHATWEWANT",
			wantErr:    false,
			wantErrStr: "",
			wantSetEnv: false,
			envVal:     "",
		},
		{
			name: "calling with a set but empty key arg",
			args: args{
				key: "THISENVVARISEMPTY",
				def: "THISISWHATWEWANT",
			},
			want:       "THISISWHATWEWANT",
			wantErr:    false,
			wantErrStr: "",
			wantSetEnv: true,
			envVal:     "",
		},
		{
			name: "SAPBEE!!",
			args: args{
				key: "SAPBE",
				def: "NOTEMPTY",
			},
			want:       "SAPBE",
			wantErr:    false,
			wantErrStr: "",
			wantSetEnv: true,
			envVal:     "SAPBE",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantSetEnv {
				t.Setenv(tt.args.key, tt.envVal)
			}

			got, err := getEnvOrDefault(tt.args.key, tt.args.def)

			if tt.wantErr { // if this is a fail case
				switch {
				case tt.wantErrStr != err.Error(): // we did get an error, but not the one we wanted
					t.Errorf(
						"getEnvOrDefault() got error <%v>, want error <%v>",
						err,
						tt.wantErrStr,
					)

				case err == nil: // we wanted an error, but didn't get one
					t.Errorf(
						"getEnvOrDefault() returned no error, want error <%v>",
						tt.wantErrStr,
					)
				}
			}

			if got != tt.want {
				t.Errorf("getEnvOrDefault() got <%v>, wanted <%v>", got, tt.want)
			}
		})
	}
}

func Test_setup(t *testing.T) {
	tests := []struct {
		name string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup()
		})
	}
}
