package runner

import "testing"

func TestWithArgs(t *testing.T) {
	var (
		examples = []struct {
			input []string
			test  func(*testing.T, *OpenBazaarRunner)
		}{
			{ // only config data flag
				input: []string{"-d", "foo"},
				test: func(t *testing.T, r *OpenBazaarRunner) {
					if r.configPath != "foo" {
						t.Errorf("expected 'foo' to be set on configPath, but was %s", r.configPath)
					}
				},
			},
			{ // data flag with additional args
				input: []string{"other", "flag", "-d", "bar", "flag2"},
				test: func(t *testing.T, r *OpenBazaarRunner) {
					if r.configPath != "bar" {
						t.Errorf("expected 'bar' to be set on configPath, but was %s", r.configPath)
					}
					if r.additionalArgsString() != "other flag flag2" {
						t.Errorf("expected -d flag values to be extracted, but were (%s)", r.additionalArgsString())
					}
				},
			},
			{ // quietly ignores missing args after -d
				input: []string{"flags", "flag2", "-d"},
				test: func(t *testing.T, r *OpenBazaarRunner) {
					if r.configPath != "" {
						t.Errorf("expected configPath to be empty, but was (%s)", r.configPath)
					}
				},
			},
		}
	)

	for _, e := range examples {
		var r = &OpenBazaarRunner{}
		r.WithArgs(e.input)
		e.test(t, r)
	}
}
