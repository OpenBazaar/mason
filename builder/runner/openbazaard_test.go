package runner

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/OpenBazaar/mason/util"
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
			{ // supports short testnet flag
				input: []string{"-t"},
				test: func(t *testing.T, r *OpenBazaarRunner) {
					if !r.enableTestnet {
						t.Errorf("expected -t to enable testnet, but did not")
					}
				},
			},
			{ // supports short testnet flag
				input: []string{"--testnet"},
				test: func(t *testing.T, r *OpenBazaarRunner) {
					if !r.enableTestnet {
						t.Errorf("expected --testnet to enable testnet, but did not")
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

func TestSetCustomDataPathChecksInitialized(t *testing.T) {
	var (
		subject   = &OpenBazaarRunner{}
		statePath = util.GenerateTempBuildPath("test_nodestatecopiesorig")
	)
	if err := os.MkdirAll(filepath.Join(statePath, "extraDir"), 0755); err != nil {
		t.Fatal(err)
	}
	if subject.state != stateReady {
		t.Errorf("expected runner to be state (%d), but was (%d)", stateReady, subject.state)
	}
	if err := subject.SetCustomDataPath(statePath); err != nil {
		t.Fatal(err)
	}
	if subject.state != stateInitialized {
		t.Errorf("expected runner to be state (%d), but was (%d)", stateInitialized, subject.state)
	}

}

func TestBeginNodeStateTransactionRequiresReadyState(t *testing.T) {
	var examples = []struct {
		state       runnerState
		expectedErr error
	}{
		{ // raw case, no error
			state:       stateReady,
			expectedErr: nil,
		},
		{ // init'd case, no error
			state:       stateInitialized,
			expectedErr: nil,
		},
		{ // cannot start state transaction while running
			state:       stateRunning,
			expectedErr: ErrCannotStartStateTransactionWhileRunning,
		},
	}

	for _, e := range examples {
		var (
			s         = &OpenBazaarRunner{}
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

		// set state
		s.state = e.state

		if err := s.BeginNodeStateTransaction(); err != e.expectedErr {
			t.Errorf("expected state (%d) to be error (%s), but was (%v)", e.state, e.expectedErr, err)
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

func TestSetConfigValue(t *testing.T) {
	type nestConfig struct {
		TwoKey   string   `json:"twoKey"`
		ThreeKey []string `json:"threeKey"`
	}
	type config struct {
		OneKey     string     `json:"oneKey"`
		NestStruct nestConfig `json:"nestStruct"`
		Untouched  string     `json:"donttouch"`
	}

	var (
		subject    = &OpenBazaarRunner{}
		statePath  = util.GenerateTempBuildPath("test_nodestatecopiesorig")
		testConfig = config{
			OneKey: "1",
			NestStruct: nestConfig{
				TwoKey: "0",
				ThreeKey: []string{
					"one",
					"two",
				},
			},
			Untouched: "nonzero",
		}
		configPath  = filepath.Join(statePath, "config")
		expectedArr = []string{"three", "four"}
	)
	if err := os.MkdirAll(filepath.Join(statePath, "extraDir"), 0755); err != nil {
		t.Fatal(err)
	}
	configContent, err := json.Marshal(testConfig)
	if err != nil {
		t.Fatal(err)
	}
	if err := ioutil.WriteFile(configPath, configContent, 0644); err != nil {
		t.Fatal(err)
	}
	if err := subject.SetCustomDataPath(statePath); err != nil {
		t.Fatal(err)
	}
	if err := subject.SetConfigValue("oneKey", "2"); err != nil {
		t.Fatal(err)
	}
	if err := subject.SetConfigValue("nestStruct.twoKey", "3"); err != nil {
		t.Fatal(err)
	}
	if err := subject.SetConfigValue("nestStruct.threeKey", expectedArr); err != nil {
		t.Fatal(err)
	}

	b, err := ioutil.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(b, &testConfig); err != nil {
		t.Fatal(err)
	}

	if testConfig.OneKey != "2" {
		t.Errorf("expected outer key to be (2), but was (%s)", testConfig.OneKey)
	}
	if testConfig.NestStruct.TwoKey != "3" {
		t.Errorf("expected nested key to be (3), but was (%s)", testConfig.NestStruct.TwoKey)
	}
	if !reflect.DeepEqual(expectedArr, testConfig.NestStruct.ThreeKey) {
		t.Errorf("expected string array to be (%v), but was (%v)", expectedArr, testConfig.NestStruct.ThreeKey)
	}
	if testConfig.Untouched != "nonzero" {
		t.Errorf("expected nested key to be (nonzero), but was (%s)", testConfig.Untouched)
	}
}

func TestSetConfigValueFailsIfNotInitialized(t *testing.T) {
	var subject = &OpenBazaarRunner{state: stateReady}
	if err := subject.SetConfigValue("", ""); err == nil {
		t.Fatal("expected uninitialized runner to error, but did not")
	} else if err != ErrInitNodeBeforeConfigValueSet {
		t.Fatalf("expected error to be (%v), but was (%v)", ErrInitNodeBeforeConfigValueSet, err)
	}
}
