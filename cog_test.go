package cog

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	fh "github.com/leonidasdeim/cog/pkg/filehandler"
	"github.com/leonidasdeim/cog/pkg/utils"
)

const (
	permissions            = 0664
	appName                = "test_app"
	activeConfig           = appName + ".%s"
	defaultConfig          = appName + ".default.%s"
	testDir                = "testDir/"
	testSetupErrorMsg      = "Error while setting up test: %v"
	expectedResultErrorMsg = "Expected config does not match the result"
)

type TestConfig struct {
	Name      string `default:"app" env:"TEST_ENV_NAME"`
	Version   int    `validate:"required"`
	IsPrefork bool   `default:"true"`
}

var (
	testData            = TestConfig{Name: "config_test", Version: 123, IsPrefork: true}
	testDataDefaultName = TestConfig{Name: "app", Version: 123, IsPrefork: true}
	testDataEnvName     = TestConfig{Name: "env_name", Version: 123, IsPrefork: true}
)

type TestCaseForFileType struct {
	Type                     fh.FileType
	TestString               string
	TestStringWithoutVersion string
	TestStringWithDefaults   string
}

var testCases = []TestCaseForFileType{
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

func TestAllCases(t *testing.T) {
	for _, tc := range testCases {
		InitTests(t, tc)
		UpdateTests(t, tc)
	}
}

func setup(fn string, path string, ft fh.FileType, data string) (*C[TestConfig], error) {
	if path != "" {
		err := os.Mkdir(path, os.ModePerm)
		if err != nil {
			return nil, err
		}
	}

	f := filepath.Join(path, fn)
	err := os.WriteFile(f, []byte(data), permissions)
	if err != nil {
		return nil, err
	}

	h, err := fh.New(fh.WithName(appName), fh.WithPath(path), fh.WithType(ft))
	if err != nil {
		return nil, err
	}

	c, err := Init[TestConfig](h)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func cleanup() {
	for _, tc := range testCases {
		os.Remove(fmt.Sprintf(activeConfig, tc.Type))
		os.Remove(fmt.Sprintf(defaultConfig, tc.Type))
	}
	os.RemoveAll(testDir)
	os.Setenv("TEST_ENV_NAME", "")
}

func InitTests(t *testing.T, tc TestCaseForFileType) {
	t.Run("Check loaded config data "+string(tc.Type), func(t *testing.T) {
		t.Cleanup(cleanup)

		c, err := setup(fmt.Sprintf(defaultConfig, string(tc.Type)), "", tc.Type, tc.TestString)
		if err != nil {
			t.Errorf(testSetupErrorMsg, err)
			t.FailNow()
		}

		want := testData
		got := c.Config()

		if !reflect.DeepEqual(want, got) {
			t.Error(expectedResultErrorMsg)
		}
	})

	t.Run("Check if file data overwrites env variable "+string(tc.Type), func(t *testing.T) {
		t.Cleanup(cleanup)
		os.Setenv("TEST_ENV_NAME", "env_name")

		c, err := setup(fmt.Sprintf(defaultConfig, string(tc.Type)), "", tc.Type, tc.TestString)
		if err != nil {
			t.Errorf(testSetupErrorMsg, err)
			t.FailNow()
		}

		want := testData
		got := c.Config()

		if !reflect.DeepEqual(want, got) {
			t.Error(expectedResultErrorMsg)
		}
	})

	t.Run("Check default handler "+string(tc.Type), func(t *testing.T) {
		t.Cleanup(cleanup)

		type Connection struct {
			Host string `json:"host" default:"localhost"`
			Port string `json:"port" default:"123"`
		}

		type ConfigNoRequiredFields struct {
			Name      string `default:"app"`
			Version   int
			Store     Connection
			IsPrefork bool `default:"true"`
		}

		_, err := Init[ConfigNoRequiredFields]()
		if err != nil {
			t.Errorf("Error while initializing library: %v", err)
			t.FailNow()
		}

		if !utils.Exists("app.json") {
			t.Error("Expected active config file to be created, but it does not exist")
		}

		os.Remove("app.json")
	})

	t.Run("Check loaded config data from active config "+string(tc.Type), func(t *testing.T) {
		t.Cleanup(cleanup)

		c, err := setup(fmt.Sprintf(activeConfig, string(tc.Type)), "", tc.Type, tc.TestString)
		if err != nil {
			t.Errorf(testSetupErrorMsg, err)
			t.FailNow()
		}

		want := testData
		got := c.Config()

		if !reflect.DeepEqual(want, got) {
			t.Error(expectedResultErrorMsg)
		}
	})

	t.Run("Create active config file "+string(tc.Type), func(t *testing.T) {
		t.Cleanup(cleanup)

		_, err := setup(fmt.Sprintf(defaultConfig, string(tc.Type)), "", tc.Type, tc.TestString)
		if err != nil {
			t.Errorf(testSetupErrorMsg, err)
			t.FailNow()
		}

		if !utils.Exists(fmt.Sprintf(activeConfig, string(tc.Type))) {
			t.Error("Expected active config file to be created, but it does not exist")
		}
	})

	t.Run("Check active config file content "+string(tc.Type), func(t *testing.T) {
		t.Cleanup(cleanup)

		c, err := setup(fmt.Sprintf(defaultConfig, string(tc.Type)), "", tc.Type, tc.TestString)
		if err != nil {
			t.Errorf(testSetupErrorMsg, err)
			t.FailNow()
		}

		fileContent := TestConfig{}

		if err = c.handler.Load(&fileContent); err != nil {
			t.Error("Parsing activeConfig file", err.Error())
		}

		want := testData
		got := fileContent

		if !reflect.DeepEqual(want, got) {
			t.Error(expectedResultErrorMsg)
		}
	})

	t.Run("Check timestamp is created "+string(tc.Type), func(t *testing.T) {
		t.Cleanup(cleanup)

		c, err := setup(fmt.Sprintf(defaultConfig, string(tc.Type)), "", tc.Type, tc.TestString)
		if err != nil {
			t.Errorf(testSetupErrorMsg, err)
			t.FailNow()
		}

		if c.GetTimestamp() == "" {
			t.Error("Timestamp is not set")
		}
	})

	t.Run("Custom config path "+string(tc.Type), func(t *testing.T) {
		t.Cleanup(cleanup)

		c, err := setup(fmt.Sprintf(defaultConfig, string(tc.Type)), testDir, tc.Type, tc.TestString)
		if err != nil {
			t.Errorf(testSetupErrorMsg, err)
			t.FailNow()
		}

		defaultConfigPth := filepath.Join(testDir, fmt.Sprintf(defaultConfig, string(tc.Type)))
		if _, err := os.Stat(defaultConfigPth); err != nil {
			t.Error("Cannot find default config in expected location")
			t.FailNow()
		}

		activeConfigPth := filepath.Join(testDir, fmt.Sprintf(activeConfig, string(tc.Type)))
		if _, err := os.Stat(activeConfigPth); err != nil {
			t.Error("Cannot find active config in expected location")
			t.FailNow()
		}

		want := testData
		got := c.Config()

		if !reflect.DeepEqual(want, got) {
			t.Error(expectedResultErrorMsg)
			t.FailNow()
		}
	})

	t.Run("Check required fields validation "+string(tc.Type), func(t *testing.T) {
		t.Cleanup(cleanup)

		_, err := setup(fmt.Sprintf(defaultConfig, string(tc.Type)), "", tc.Type, tc.TestStringWithoutVersion)
		if err == nil {
			t.Errorf("Error is not returned")
			t.FailNow()
		}
		if !strings.Contains(err.Error(), "failed at validate config") {
			t.Errorf("Validation error is not returned")
		}
	})

	t.Run("Check if default values are set "+string(tc.Type), func(t *testing.T) {
		t.Cleanup(cleanup)

		c, err := setup(fmt.Sprintf(defaultConfig, string(tc.Type)), "", tc.Type, tc.TestStringWithDefaults)
		if err != nil {
			t.Errorf("Failed to set default values")
			t.FailNow()
		}

		want := testDataDefaultName
		got := c.Config()

		if !reflect.DeepEqual(want, got) {
			t.Error(expectedResultErrorMsg)
		}
	})

	t.Run("Check if environment values are set "+string(tc.Type), func(t *testing.T) {
		t.Cleanup(cleanup)
		os.Setenv("TEST_ENV_NAME", "env_name")

		c, err := setup(fmt.Sprintf(defaultConfig, string(tc.Type)), "", tc.Type, tc.TestStringWithDefaults)
		if err != nil {
			t.Errorf("Failed to set default values")
			t.FailNow()
		}

		want := testDataEnvName
		got := c.Config()

		if !reflect.DeepEqual(want, got) {
			t.Error(expectedResultErrorMsg)
		}
	})

	t.Run("Check if dynamic type is resolved correctly "+string(tc.Type), func(t *testing.T) {
		t.Cleanup(cleanup)

		c, err := setup(fmt.Sprintf(defaultConfig, string(tc.Type)), "", fh.DYNAMIC, tc.TestString)
		if err != nil {
			t.Errorf(testSetupErrorMsg, err)
			t.FailNow()
		}

		want := testData
		got := c.Config()

		if !reflect.DeepEqual(want, got) {
			t.Error(expectedResultErrorMsg)
		}

		if !utils.Exists(fmt.Sprintf(activeConfig, string(tc.Type))) {
			t.Error("Expected active config file to be created with correct filetype")
		}
	})

	t.Run("Check subscribers being registered "+string(tc.Type), func(t *testing.T) {
		t.Cleanup(cleanup)

		subs := [3]Subscriber[TestConfig]{
			func(tc TestConfig) error {
				return nil
			},
			func(tc TestConfig) error {
				return nil
			},
			func(tc TestConfig) error {
				return nil
			},
		}
		c, err := setup(fmt.Sprintf(defaultConfig, string(tc.Type)), "", tc.Type, tc.TestString)
		if err != nil {
			t.Errorf(testSetupErrorMsg, err)
			t.FailNow()
		}

		for _, cb := range subs {
			c.AddSubscriber(cb)
		}

		if len(c.subscribers) != len(subs) {
			t.Error("Expected number of subscribers is not correct")
		}
	})
}

func UpdateTests(t *testing.T, tc TestCaseForFileType) {
	newData := TestConfig{Name: "new_data", Version: 456}
	newDataWithoutRequired := TestConfig{Name: "new_data"}

	t.Run("Check if config is updated "+string(tc.Type), func(t *testing.T) {
		t.Cleanup(cleanup)

		c, err := setup(fmt.Sprintf(defaultConfig, string(tc.Type)), "", tc.Type, tc.TestString)
		if err != nil {
			t.Errorf(testSetupErrorMsg, err)
			t.FailNow()
		}

		err = c.Update(newData)
		if err != nil {
			t.Errorf("Error while updating config: %v", err)
			t.FailNow()
		}

		want := newData
		got := c.Config()

		if !reflect.DeepEqual(want, got) {
			t.Error(expectedResultErrorMsg)
		}
	})

	t.Run("Check subscribers are being notified "+string(tc.Type), func(t *testing.T) {
		t.Cleanup(cleanup)

		cb1 := 0
		cb2 := 0
		subs := [2]Subscriber[TestConfig]{
			func(tc TestConfig) error {
				cb1++
				return nil
			},
			func(tc TestConfig) error {
				cb2++
				return nil
			},
		}
		c, err := setup(fmt.Sprintf(defaultConfig, string(tc.Type)), "", tc.Type, tc.TestString)
		if err != nil {
			t.Errorf(testSetupErrorMsg, err)
			t.FailNow()
		}

		for _, f := range subs {
			c.AddSubscriber(f)
		}

		c.Update(newData)
		c.Update(newData)
		c.Update(newData)
		c.Update(newData)

		if cb1 != 4 || cb2 != 4 {
			t.Error("Bound subscribers are not being called")
		}
	})

	t.Run("Check subscribers error "+string(tc.Type), func(t *testing.T) {
		t.Cleanup(cleanup)

		var callTimes uint64
		subs := [2]Subscriber[TestConfig]{
			func(tc TestConfig) error {
				callTimes++
				return nil
			},
			func(tc TestConfig) error {
				return errors.New("test error")
			},
		}
		c, err := setup(fmt.Sprintf(defaultConfig, string(tc.Type)), "", tc.Type, tc.TestString)
		if err != nil {
			t.Errorf(testSetupErrorMsg, err)
			t.FailNow()
		}

		for _, f := range subs {
			c.AddSubscriber(f)
		}

		err = c.Update(newData)
		if err == nil {
			t.Error("config update should fail")
			t.FailNow()
		}
		t.Logf("error: %v", err)

		want := testData
		got := c.Config()

		if reflect.DeepEqual(newData, got) {
			t.Error("config was updated to new data")
		}

		if !reflect.DeepEqual(want, got) {
			t.Error("config is not equal to old data")
		}

		if (callTimes % 2) == 1 {
			t.Errorf("Updated subscriber is not rolled back: %d", callTimes)
		}
	})

	t.Run("Check if config is validated "+string(tc.Type), func(t *testing.T) {
		t.Cleanup(cleanup)

		c, err := setup(fmt.Sprintf(defaultConfig, string(tc.Type)), "", tc.Type, tc.TestString)
		if err != nil {
			t.Errorf(testSetupErrorMsg, err)
			t.FailNow()
		}

		err = c.Update(newDataWithoutRequired)
		if err == nil {
			t.Errorf("Expected error not thrown: %v", err)
			t.FailNow()
		}

		// config should not be updated
		want := testData
		got := c.Config()

		if !reflect.DeepEqual(want, got) {
			t.Error(expectedResultErrorMsg)
		}
	})
}
