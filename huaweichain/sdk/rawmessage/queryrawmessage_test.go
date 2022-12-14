package rawmessage

import "testing"

func TestSubject_conv(t *testing.T) {
	tests := []struct {
		name    string
		s       Subject
		want    string
		wantErr bool
	}{
		{"policy", SubjectPolicy, "cfgPolicy", false},
		{"org", SubjectOrg, "organization", false},
		{"consensus", SubjectConsensus, "consensus_configuration_change_pending", false},
		{"domain", SubjectDomain, "DOMAIN", false},
		{"zone", SubjectZone, "ZONE", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.s.conv()
			if (err != nil) != tt.wantErr {
				t.Errorf("conv() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("conv() got = %v, want %v", got, tt.want)
			}
		})
	}
}
