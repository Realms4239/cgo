package campagne

import (
	"encoding/json"
	"testing"
)

// Sans delay_model gelé, une ligne dist est indiscernable d'une ligne scalar
// dans le manifeste : le config_sha256 ne sépare plus les deux bras d'un A/B.
func TestBuildMatrixConfigFreezesDelayModel(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   string
		want string
	}{
		{"dist explicite", "dist", "dist"},
		{"scalar par défaut", "", "scalar"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg, err := buildMatrixConfig([]string{"P2"}, []string{"cake"}, []string{"cubic"}, 1, Deps{DelayModel: tc.in})
			if err != nil {
				t.Fatal(err)
			}
			var got struct {
				DelayModel string `json:"delay_model"`
			}
			if err := json.Unmarshal(cfg, &got); err != nil {
				t.Fatal(err)
			}
			if got.DelayModel != tc.want {
				t.Fatalf("delay_model gelé = %q, voulu %q (config: %s)", got.DelayModel, tc.want, cfg)
			}
		})
	}
}
