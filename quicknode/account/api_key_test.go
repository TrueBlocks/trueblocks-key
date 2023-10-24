package qnaccount

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigateway/types"
)

func Test_findPlanApiKey(t *testing.T) {
	type args struct {
		qnPlanSlug string
		planToKey  map[string]*types.ApiKey
	}
	tests := []struct {
		name       string
		args       args
		wantApiKey *ApiKey
		wantErr    bool
	}{
		{
			name: "error when plan not found",
			args: args{
				qnPlanSlug: "not-to-be-found",
				planToKey: map[string]*types.ApiKey{
					"other-plan": {
						Value: aws.String("0th3r-pl4n"),
						Name:  aws.String("other-plan"),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "finds plan",
			args: args{
				qnPlanSlug: "to-be-found",
				planToKey: map[string]*types.ApiKey{
					"to-be-found": {
						Value: aws.String("s3cr3t"),
						Name:  aws.String("to-be-found"),
					},
				},
			},
			wantApiKey: &ApiKey{
				Name:  "to-be-found",
				Value: "s3cr3t",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotApiKey, err := findPlanApiKey(tt.args.qnPlanSlug, tt.args.planToKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("findPlanApiKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotApiKey, tt.wantApiKey) {
				t.Errorf("findPlanApiKey() = %v, want %v", gotApiKey, tt.wantApiKey)
			}
		})
	}
}
