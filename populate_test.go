package env

import (
	"os"
	"testing"
)

var (
	testEnvRestrictTo = []string{
		"AWS_REGION",
		"PORT",
		"SYSLOG_ADDRESS",
	}
)

type testEnv struct {
	AwsRegion     string `env:"AWS_REGION,optional"`
	Port          int    `env:"PORT,optional"`
	SyslogAddress string `env:"SYSLOG_ADDRESS,required"`
}

// TODO(pedge): if tests are run in parallel, this is affecting global state
func TestBasic(t *testing.T) {
	testState := newTestState(testEnvRestrictTo)
	defer testState.reset()

	_ = os.Unsetenv("AWS_REGION")
	_ = os.Setenv("PORT", "1234")
	_ = os.Setenv("SYSLOG_ADDRESS", "foo")

	testEnv := &testEnv{}
	if err := Populate(
		testEnv,
		PopulateOptions{
			RestrictTo: []string{
				"AWS_REGION",
				"PORT",
				"SYSLOG_ADDRESS",
			},
		},
	); err != nil {
		t.Error(err)
	}
	if testEnv.AwsRegion != "" {
		t.Errorf("expected no aws region, got %s", testEnv.AwsRegion)
	}
	if testEnv.Port != 1234 {
		t.Errorf("expected port of 1234, got %d", testEnv.Port)
	}
	if testEnv.SyslogAddress != "foo" {
		t.Errorf("expected syslog address of foo, got %s", testEnv.SyslogAddress)
	}
}

type testState struct {
	originalEnv map[string]string
}

func newTestState(restrictTo []string) *testState {
	originalEnv := make(map[string]string)
	for _, elem := range restrictTo {
		originalEnv[elem] = os.Getenv(elem)
	}
	return &testState{
		originalEnv,
	}
}

func (t *testState) reset() {
	for key, value := range t.originalEnv {
		_ = os.Setenv(key, value)
	}
}
