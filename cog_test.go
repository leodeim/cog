package cog

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

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
			t.Errorf("error while initializing library: %v", err)
			t.FailNow()
		}

		if !utils.Exists("app.json") {
			t.Error("expected active config file to be created, but it does not exist")
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
			t.Error("expected active config file to be created, but it does not exist")
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
			t.Error("parsing activeConfig file", err.Error())
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
			t.Error("timestamp is not set")
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
			t.Error("cannot find default config in expected location")
			t.FailNow()
		}

		activeConfigPth := filepath.Join(testDir, fmt.Sprintf(activeConfig, string(tc.Type)))
		if _, err := os.Stat(activeConfigPth); err != nil {
			t.Error("cannot find active config in expected location")
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
			t.Errorf("error is not returned")
			t.FailNow()
		}
		if !strings.Contains(err.Error(), "failed at validate config") {
			t.Errorf("validation error is not returned")
		}
	})

	t.Run("Check if default values are set "+string(tc.Type), func(t *testing.T) {
		t.Cleanup(cleanup)

		c, err := setup(fmt.Sprintf(defaultConfig, string(tc.Type)), "", tc.Type, tc.TestStringWithDefaults)
		if err != nil {
			t.Errorf("failed to set default values")
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
			t.Errorf("failed to set default values")
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
			t.Error("expected active config file to be created with correct filetype")
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
			t.Error("expected number of subscribers is not correct")
		}
	})

	t.Run("Check callbacks being registered "+string(tc.Type), func(t *testing.T) {
		t.Cleanup(cleanup)

		cbs := [3]Callback[TestConfig]{
			func(tc TestConfig) {},
			func(tc TestConfig) {},
			func(tc TestConfig) {},
		}
		c, err := setup(fmt.Sprintf(defaultConfig, string(tc.Type)), "", tc.Type, tc.TestString)
		if err != nil {
			t.Errorf(testSetupErrorMsg, err)
			t.FailNow()
		}
		for _, f := range cbs {
			c.AddCallback(f)
		}

		if len(c.callbacks) != len(cbs) {
			t.Error("expected number of callbacks is not correct")
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
			t.Errorf("error while updating config: %v", err)
			t.FailNow()
		}

		want := newData
		got := c.Config()

		if !reflect.DeepEqual(want, got) {
			t.Error(expectedResultErrorMsg)
		}
	})

	t.Run("Check callbacks are being notified "+string(tc.Type), func(t *testing.T) {
		t.Cleanup(cleanup)

		var calls1, calls2 int
		cbs := [2]Callback[TestConfig]{
			func(tc TestConfig) { calls1++ },
			func(tc TestConfig) { calls2++ },
		}
		c, err := setup(fmt.Sprintf(defaultConfig, string(tc.Type)), "", tc.Type, tc.TestString)
		if err != nil {
			t.Errorf(testSetupErrorMsg, err)
			t.FailNow()
		}

		c.AddCallback(cbs[0])
		callbackId := c.AddCallback(cbs[1])

		c.Update(newData)
		c.Update(newData)
		time.Sleep(100 * time.Millisecond)

		c.RemoveCallback(callbackId)

		c.Update(newData)
		c.Update(newData)
		time.Sleep(100 * time.Millisecond)

		if calls1 != 4 || calls2 != 2 {
			t.Errorf("wrong number of calls for callbacks: calls1=%d calls2=%d", calls1, calls2)
		}
	})

	t.Run("Check subscribers are being notified "+string(tc.Type), func(t *testing.T) {
		t.Cleanup(cleanup)

		var calls1, calls2 int
		subs := [2]Subscriber[TestConfig]{
			func(tc TestConfig) error {
				calls1++
				return nil
			},
			func(tc TestConfig) error {
				calls2++
				return nil
			},
		}
		c, err := setup(fmt.Sprintf(defaultConfig, string(tc.Type)), "", tc.Type, tc.TestString)
		if err != nil {
			t.Errorf(testSetupErrorMsg, err)
			t.FailNow()
		}

		c.AddSubscriber(subs[0])
		cb2Id := c.AddSubscriber(subs[1])

		c.Update(newData)
		c.Update(newData)
		time.Sleep(100 * time.Millisecond)

		c.RemoveSubscriber(cb2Id)

		c.Update(newData)
		c.Update(newData)
		time.Sleep(100 * time.Millisecond)

		if calls1 != 4 || calls2 != 2 {
			t.Errorf("wrong number of calls for subscribers: calls1=%d calls2=%d", calls1, calls2)
		}
	})

	t.Run("Check subscribers error "+string(tc.Type), func(t *testing.T) {
		t.Cleanup(cleanup)

		var subCalls uint64
		subs := [2]Subscriber[TestConfig]{
			func(tc TestConfig) error {
				subCalls++
				return nil
			},
			func(tc TestConfig) error {
				return errors.New("test error")
			},
		}
		var cbCalls int
		cbs := [2]Callback[TestConfig]{
			func(tc TestConfig) { cbCalls++ },
			func(tc TestConfig) { cbCalls++ },
		}
		c, err := setup(fmt.Sprintf(defaultConfig, string(tc.Type)), "", tc.Type, tc.TestString)
		if err != nil {
			t.Errorf(testSetupErrorMsg, err)
			t.FailNow()
		}

		for _, f := range subs {
			c.AddSubscriber(f)
		}

		for _, f := range cbs {
			c.AddCallback(f)
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

		if (subCalls % 2) == 1 {
			t.Errorf("updated subscriber is not rolled back: %d", subCalls)
		}

		if cbCalls > 0 {
			t.Errorf("callbacks are called in case of subscriber error: %d", cbCalls)
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
			t.Errorf("expected error not thrown: %v", err)
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
