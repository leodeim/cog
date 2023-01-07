package goconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/leonidasdeim/goconfig/internal/files"
	"github.com/leonidasdeim/goconfig/pkg/handler"
)

type TestConfig struct {
	Name      string `default:"app"`
	Version   int    `validate:"required"`
	IsPrefork bool   `default:"true"`
}

var testData = TestConfig{Name: "config_test", Version: 123, IsPrefork: true}
var testDataDefaultName = TestConfig{Name: "app", Version: 123, IsPrefork: true}

const permissions = 0664
const appName = "test_app"
const activeConfig = appName + ".%s"
const defaultConfig = appName + ".default.%s"
const testDir = "testDir/"

type TestCaseForFileType struct {
	Type                     handler.FileType
	TestString               string
	TestStringWithoutVersion string
	TestStringWithDefaults   string
}

var testCases = []TestCaseForFileType{
	{
		handler.JSON,
		"{\"name\":\"config_test\",\"version\":123}",
		"{\"name\":\"config_test\"}",
		"{\"version\":123}",
	},
	{
		handler.YAML,
		"name: config_test\nversion: 123\n",
		"name: config_test\n",
		"version: 123\n",
	},
}

func Test_AllCases(t *testing.T) {
	for _, ft := range testCases {
		InitTests(t, ft)
		UpdateTests(t, ft)
	}
}

func setUp(fn string, path string, ft handler.FileType, data string, subs []string) (*Config[TestConfig], error) {
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

	h, err := handler.New(handler.WithName(appName), handler.WithPath(path), handler.WithType(ft))
	if err != nil {
		return nil, err
	}

	c, err := Init[TestConfig](h)
	if err != nil {
		return nil, err
	}

	for _, s := range subs {
		c.AddSubscriber(s)
	}

	return c, nil
}

func cleanUp() {
	for _, tc := range testCases {
		os.Remove(fmt.Sprintf(activeConfig, tc.Type))
		os.Remove(fmt.Sprintf(defaultConfig, tc.Type))
	}
	os.RemoveAll(testDir)
}

func InitTests(t *testing.T, tc TestCaseForFileType) {
	t.Run("Check loaded config data "+string(tc.Type), func(t *testing.T) {
		t.Cleanup(cleanUp)

		c, err := setUp(fmt.Sprintf(defaultConfig, string(tc.Type)), "", tc.Type, tc.TestString, []string{})
		if err != nil {
			t.Errorf("Error while setting up test: %v", err)
			t.FailNow()
		}

		want := testData
		got := c.GetCfg()

		if !reflect.DeepEqual(want, got) {
			t.Error("Expected config does not match the result")
		}
	})

	t.Run("Check loaded config data from active config "+string(tc.Type), func(t *testing.T) {
		t.Cleanup(cleanUp)

		c, err := setUp(fmt.Sprintf(activeConfig, string(tc.Type)), "", tc.Type, tc.TestString, []string{})
		if err != nil {
			t.Errorf("Error while setting up test: %v", err)
			t.FailNow()
		}

		want := testData
		got := c.GetCfg()
		if !reflect.DeepEqual(want, got) {
			t.Error("Expected config does not match the result")
		}
	})

	t.Run("Create active config file "+string(tc.Type), func(t *testing.T) {
		t.Cleanup(cleanUp)

		_, err := setUp(fmt.Sprintf(activeConfig, string(tc.Type)), "", tc.Type, tc.TestString, []string{})
		if err != nil {
			t.Errorf("Error while setting up test: %v", err)
			t.FailNow()
		}

		if !files.Exists(fmt.Sprintf(activeConfig, string(tc.Type))) {
			t.Error("Expected active config file to be created, but it does not exist")
		}
		os.Remove(fmt.Sprintf(activeConfig, string(tc.Type)))
	})

	t.Run("Check active config file content "+string(tc.Type), func(t *testing.T) {
		t.Cleanup(cleanUp)

		c, err := setUp(fmt.Sprintf(defaultConfig, string(tc.Type)), "", tc.Type, tc.TestString, []string{})
		if err != nil {
			t.Errorf("Error while setting up test: %v", err)
			t.FailNow()
		}

		fileContent := TestConfig{}

		if err = c.handler.Load(&fileContent); err != nil {
			t.Error("Parsing activeConfig file", err.Error())
		}

		want := testData
		got := fileContent

		if !reflect.DeepEqual(want, got) {
			t.Error("Expected config does not match the result")
		}
	})

	t.Run("Check timestamp is created "+string(tc.Type), func(t *testing.T) {
		t.Cleanup(cleanUp)

		c, err := setUp(fmt.Sprintf(defaultConfig, string(tc.Type)), "", tc.Type, tc.TestString, []string{})
		if err != nil {
			t.Errorf("Error while setting up test: %v", err)
			t.FailNow()
		}

		if c.GetTimestamp() == "" {
			t.Error("Timestamp is not set")
		}
	})

	t.Run("Check subscribers being created "+string(tc.Type), func(t *testing.T) {
		t.Cleanup(cleanUp)

		subscribers := [5]string{"test1", "test2", "test3", "test4", "test5"}
		c, err := setUp(fmt.Sprintf(defaultConfig, string(tc.Type)), "", tc.Type, tc.TestString, subscribers[:])
		if err != nil {
			t.Errorf("Error while setting up test: %v", err)
			t.FailNow()
		}

		if len(c.subs) != len(subscribers) {
			t.Error("Expected number of subscribers is not correct")
		}
	})

	t.Run("Check subscribers not being notified "+string(tc.Type), func(t *testing.T) {
		t.Cleanup(cleanUp)

		subscribers := [5]string{"test1"}
		c, err := setUp(fmt.Sprintf(defaultConfig, string(tc.Type)), "", tc.Type, tc.TestString, subscribers[:])
		if err != nil {
			t.Error("Error while setting up test")
			t.FailNow()
		}

		if len(c.GetSubscriber("test1")) != 0 {
			t.Error("Subscribers has been notified")
		}
	})

	t.Run("Custom config path "+string(tc.Type), func(t *testing.T) {
		t.Cleanup(cleanUp)

		c, err := setUp(fmt.Sprintf(defaultConfig, string(tc.Type)), testDir, tc.Type, tc.TestString, []string{})
		if err != nil {
			t.Errorf("Error while setting up test: %v", err)
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
		got := c.GetCfg()

		if !reflect.DeepEqual(want, got) {
			t.Error("Expected config does not match the result")
			t.FailNow()
		}
	})

	t.Run("Check required fields validation "+string(tc.Type), func(t *testing.T) {
		t.Cleanup(cleanUp)

		_, err := setUp(fmt.Sprintf(defaultConfig, string(tc.Type)), "", tc.Type, tc.TestStringWithoutVersion, []string{})
		if err == nil {
			t.Errorf("Error is not returned")
			t.FailNow()
		}
		if !strings.Contains(err.Error(), "failed at validate config") {
			t.Errorf("Validation error is not returned")
		}
	})

	t.Run("Check if default values are set "+string(tc.Type), func(t *testing.T) {
		t.Cleanup(cleanUp)

		c, err := setUp(fmt.Sprintf(defaultConfig, string(tc.Type)), "", tc.Type, tc.TestStringWithDefaults, []string{})
		if err != nil {
			t.Errorf("Failed to set default values")
			t.FailNow()
		}

		want := testDataDefaultName
		got := c.GetCfg()
		if !reflect.DeepEqual(want, got) {
			t.Error("Expected config does not match the result")
		}
	})
}

func UpdateTests(t *testing.T, tc TestCaseForFileType) {
	newData := TestConfig{Name: "new_data", Version: 456}

	t.Run("Check if config is updated "+string(tc.Type), func(t *testing.T) {
		t.Cleanup(cleanUp)

		c, err := setUp(fmt.Sprintf(defaultConfig, string(tc.Type)), "", tc.Type, tc.TestString, []string{})
		if err != nil {
			t.Errorf("Error while setting up test: %v", err)
			t.FailNow()
		}

		c.Update(newData)

		want := newData
		got := c.GetCfg()
		if !reflect.DeepEqual(want, got) {
			t.Error("Expected config does not match the result")
		}
	})

	t.Run("Check if subscribers are being notified "+string(tc.Type), func(t *testing.T) {
		t.Cleanup(cleanUp)

		subscribers := [5]string{"test1", "test2", "test3"}
		c, err := setUp(fmt.Sprintf(defaultConfig, string(tc.Type)), "", tc.Type, tc.TestString, subscribers[:])
		if err != nil {
			t.Errorf("Error while setting up test: %v", err)
			t.FailNow()
		}

		c.Update(newData)

		if len(c.subs["test1"]) != 1 || len(c.subs["test2"]) != 1 || len(c.subs["test3"]) != 1 {
			t.Error("Subscribers not being notified")
		}
	})

	t.Run("Check channel read "+string(tc.Type), func(t *testing.T) {
		t.Cleanup(cleanUp)

		subscribers := [1]string{"test1"}
		c, err := setUp(fmt.Sprintf(defaultConfig, string(tc.Type)), "", tc.Type, tc.TestString, subscribers[:])
		if err != nil {
			t.Errorf("Error while setting up test: %v", err)
			t.FailNow()
		}

		c.Update(newData)
		ch := c.GetSubscriber("test1")

		select {
		case <-ch:
			return
		default:
			t.Error("Channel not notified")
			t.FailNow()
		}
	})

	t.Run("Check if channels not being overloaded "+string(tc.Type), func(t *testing.T) {
		t.Cleanup(cleanUp)

		subscribers := [1]string{"test1"}
		c, err := setUp(fmt.Sprintf(defaultConfig, string(tc.Type)), "", tc.Type, tc.TestString, subscribers[:])
		if err != nil {
			t.Errorf("Error while setting up test: %v", err)
			t.FailNow()
		}

		c.Update(newData)
		c.Update(newData)
		c.Update(newData)

		if len(c.subs["test1"]) != 1 {
			t.Error("Subscribers not being notified")
		}
	})
}
