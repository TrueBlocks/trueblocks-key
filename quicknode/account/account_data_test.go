package qnaccount

import (
	"encoding/json"
	"reflect"
	"testing"
)

type testRequestContext struct {
	Json    []byte
	Headers map[string]string
}

func (t *testRequestContext) GetHeader(key string) string {
	return t.Headers[key]
}

func (t *testRequestContext) BindJSON(obj any) error {
	return json.Unmarshal(t.Json, obj)
}

func TestFromRequestContext(t *testing.T) {
	type args struct {
		c requestContext
	}
	tests := []struct {
		name            string
		args            args
		wantAccountData *AccountData
		wantErr         bool
	}{
		{
			name: "invalid JSON",
			args: args{
				c: func() *testRequestContext {
					return &testRequestContext{
						Json: []byte("invalid json"),
					}
				}(),
			},
			wantAccountData: &AccountData{},
			wantErr:         true,
		},
		{
			name: "valid JSON",
			args: args{
				c: func() *testRequestContext {
					encoded, err := json.Marshal(AccountData{
						QuicknodeId: "test",
						Chain:       "ethereum",
						Network:     "mainnet",
					})
					if err != nil {
						t.Fatal(err)
					}
					return &testRequestContext{
						Json: encoded,
					}
				}(),
			},
			wantAccountData: &AccountData{
				QuicknodeId: "test",
				Chain:       "ethereum",
				Network:     "mainnet",
			},
		},
		{
			name: "test header",
			args: args{
				c: func() *testRequestContext {
					encoded, err := json.Marshal(AccountData{
						QuicknodeId: "test",
						Chain:       "ethereum",
						Network:     "mainnet",
					})
					if err != nil {
						t.Fatal(err)
					}
					return &testRequestContext{
						Json: encoded,
						Headers: map[string]string{
							"X-QN-TESTING": "true",
						},
					}
				}(),
			},
			wantAccountData: &AccountData{
				QuicknodeId: "test",
				Chain:       "ethereum",
				Network:     "mainnet",
				Test:        true,
			},
			wantErr: false,
		},
		{
			name: "invalid chain",
			args: args{
				c: func() *testRequestContext {
					encoded, err := json.Marshal(AccountData{
						QuicknodeId: "test",
						Chain:       "notethereum",
						Network:     "mainnet",
					})
					if err != nil {
						t.Fatal(err)
					}
					return &testRequestContext{
						Json: encoded,
					}
				}(),
			},
			wantAccountData: &AccountData{
				QuicknodeId: "test",
				Chain:       "notethereum",
				Network:     "mainnet",
			},
			wantErr: true,
		},
		{
			name: "invalid network",
			args: args{
				c: func() *testRequestContext {
					encoded, err := json.Marshal(AccountData{
						QuicknodeId: "test",
						Chain:       "ethereum",
						Network:     "nonet",
					})
					if err != nil {
						t.Fatal(err)
					}
					return &testRequestContext{
						Json: encoded,
					}
				}(),
			},
			wantAccountData: &AccountData{
				QuicknodeId: "test",
				Chain:       "ethereum",
				Network:     "nonet",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAccountData, err := NewAccountData(tt.args.c)
			if (err != nil) != tt.wantErr {
				t.Errorf("FromRequestContext() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotAccountData, tt.wantAccountData) {
				t.Errorf("FromRequestContext() = %v, want %v", gotAccountData, tt.wantAccountData)
			}
		})
	}
}
