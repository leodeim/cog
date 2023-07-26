package cog

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	fh "github.com/leonidasdeim/cog/filehandler"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
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

type cogTestSuite struct {
	suite.Suite
	tc testCase
}

func TestFileTypes(t *testing.T) {
	for _, tc := range testCases {
		suite.Run(t, &cogTestSuite{tc: tc})
	}
}

func (s *cogTestSuite) TearDownTest() {
	for _, tc := range testCases {
		os.Remove(fmt.Sprintf(activeConfig, tc.Type))
		os.Remove(fmt.Sprintf(defaultConfig, tc.Type))
	}
	os.RemoveAll(testDir)
	os.Setenv("TEST_ENV_NAME", "")
}

func (s *cogTestSuite) TestConfigLoading() {
	c, err := setup(s.T(), fmt.Sprintf(defaultConfig, string(s.tc.Type)), "", s.tc.Type, s.tc.TestString)
	require.NoErrorf(s.T(), err, testSetupErrorMsg)

	got := c.Config()
	assert.Equalf(s.T(), testData, got, expectedResultErrorMsg)
}

func (s *cogTestSuite) TestEnvironmentVarIsOverwritten() {
	os.Setenv("TEST_ENV_NAME", "env_name")

	c, err := setup(s.T(), fmt.Sprintf(defaultConfig, string(s.tc.Type)), "", s.tc.Type, s.tc.TestString)
	require.NoErrorf(s.T(), err, testSetupErrorMsg)

	got := c.Config()
	assert.Equalf(s.T(), testData, got, expectedResultErrorMsg)
}

func (s *cogTestSuite) TestConfigDefaults() {
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
	require.NoErrorf(s.T(), err, testSetupErrorMsg)

	assert.FileExistsf(s.T(), "app.json", "active config file is not created")
	assert.Equalf(s.T(), "app", c.Config().Name, "default name is not set")
	assert.Equalf(s.T(), true, c.Config().IsPrefork, "default isPrefork is not set")
	assert.Equalf(s.T(), "localhost", c.Config().Store.Host, "default host is not set")
	assert.Equalf(s.T(), "123", c.Config().Store.Port, "default port is not set")

	os.Remove("app.json")
}

func (s *cogTestSuite) TestLoadFromActiveConfig() {
	c, err := setup(s.T(), fmt.Sprintf(activeConfig, string(s.tc.Type)), "", s.tc.Type, s.tc.TestString)
	require.NoErrorf(s.T(), err, testSetupErrorMsg)

	got := c.Config()
	assert.Equalf(s.T(), testData, got, expectedResultErrorMsg)
}

func (s *cogTestSuite) TestActiveConfigCreated() {
	_, err := setup(s.T(), fmt.Sprintf(defaultConfig, string(s.tc.Type)), "", s.tc.Type, s.tc.TestString)
	require.NoErrorf(s.T(), err, testSetupErrorMsg)

	assert.FileExistsf(s.T(), fmt.Sprintf(activeConfig, string(s.tc.Type)), "active config file is not created")
}

func (s *cogTestSuite) TestActiveConfigContent() {
	c, err := setup(s.T(), fmt.Sprintf(defaultConfig, string(s.tc.Type)), "", s.tc.Type, s.tc.TestString)
	require.NoErrorf(s.T(), err, testSetupErrorMsg)

	got := testConfig{}
	err = c.handler.Load(&got)
	assert.NoErrorf(s.T(), err, "error while parsing active config file")

	assert.Equalf(s.T(), testData, got, expectedResultErrorMsg)
}

func (s *cogTestSuite) TestTimestampIsCreated() {
	c, err := setup(s.T(), fmt.Sprintf(defaultConfig, string(s.tc.Type)), "", s.tc.Type, s.tc.TestString)
	require.NoErrorf(s.T(), err, testSetupErrorMsg)

	assert.NotEmptyf(s.T(), c.GetTimestamp(), "timestamp is not set")
}

func (s *cogTestSuite) TestCustomConfigPath() {
	c, err := setup(s.T(), fmt.Sprintf(defaultConfig, string(s.tc.Type)), testDir, s.tc.Type, s.tc.TestString)
	require.NoErrorf(s.T(), err, testSetupErrorMsg)

	defaultConfigPath := filepath.Join(testDir, fmt.Sprintf(defaultConfig, string(s.tc.Type)))
	assert.FileExists(s.T(), defaultConfigPath, "cannot find default config in expected location")

	activeConfigPth := filepath.Join(testDir, fmt.Sprintf(activeConfig, string(s.tc.Type)))
	assert.FileExists(s.T(), activeConfigPth, "cannot find active config in expected location")

	got := c.Config()
	assert.Equalf(s.T(), testData, got, expectedResultErrorMsg)
}

func (s *cogTestSuite) TestDataWithoutRequiredField() {
	c, err := setup(s.T(), fmt.Sprintf(defaultConfig, string(s.tc.Type)), "", s.tc.Type, s.tc.TestStringWithoutVersion)
	require.Errorf(s.T(), err, "error is not returned")
	require.Nilf(s.T(), c, "cog instance should be nil")

	assert.Containsf(s.T(), err.Error(), "failed at validate config", "wrong error is returned")
}

func (s *cogTestSuite) TestDefaultValuesAreSet() {
	c, err := setup(s.T(), fmt.Sprintf(defaultConfig, string(s.tc.Type)), "", s.tc.Type, s.tc.TestStringWithDefaults)
	require.NoErrorf(s.T(), err, testSetupErrorMsg)

	got := c.Config()
	assert.Equalf(s.T(), testDataDefaultName, got, expectedResultErrorMsg)
}

func (s *cogTestSuite) TestEnvironmentValuesAreSet() {
	os.Setenv("TEST_ENV_NAME", "env_name")

	c, err := setup(s.T(), fmt.Sprintf(defaultConfig, string(s.tc.Type)), "", s.tc.Type, s.tc.TestStringWithDefaults)
	require.NoErrorf(s.T(), err, testSetupErrorMsg)

	got := c.Config()
	assert.Equalf(s.T(), testDataEnvName, got, expectedResultErrorMsg)
}

func (s *cogTestSuite) TestDynamicTypeIsResolved() {
	c, err := setup(s.T(), fmt.Sprintf(defaultConfig, string(s.tc.Type)), "", fh.DYNAMIC, s.tc.TestString)
	require.NoErrorf(s.T(), err, testSetupErrorMsg)

	got := c.Config()
	assert.Equalf(s.T(), testData, got, expectedResultErrorMsg)

	assert.FileExistsf(s.T(), fmt.Sprintf(activeConfig, string(s.tc.Type)), "expected active config file not exists")
}

func (s *cogTestSuite) TestSubscribersAreRegistered() {
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

	c, err := setup(s.T(), fmt.Sprintf(defaultConfig, string(s.tc.Type)), "", s.tc.Type, s.tc.TestString)
	require.NoErrorf(s.T(), err, testSetupErrorMsg)

	for _, cb := range subs {
		c.AddSubscriber(cb)
	}

	assert.Equalf(s.T(), len(subs), len(c.subscribers), "expected number of subscribers")
}

func (s *cogTestSuite) TestCallbacksAreRegistered() {
	cbs := [3]Callback[testConfig]{
		func(tc testConfig) {},
		func(tc testConfig) {},
		func(tc testConfig) {},
	}
	c, err := setup(s.T(), fmt.Sprintf(defaultConfig, string(s.tc.Type)), "", s.tc.Type, s.tc.TestString)
	require.NoErrorf(s.T(), err, testSetupErrorMsg)

	for _, f := range cbs {
		c.AddCallback(f)
	}

	assert.Equalf(s.T(), len(cbs), len(c.callbacks), "expected number of callbacks")
}

var (
	newData                = testConfig{Name: "new_data", Version: 456}
	newDataWithoutRequired = testConfig{Name: "new_data"}
)

func (s *cogTestSuite) TestConfigUpdated() {
	c, err := setup(s.T(), fmt.Sprintf(defaultConfig, string(s.tc.Type)), "", s.tc.Type, s.tc.TestString)
	require.NoErrorf(s.T(), err, testSetupErrorMsg)

	err = c.Update(newData)
	require.NoErrorf(s.T(), err, "error while updating config: %v", err)

	got := c.Config()
	assert.Equalf(s.T(), newData, got, expectedResultErrorMsg)
}

func (s *cogTestSuite) TestCallbacksAreNotifiedAndRemoved() {
	var calls1, calls2 int
	cbs := [3]Callback[testConfig]{
		func(tc testConfig) { calls1++ },
		func(tc testConfig) { calls2++ },
		nil,
	}

	c, err := setup(s.T(), fmt.Sprintf(defaultConfig, string(s.tc.Type)), "", s.tc.Type, s.tc.TestString)
	require.NoErrorf(s.T(), err, testSetupErrorMsg)

	c.AddCallback(cbs[0])
	callbackId := c.AddCallback(cbs[1])
	c.AddCallback(cbs[2])

	c.Update(newData)
	c.Update(newData)
	time.Sleep(100 * time.Millisecond)

	c.RemoveCallback(callbackId)

	c.Update(newData)
	c.Update(newData)
	time.Sleep(100 * time.Millisecond)

	assert.Equal(s.T(), 4, calls1)
	assert.Equal(s.T(), 2, calls2)
}

func (s *cogTestSuite) TestRemoveCallbackWrongId() {
	c, err := setup(s.T(), fmt.Sprintf(defaultConfig, string(s.tc.Type)), "", s.tc.Type, s.tc.TestString)
	require.NoErrorf(s.T(), err, testSetupErrorMsg)

	id := c.AddCallback(func(tc testConfig) {})

	err = c.RemoveCallback(id + 1)
	require.Errorf(s.T(), err, "RemoveCallback should return error")
}

func (s *cogTestSuite) TestSubscribersAreNotifiedAndRemoved() {
	var calls1, calls2 int
	subs := [3]Subscriber[testConfig]{
		func(tc testConfig) error {
			calls1++
			return nil
		},
		func(tc testConfig) error {
			calls2++
			return nil
		},
		nil,
	}

	c, err := setup(s.T(), fmt.Sprintf(defaultConfig, string(s.tc.Type)), "", s.tc.Type, s.tc.TestString)
	require.NoErrorf(s.T(), err, testSetupErrorMsg)

	c.AddSubscriber(subs[0])
	subscriberId := c.AddSubscriber(subs[1])
	c.AddSubscriber(subs[2])

	c.Update(newData)
	c.Update(newData)
	time.Sleep(100 * time.Millisecond)

	c.RemoveSubscriber(subscriberId)

	c.Update(newData)
	c.Update(newData)
	time.Sleep(100 * time.Millisecond)

	assert.Equal(s.T(), 4, calls1)
	assert.Equal(s.T(), 2, calls2)
}

func (s *cogTestSuite) TestRemoveSubscriberWrongId() {
	c, err := setup(s.T(), fmt.Sprintf(defaultConfig, string(s.tc.Type)), "", s.tc.Type, s.tc.TestString)
	require.NoErrorf(s.T(), err, testSetupErrorMsg)

	id := c.AddSubscriber(func(tc testConfig) error { return nil })

	err = c.RemoveSubscriber(id + 1)
	require.Errorf(s.T(), err, "RemoveCallback should return error")
}

func (s *cogTestSuite) TestSubscriberReturnsError() {
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

	c, err := setup(s.T(), fmt.Sprintf(defaultConfig, string(s.tc.Type)), "", s.tc.Type, s.tc.TestString)
	require.NoErrorf(s.T(), err, testSetupErrorMsg)

	for _, f := range subs {
		c.AddSubscriber(f)
	}

	for _, f := range cbs {
		c.AddCallback(f)
	}

	err = c.Update(newData)
	require.Errorf(s.T(), err, "update config did not failed")

	want := testData
	got := c.Config()

	assert.Equalf(s.T(), want, got, "config is not equal to old data")
	assert.NotEqualf(s.T(), newData, got, "config was updated to new data")
	assert.NotEqualf(s.T(), 1, (subCalls % 2), "updated subscriber is not rolled back: %d", subCalls)
	assert.Zero(s.T(), cbCalls, "callbacks are called in case of subscriber error: %d", cbCalls)
}

func (s *cogTestSuite) TestUpdateConfigIsValidated() {
	c, err := setup(s.T(), fmt.Sprintf(defaultConfig, string(s.tc.Type)), "", s.tc.Type, s.tc.TestString)
	require.NoErrorf(s.T(), err, testSetupErrorMsg)

	err = c.Update(newDataWithoutRequired)
	require.Errorf(s.T(), err, "expected error not thrown")

	// config should not be updated
	got := c.Config()
	assert.Equalf(s.T(), testData, got, expectedResultErrorMsg)
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
