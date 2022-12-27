package goconfig

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

type TestConfig struct {
	Name      string `default:"app"`
	Version   int    `validate:"required"`
	IsPrefork bool   `default:"true"`
}

var testData = TestConfig{Name: "config_test", Version: 123, IsPrefork: true}
var testDataWithDefaultName = TestConfig{Name: "app", Version: 123, IsPrefork: true}

const testConfigName = "test"
const testDir = "testDir/"
const testString = "{\"name\":\"config_test\",\"version\":123}"
const testStringWithoutVersion = "{\"name\":\"config_test\"}"
const testStringWithDefaults = "{\"version\":123}"

func setUp(file string, path string, data string, subscribers []string) (*Config[TestConfig], error) {
	if path != "" {
		err := os.Mkdir(path, os.ModePerm)
		if err != nil {
			return nil, err
		}
	}
	configName := filepath.Join(path, fmt.Sprintf(file, testConfigName))
	err := os.WriteFile(configName, []byte(data), permissionRwRwR)

	if err != nil {
		return nil, err
	}

	c, err := Init[TestConfig](WithName(testConfigName), WithPath(path))
	if err != nil {
		return nil, err
	}

	for _, subscriber := range subscribers {
		c.AddSubscriber(subscriber)
	}

	return c, nil
}

func cleanUp() {
	os.Remove(fmt.Sprintf(defaultConfig, testConfigName))
	os.Remove(fmt.Sprintf(activeConfig, testConfigName))
	os.RemoveAll(testDir)
}

func Test_Init(t *testing.T) {
	t.Run("No configuration files", func(t *testing.T) {
		_, err := Init[TestConfig](WithName("not_exist"))
		if err == nil {
			t.Errorf("Error is not returned")
		}
	})

	t.Run("Check loaded config data", func(t *testing.T) {
		t.Cleanup(cleanUp)

		c, err := setUp(defaultConfig, "", testString, []string{})
		if err != nil {
			t.Errorf("Error while setting up test: %v", err)
			t.FailNow()
		}

		want := testData
		got := *c.GetCfg()

		if !reflect.DeepEqual(want, got) {
			t.Error("Expected config does not match the result")
		}
	})

	t.Run("Check loaded config data from active config", func(t *testing.T) {
		t.Cleanup(cleanUp)

		c, err := setUp(activeConfig, "", testString, []string{})
		if err != nil {
			t.Errorf("Error while setting up test: %v", err)
			t.FailNow()
		}

		want := testData
		got := *c.GetCfg()
		if !reflect.DeepEqual(want, got) {
			t.Error("Expected config does not match the result")
		}
	})

	t.Run("Create active config file", func(t *testing.T) {
		t.Cleanup(cleanUp)

		_, err := setUp(activeConfig, "", testString, []string{})
		if err != nil {
			t.Errorf("Error while setting up test: %v", err)
			t.FailNow()
		}

		if !fileExists(fmt.Sprintf(activeConfig, testConfigName)) {
			t.Error("Expected active config file to be created, but it does not exist")
		}
		os.Remove(fmt.Sprintf(activeConfig, testConfigName))
	})

	t.Run("Check active config file content", func(t *testing.T) {
		t.Cleanup(cleanUp)

		_, err := setUp(defaultConfig, "", testString, []string{})
		if err != nil {
			t.Errorf("Error while setting up test: %v", err)
			t.FailNow()
		}

		fileContent := TestConfig{}
		configFile, err := os.Open(fmt.Sprintf(activeConfig, testConfigName))
		if err != nil {
			t.Error("Opening activeConfig file", err.Error())
		}

		jsonParser := json.NewDecoder(configFile)
		if err = jsonParser.Decode(&fileContent); err != nil {
			t.Error("Parsing activeConfig file", err.Error())
		}

		want := testData
		got := fileContent

		if !reflect.DeepEqual(want, got) {
			t.Error("Expected config does not match the result")
		}
	})

	t.Run("Check timestamp is created", func(t *testing.T) {
		t.Cleanup(cleanUp)

		c, err := setUp(defaultConfig, "", testString, []string{})
		if err != nil {
			t.Errorf("Error while setting up test: %v", err)
			t.FailNow()
		}

		if c.GetTimestamp() == "" {
			t.Error("Timestamp is not set")
		}
	})

	t.Run("Check subscribers being created", func(t *testing.T) {
		t.Cleanup(cleanUp)

		subscribers := [5]string{"test1", "test2", "test3", "test4", "test5"}
		c, err := setUp(defaultConfig, "", testString, subscribers[:])
		if err != nil {
			t.Errorf("Error while setting up test: %v", err)
			t.FailNow()
		}

		if len(c.subscribers) != len(subscribers) {
			t.Error("Expected number of subscribers is not correct")
		}
	})

	t.Run("Check subscribers not being notified", func(t *testing.T) {
		t.Cleanup(cleanUp)

		subscribers := [5]string{"test1"}
		c, err := setUp(defaultConfig, "", testString, subscribers[:])
		if err != nil {
			t.Error("Error while setting up test")
			t.FailNow()
		}

		if len(c.GetSubscriber("test1")) != 0 {
			t.Error("Subscribers has been notified")
		}
	})

	t.Run("Custom config path", func(t *testing.T) {
		t.Cleanup(cleanUp)

		c, err := setUp(defaultConfig, testDir, testString, []string{})
		if err != nil {
			t.Errorf("Error while setting up test: %v", err)
			t.FailNow()
		}

		defaultConfigPth := filepath.Join(testDir, fmt.Sprintf(defaultConfig, testConfigName))
		if _, err := os.Stat(defaultConfigPth); err != nil {
			t.Error("Cannot find default config in expected location")
			t.FailNow()
		}

		activeConfigPth := filepath.Join(testDir, fmt.Sprintf(activeConfig, testConfigName))
		if _, err := os.Stat(activeConfigPth); err != nil {
			t.Error("Cannot find active config in expected location")
			t.FailNow()
		}

		want := testData
		got := *c.GetCfg()

		if !reflect.DeepEqual(want, got) {
			t.Error("Expected config does not match the result")
			t.FailNow()
		}
	})

	t.Run("Check required fields validation", func(t *testing.T) {
		t.Cleanup(cleanUp)

		_, err := setUp(defaultConfig, "", testStringWithoutVersion, []string{})
		if err == nil {
			t.Errorf("Error is not returned")
			t.FailNow()
		}
		if !strings.Contains(err.Error(), "failed at validate config") {
			t.Errorf("Validation error is not returned")
		}
	})

	t.Run("Check if default values are set", func(t *testing.T) {
		t.Cleanup(cleanUp)

		c, err := setUp(defaultConfig, "", testStringWithDefaults, []string{})
		if err != nil {
			t.Errorf("Failed to set default values")
			t.FailNow()
		}

		want := testDataWithDefaultName
		got := *c.GetCfg()
		if !reflect.DeepEqual(want, got) {
			t.Error("Expected config does not match the result")
		}
	})
}

func Test_Update(t *testing.T) {
	newData := TestConfig{Name: "new_data", Version: 456}

	t.Run("Check if config is updated", func(t *testing.T) {
		t.Cleanup(cleanUp)

		c, err := setUp(defaultConfig, "", testString, []string{})
		if err != nil {
			t.Errorf("Error while setting up test: %v", err)
			t.FailNow()
		}

		c.Update(newData)

		want := newData
		got := *c.GetCfg()
		if !reflect.DeepEqual(want, got) {
			t.Error("Expected config does not match the result")
		}
	})

	t.Run("Check if subscribers are being notified", func(t *testing.T) {
		t.Cleanup(cleanUp)

		subscribers := [5]string{"test1", "test2", "test3"}
		c, err := setUp(defaultConfig, "", testString, subscribers[:])
		if err != nil {
			t.Errorf("Error while setting up test: %v", err)
			t.FailNow()
		}

		c.Update(newData)

		if len(c.subscribers["test1"]) != 1 || len(c.subscribers["test2"]) != 1 || len(c.subscribers["test3"]) != 1 {
			t.Error("Subscribers not being notified")
		}
	})

	t.Run("Check channel read", func(t *testing.T) {
		t.Cleanup(cleanUp)

		subscribers := [1]string{"test1"}
		c, err := setUp(defaultConfig, "", testString, subscribers[:])
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

	t.Run("Check if channels not being overloaded", func(t *testing.T) {
		t.Cleanup(cleanUp)

		subscribers := [1]string{"test1"}
		c, err := setUp(defaultConfig, "", testString, subscribers[:])
		if err != nil {
			t.Errorf("Error while setting up test: %v", err)
			t.FailNow()
		}

		c.Update(newData)
		c.Update(newData)
		c.Update(newData)

		if len(c.subscribers["test1"]) != 1 {
			t.Error("Subscribers not being notified")
		}
	})
}
