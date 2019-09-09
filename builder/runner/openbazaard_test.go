package runner

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/OpenBazaar/samulator/util"
)

func TestWithArgs(t *testing.T) {
	var (
		examples = []struct {
			input []string
			test  func(*testing.T, *OpenBazaarRunner)
		}{
			{ // only config data flag
				input: []string{"-d", "foo"},
				test: func(t *testing.T, r *OpenBazaarRunner) {
					if r.dataPath != "foo" {
						t.Errorf("expected 'foo' to be set on dataPath, but was %s", r.dataPath)
					}
				},
			},
			{ // data flag with additional args
				input: []string{"other", "flag", "-d", "bar", "flag2"},
				test: func(t *testing.T, r *OpenBazaarRunner) {
					if r.dataPath != "bar" {
						t.Errorf("expected 'bar' to be set on dataPath, but was %s", r.dataPath)
					}
					if r.additionalArgsString() != "other flag flag2" {
						t.Errorf("expected -d flag values to be extracted, but were (%s)", r.additionalArgsString())
					}
				},
			},
			{ // quietly ignores missing args after -d
				input: []string{"flags", "flag2", "-d"},
				test: func(t *testing.T, r *OpenBazaarRunner) {
					if r.dataPath != "" {
						t.Errorf("expected dataPath to be empty, but was (%s)", r.dataPath)
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

func TestBeginNodeStateTransactionRequiresReadyState(t *testing.T) {
	var examples = []struct {
		state       runnerState
		expectedErr error
	}{
		{ // normal case, no error
			state:       stateReady,
			expectedErr: nil,
		},
		{ // cannot start state transaction while running
			state:       stateRunning,
			expectedErr: ErrCannotStartStateTransactionWhileRunning,
		},
	}

	for _, e := range examples {
		var (
			s = &OpenBazaarRunner{
				state: e.state,
			}
			statePath = util.GenerateTempPath("stateTransactionRequiresReady")
		)
		if err := os.MkdirAll(filepath.Join(statePath, "extraDir"), 0755); err != nil {
			t.Fatal(err)
		}
		if err := ioutil.WriteFile(filepath.Join(statePath, "config"), []byte("configcontent"), 0644); err != nil {
			t.Fatal(err)
		}
		if err := s.SetCustomDataPath(statePath); err != nil {
			t.Fatal(err)
		}

		if err := s.BeginNodeStateTransaction(); err != e.expectedErr {
			t.Errorf("expected state (%d) to be error (%s), but was (%s)", e.state, e.expectedErr, err)
		}
	}
}

func TestBeginNodeStateCopiesStateTree(t *testing.T) {
	var (
		subject       = &OpenBazaarRunner{}
		statePath     = util.GenerateTempBuildPath("test_nodestatecopiesorig")
		configContent = []byte(`{"version":"1"}`)
	)
	if err := os.MkdirAll(filepath.Join(statePath, "extraDir"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := ioutil.WriteFile(filepath.Join(statePath, "config"), configContent, 0644); err != nil {
		t.Fatal(err)
	}
	if err := subject.SetCustomDataPath(statePath); err != nil {
		t.Fatal(err)
	}

	if err := subject.BeginNodeStateTransaction(); err != nil {
		t.Fatal(err)
	}
	if subject.txDataPath == "" {
		t.Fatal("expected transaction data path to be present, but was empty")
	}
	if subject.txDataPath == subject.dataPath {
		t.Fatal("expected transaction data path to be different from source data path, but was not")
	}
	t.Logf("original path: %s", subject.dataPath)
	t.Logf("temp path: %s", subject.txDataPath)

	if actualContent, err := ioutil.ReadFile(filepath.Join(subject.txDataPath, "config")); !bytes.Equal(configContent, actualContent) {
		t.Error("expected copied node state to be equivalent, but  was not")
		t.Logf("\texpected: (%s)\n\tactual: (%s)", configContent, actualContent)
		t.Logf("\terror returned: %s", err.Error())
	}

	if s, err := os.Stat(filepath.Join(subject.txDataPath, "config")); s.Mode().Perm() != 0644 {
		t.Errorf("expected copied node state config file to have the file mode (%o), but was (%o)", 0644, s.Mode())
		t.Logf("\terror returned: %s", err.Error())
	}

	if s, err := os.Stat(filepath.Join(subject.txDataPath, "extraDir")); err != nil {
		t.Logf("\terror returned: %s", err.Error())
	} else {
		if s.Mode().Perm() != 0755 {
			t.Errorf("expected copied node state nested folder to have the file permissions (%o), but was (%o)", 0755, s.Mode())
		}
	}
}
