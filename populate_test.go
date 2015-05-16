package env

import (
	"os"
	"reflect"
	"strings"
	"testing"
)

var (
	testEnvRestrictTo = []string{
		"REQUIRED_STRING",
		"OPTIONAL_STRING",
		"OPTIONAL_INT",
		"OPTIONAL_BOOL",
		"STRUCT_OPTIONAL_INT",
		"OPTIONAL_STRUCT",
	}
	testEnvRestrictToWithoutOptionalBool = []string{
		"OPTIONAL_STRING",
		"OPTIONAL_INT",
		"REQUIRED_STRING",
		"STRUCT_OPTIONAL_INT",
		"OPTIONAL_STRUCT",
	}
)

type testEnv struct {
	RequiredString string `env:"REQUIRED_STRING,required"`
	OptionalString string `env:"OPTIONAL_STRING,optional"`
	OptionalInt    int    `env:"OPTIONAL_INT,optional"`
	OptionalBool   bool   `env:"OPTIONAL_BOOL,optional"`
	OptionalStruct struct {
		StructOptionalInt int `env:"STRUCT_OPTIONAL_INT,optional"`
	} `env:"OPTIONAL_STRUCT,optional,FOO"`
}

// TODO(pedge): if tests are run in parallel, this is affecting global state

func TestBasic(t *testing.T) {
	testState := newTestState(testEnvRestrictTo)
	defer testState.reset()
	testSetenv(map[string]string{
		"REQUIRED_STRING": "foo",
		"OPTIONAL_STRING": "",
		"OPTIONAL_INT":    "1234",
	})
	testEnv := populateTestEnv(t)
	checkEqual(t, "", testEnv.OptionalString)
	checkEqual(t, 1234, testEnv.OptionalInt)
	checkEqual(t, "foo", testEnv.RequiredString)
}

func TestMissing(t *testing.T) {
	testState := newTestState(testEnvRestrictTo)
	defer testState.reset()
	testSetenv(map[string]string{
		"REQUIRED_STRING": "",
	})
	populateTestEnvExpectError(t, envKeyNotSetWhenRequiredErr)
}

func TestOutsideOfRestrictToRange(t *testing.T) {
	testState := newTestState(testEnvRestrictToWithoutOptionalBool)
	defer testState.reset()
	testSetenv(map[string]string{
		"REQUIRED_STRING": "foo",
		"OPTIONAL_BOOL":   "1",
	})
	populateTestEnvExpectErrorLong(t, invalidTagRestrictToErr, testEnvRestrictToWithoutOptionalBool, nil)
}

func TestCannotParse(t *testing.T) {
	testState := newTestState(testEnvRestrictTo)
	defer testState.reset()
	testSetenv(map[string]string{
		"REQUIRED_STRING": "foo",
		"OPTIONAL_INT":    "abc",
	})
	populateTestEnvExpectError(t, cannotParseErr)
}

func TestParsingBool(t *testing.T) {
	testState := newTestState(testEnvRestrictTo)
	defer testState.reset()
	testSetenv(map[string]string{
		"REQUIRED_STRING": "foo",
		"OPTIONAL_BOOL":   "1",
	})
	testEnv := populateTestEnv(t)
	checkEqual(t, true, testEnv.OptionalBool)
	testSetenv(map[string]string{
		"REQUIRED_STRING": "foo",
		"OPTIONAL_BOOL":   "",
	})
	testEnv = populateTestEnv(t)
	checkEqual(t, false, testEnv.OptionalBool)
	testSetenv(map[string]string{
		"REQUIRED_STRING": "foo",
		"OPTIONAL_BOOL":   "false",
	})
	testEnv = populateTestEnv(t)
	checkEqual(t, false, testEnv.OptionalBool)
}

func TestOptionalStruct(t *testing.T) {
	testState := newTestState(testEnvRestrictTo)
	defer testState.reset()
	testSetenv(map[string]string{
		"REQUIRED_STRING":     "foo",
		"OPTIONAL_STRUCT":     "FOO",
		"STRUCT_OPTIONAL_INT": "1234",
	})
	testEnv := populateTestEnv(t)
	checkEqual(t, 1234, testEnv.OptionalStruct.StructOptionalInt)
	testSetenv(map[string]string{
		"REQUIRED_STRING":     "foo",
		"OPTIONAL_STRUCT":     "",
		"STRUCT_OPTIONAL_INT": "1234",
	})
	testEnv = populateTestEnv(t)
	checkEqual(t, 0, testEnv.OptionalStruct.StructOptionalInt)
	testSetenv(map[string]string{
		"REQUIRED_STRING":     "foo",
		"OPTIONAL_STRUCT":     "BAR",
		"STRUCT_OPTIONAL_INT": "1234",
	})
	testEnv = populateTestEnv(t)
	checkEqual(t, 0, testEnv.OptionalStruct.StructOptionalInt)
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
	testSetenv(t.originalEnv)
}

func testSetenv(env map[string]string) {
	for key, value := range env {
		_ = os.Setenv(key, value)
	}
}

func populateTestEnv(t *testing.T) *testEnv {
	testEnv := &testEnv{}
	if err := Populate(
		testEnv,
		PopulateOptions{
			RestrictTo: testEnvRestrictTo,
		},
	); err != nil {
		t.Error(err)
	}
	return testEnv
}

func populateTestEnvExpectError(t *testing.T, expected string) {
	populateTestEnvExpectErrorLong(t, expected, testEnvRestrictTo, nil)
}

func populateTestEnvExpectErrorLong(t *testing.T, expected string, restrictTo []string, decoders []Decoder) {
	testEnv := &testEnv{}
	err := Populate(
		testEnv,
		PopulateOptions{
			RestrictTo: restrictTo,
			Decoders:   decoders,
		},
	)
	if err == nil {
		t.Error("expected error")
	} else if !strings.HasPrefix(err.Error(), expected) {
		t.Errorf("expected error type %s, got error %s", expected, err.Error())
	}
}

func checkEqual(t *testing.T, expected interface{}, actual interface{}) {
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("expected %v, got %v", expected, actual)
	}
}
