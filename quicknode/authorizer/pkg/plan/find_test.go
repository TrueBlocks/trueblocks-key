package plan

import (
	"testing"
)

func Test_findPlanApiKey(t *testing.T) {
	type args struct {
		qnPlanSlug string
		planToKey  map[string]string
	}
	tests := []struct {
		name         string
		args         args
		wantKeyValue string
		wantErr      bool
	}{
		{
			name: "error when plan not found",
			args: args{
				qnPlanSlug: "not-to-be-found",
				planToKey: map[string]string{
					"other-plan": "0th3r-pl4n",
				},
			},
			wantErr: true,
		},
		{
			name: "finds plan",
			args: args{
				qnPlanSlug: "to-be-found",
				planToKey: map[string]string{
					"to-be-found": "s3cr3t",
				},
			},
			wantKeyValue: "s3cr3t",
			wantErr:      false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotKeyValue, err := findPlanApiKey(tt.args.qnPlanSlug, tt.args.planToKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("findPlanApiKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotKeyValue != tt.wantKeyValue {
				t.Errorf("findPlanApiKey() = %v, want %v", gotKeyValue, tt.wantKeyValue)
			}
		})
	}
}
