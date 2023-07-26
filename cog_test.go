package cog

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	fh "github.com/leonidasdeim/cog/filehandler"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	permissions            = 0664
	appName                = "test_app"
	activeConfig           = appName + ".%s"
	defaultConfig          = appName + ".default.%s"
	testDir                = "testDir/"
	testSetupErrorMsg      = "error while setting up test: %v"
	expectedResultErrorMsg = "expected config does not match the result"
)

type testConfig struct {
	Name      string `default:"app" env:"TEST_ENV_NAME"`
	Version   int    `validate:"required"`
	IsPrefork bool   `default:"true"`
}

var (
	testData            = testConfig{Name: "config_test", Version: 123, IsPrefork: true}
	testDataDefaultName = testConfig{Name: "app", Version: 123, IsPrefork: true}
	testDataEnvName     = testConfig{Name: "env_name", Version: 123, IsPrefork: true}
)

type testCase struct {
	Type                     fh.FileType
	TestString               string
	TestStringWithoutVersion string
	TestStringWithDefaults   string
}

var testCases = []testCase{
	{
		fh.JSON,
		"{\"name\":\"config_test\",\"version\":123}",
		"{\"name\":\"config_test\"}",
		"{\"version\":123}",
	},
	{
		fh.YAML,
		"name: config_test\nversion: 123\n",
		"name: config_test\n",
		"version: 123\n",
	},
	{
		fh.TOML,
		"name = \"config_test\"\nversion = 123\n",
		"name = \"config_test\"\n",
		"version = 123\n",
	},
}

func TestFileTypes(t *testing.T) {
	for _, tc := range testCases {
		t.Run(string(tc.Type), func(t *testing.T) {
			test(t, tc, testConfigLoading)
			test(t, tc, testEnvironmentVarIsOverwritten)
			test(t, tc, testConfigDefaults)
			test(t, tc, testLoadFromActiveConfig)
			test(t, tc, testActiveConfigCreated)
			test(t, tc, testActiveConfigContent)
			test(t, tc, testTimestampIsCreated)
			test(t, tc, testCustomConfigPath)
			test(t, tc, testDataWithoutRequiredField)
			test(t, tc, testDefaultValuesAreSet)
			test(t, tc, testEnvironmentValuesAreSet)
			test(t, tc, testDynamicTypeIsResolved)
			test(t, tc, testSubscribersAreRegistered)
			test(t, tc, testCallbacksAreRegistered)
			test(t, tc, testConfigUpdated)
			test(t, tc, testCallbacksAreNotifiedAndRemoved)
			test(t, tc, testSubscribersAreNotifiedAndRemoved)
			test(t, tc, testSubscriberReturnsError)
			test(t, tc, testUpdateConfigIsValidated)
		})
	}
}

func testConfigLoading(t *testing.T, tc testCase) {
	c, err := setup(t, fmt.Sprintf(defaultConfig, string(tc.Type)), "", tc.Type, tc.TestString)
	require.NoErrorf(t, err, testSetupErrorMsg)

	got := c.Config()
	assert.Equalf(t, testData, got, expectedResultErrorMsg)
}

func testEnvironmentVarIsOverwritten(t *testing.T, tc testCase) {
	os.Setenv("TEST_ENV_NAME", "env_name")

	c, err := setup(t, fmt.Sprintf(defaultConfig, string(tc.Type)), "", tc.Type, tc.TestString)
	require.NoErrorf(t, err, testSetupErrorMsg)

	got := c.Config()
	assert.Equalf(t, testData, got, expectedResultErrorMsg)
}

func testConfigDefaults(t *testing.T, tc testCase) {
	type ConfigNoRequiredFields struct {
		Name    string `default:"app"`
		Version int
		Store   struct {
			Host string `default:"localhost"`
			Port string `default:"123"`
		}
		IsPrefork bool `default:"true"`
	}

	c, err := Init[ConfigNoRequiredFields]()
	require.NoErrorf(t, err, testSetupErrorMsg)

	assert.FileExistsf(t, "app.json", "active config file is not created")
	assert.Equalf(t, "app", c.Config().Name, "default name is not set")
	assert.Equalf(t, true, c.Config().IsPrefork, "default isPrefork is not set")
	assert.Equalf(t, "localhost", c.Config().Store.Host, "default host is not set")
	assert.Equalf(t, "123", c.Config().Store.Port, "default port is not set")

	os.Remove("app.json")
}

func testLoadFromActiveConfig(t *testing.T, tc testCase) {
	c, err := setup(t, fmt.Sprintf(activeConfig, string(tc.Type)), "", tc.Type, tc.TestString)
	require.NoErrorf(t, err, testSetupErrorMsg)

	got := c.Config()
	assert.Equalf(t, testData, got, expectedResultErrorMsg)
}

func testActiveConfigCreated(t *testing.T, tc testCase) {
	_, err := setup(t, fmt.Sprintf(defaultConfig, string(tc.Type)), "", tc.Type, tc.TestString)
	require.NoErrorf(t, err, testSetupErrorMsg)

	assert.FileExistsf(t, fmt.Sprintf(activeConfig, string(tc.Type)), "active config file is not created")
}

func testActiveConfigContent(t *testing.T, tc testCase) {
	c, err := setup(t, fmt.Sprintf(defaultConfig, string(tc.Type)), "", tc.Type, tc.TestString)
	require.NoErrorf(t, err, testSetupErrorMsg)

	got := testConfig{}
	err = c.handler.Load(&got)
	assert.NoErrorf(t, err, "error while parsing active config file")

	assert.Equalf(t, testData, got, expectedResultErrorMsg)
}

func testTimestampIsCreated(t *testing.T, tc testCase) {
	c, err := setup(t, fmt.Sprintf(defaultConfig, string(tc.Type)), "", tc.Type, tc.TestString)
	require.NoErrorf(t, err, testSetupErrorMsg)

	assert.NotEmptyf(t, c.GetTimestamp(), "timestamp is not set")
}

func testCustomConfigPath(t *testing.T, tc testCase) {
	c, err := setup(t, fmt.Sprintf(defaultConfig, string(tc.Type)), testDir, tc.Type, tc.TestString)
	require.NoErrorf(t, err, testSetupErrorMsg)

	defaultConfigPath := filepath.Join(testDir, fmt.Sprintf(defaultConfig, string(tc.Type)))
	assert.FileExists(t, defaultConfigPath, "cannot find default config in expected location")

	activeConfigPth := filepath.Join(testDir, fmt.Sprintf(activeConfig, string(tc.Type)))
	assert.FileExists(t, activeConfigPth, "cannot find active config in expected location")

	got := c.Config()
	assert.Equalf(t, testData, got, expectedResultErrorMsg)
}

func testDataWithoutRequiredField(t *testing.T, tc testCase) {
	c, err := setup(t, fmt.Sprintf(defaultConfig, string(tc.Type)), "", tc.Type, tc.TestStringWithoutVersion)
	require.Errorf(t, err, "error is not returned")
	require.Nilf(t, c, "cog instance should be nil")

	assert.Containsf(t, err.Error(), "failed at validate config", "wrong error is returned")
}

func testDefaultValuesAreSet(t *testing.T, tc testCase) {
	c, err := setup(t, fmt.Sprintf(defaultConfig, string(tc.Type)), "", tc.Type, tc.TestStringWithDefaults)
	require.NoErrorf(t, err, testSetupErrorMsg)

	got := c.Config()
	assert.Equalf(t, testDataDefaultName, got, expectedResultErrorMsg)
}

func testEnvironmentValuesAreSet(t *testing.T, tc testCase) {
	os.Setenv("TEST_ENV_NAME", "env_name")

	c, err := setup(t, fmt.Sprintf(defaultConfig, string(tc.Type)), "", tc.Type, tc.TestStringWithDefaults)
	require.NoErrorf(t, err, testSetupErrorMsg)

	got := c.Config()
	assert.Equalf(t, testDataEnvName, got, expectedResultErrorMsg)
}

func testDynamicTypeIsResolved(t *testing.T, tc testCase) {
	c, err := setup(t, fmt.Sprintf(defaultConfig, string(tc.Type)), "", fh.DYNAMIC, tc.TestString)
	require.NoErrorf(t, err, testSetupErrorMsg)

	got := c.Config()
	assert.Equalf(t, testData, got, expectedResultErrorMsg)

	assert.FileExistsf(t, fmt.Sprintf(activeConfig, string(tc.Type)), "expected active config file not exists")
}

func testSubscribersAreRegistered(t *testing.T, tc testCase) {
	subs := [3]Subscriber[testConfig]{
		func(tc testConfig) error {
			return nil
		},
		func(tc testConfig) error {
			return nil
		},
		func(tc testConfig) error {
			return nil
		},
	}

	c, err := setup(t, fmt.Sprintf(defaultConfig, string(tc.Type)), "", tc.Type, tc.TestString)
	require.NoErrorf(t, err, testSetupErrorMsg)

	for _, cb := range subs {
		c.AddSubscriber(cb)
	}

	assert.Equalf(t, len(subs), len(c.subscribers), "expected number of subscribers")
}

func testCallbacksAreRegistered(t *testing.T, tc testCase) {
	cbs := [3]Callback[testConfig]{
		func(tc testConfig) {},
		func(tc testConfig) {},
		func(tc testConfig) {},
	}
	c, err := setup(t, fmt.Sprintf(defaultConfig, string(tc.Type)), "", tc.Type, tc.TestString)
	require.NoErrorf(t, err, testSetupErrorMsg)

	for _, f := range cbs {
		c.AddCallback(f)
	}

	assert.Equalf(t, len(cbs), len(c.callbacks), "expected number of callbacks")
}

var (
	newData                = testConfig{Name: "new_data", Version: 456}
	newDataWithoutRequired = testConfig{Name: "new_data"}
)

func testConfigUpdated(t *testing.T, tc testCase) {
	c, err := setup(t, fmt.Sprintf(defaultConfig, string(tc.Type)), "", tc.Type, tc.TestString)
	require.NoErrorf(t, err, testSetupErrorMsg)

	err = c.Update(newData)
	require.NoErrorf(t, err, "error while updating config: %v", err)

	got := c.Config()
	assert.Equalf(t, newData, got, expectedResultErrorMsg)
}

func testCallbacksAreNotifiedAndRemoved(t *testing.T, tc testCase) {
	var calls1, calls2 int
	cbs := [2]Callback[testConfig]{
		func(tc testConfig) { calls1++ },
		func(tc testConfig) { calls2++ },
	}

	c, err := setup(t, fmt.Sprintf(defaultConfig, string(tc.Type)), "", tc.Type, tc.TestString)
	require.NoErrorf(t, err, testSetupErrorMsg)

	c.AddCallback(cbs[0])
	callbackId := c.AddCallback(cbs[1])

	c.Update(newData)
	c.Update(newData)
	time.Sleep(100 * time.Millisecond)

	c.RemoveCallback(callbackId)

	c.Update(newData)
	c.Update(newData)
	time.Sleep(100 * time.Millisecond)

	assert.Equal(t, 4, calls1)
	assert.Equal(t, 2, calls2)
}

func testSubscribersAreNotifiedAndRemoved(t *testing.T, tc testCase) {
	var calls1, calls2 int
	subs := [2]Subscriber[testConfig]{
		func(tc testConfig) error {
			calls1++
			return nil
		},
		func(tc testConfig) error {
			calls2++
			return nil
		},
	}

	c, err := setup(t, fmt.Sprintf(defaultConfig, string(tc.Type)), "", tc.Type, tc.TestString)
	require.NoErrorf(t, err, testSetupErrorMsg)

	c.AddSubscriber(subs[0])
	subscriberId := c.AddSubscriber(subs[1])

	c.Update(newData)
	c.Update(newData)
	time.Sleep(100 * time.Millisecond)

	c.RemoveSubscriber(subscriberId)

	c.Update(newData)
	c.Update(newData)
	time.Sleep(100 * time.Millisecond)

	assert.Equal(t, 4, calls1)
	assert.Equal(t, 2, calls2)
}

func testSubscriberReturnsError(t *testing.T, tc testCase) {
	var subCalls uint64
	subs := [2]Subscriber[testConfig]{
		func(tc testConfig) error {
			subCalls++
			return nil
		},
		func(tc testConfig) error {
			return errors.New("test error")
		},
	}

	var cbCalls int
	cbs := [2]Callback[testConfig]{
		func(tc testConfig) { cbCalls++ },
		func(tc testConfig) { cbCalls++ },
	}

	c, err := setup(t, fmt.Sprintf(defaultConfig, string(tc.Type)), "", tc.Type, tc.TestString)
	require.NoErrorf(t, err, testSetupErrorMsg)

	for _, f := range subs {
		c.AddSubscriber(f)
	}

	for _, f := range cbs {
		c.AddCallback(f)
	}

	err = c.Update(newData)
	require.Errorf(t, err, "update config did not failed")

	want := testData
	got := c.Config()

	assert.Equalf(t, want, got, "config is not equal to old data")
	assert.NotEqualf(t, newData, got, "config was updated to new data")
	assert.NotEqualf(t, 1, (subCalls % 2), "updated subscriber is not rolled back: %d", subCalls)
	assert.Zero(t, cbCalls, "callbacks are called in case of subscriber error: %d", cbCalls)
}

func testUpdateConfigIsValidated(t *testing.T, tc testCase) {
	c, err := setup(t, fmt.Sprintf(defaultConfig, string(tc.Type)), "", tc.Type, tc.TestString)
	require.NoErrorf(t, err, testSetupErrorMsg)

	err = c.Update(newDataWithoutRequired)
	require.Errorf(t, err, "expected error not thrown")

	// config should not be updated
	got := c.Config()
	assert.Equalf(t, testData, got, expectedResultErrorMsg)
}

func test(t *testing.T, tc testCase, f func(*testing.T, testCase)) {
	fullName := strings.Split((runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()), ".")
	functionName := fullName[len(fullName)-1]

	t.Run(functionName, func(t *testing.T) {
		t.Cleanup(cleanup)
		f(t, tc)
	})
}

func setup(
	t *testing.T,
	file string,
	path string,
	fileType fh.FileType,
	data string,
) (*C[testConfig], error) {

	if path != "" {
		err := os.Mkdir(path, os.ModePerm)
		require.NoErrorf(t, err, "setup: error while creating directory")
	}

	f := filepath.Join(path, file)
	err := os.WriteFile(f, []byte(data), permissions)
	require.NoErrorf(t, err, "setup: error while write to file")

	h, err := fh.New(fh.WithName(appName), fh.WithPath(path), fh.WithType(fileType))
	require.NoErrorf(t, err, "setup: error while creating file handler")

	return Init[testConfig](h)
}

func cleanup() {
	for _, tc := range testCases {
		os.Remove(fmt.Sprintf(activeConfig, tc.Type))
		os.Remove(fmt.Sprintf(defaultConfig, tc.Type))
	}
	os.RemoveAll(testDir)
	os.Setenv("TEST_ENV_NAME", "")
}
