<p align="center">
 <img src="assets/goconfig.png" width="450">
</p>

<div align="center">

  <a href="">![Tests](https://github.com/leonidasdeim/goconfig/actions/workflows/go.yml/badge.svg)</a>
  <a href="">![Code Scanning](https://github.com/leonidasdeim/goconfig/actions/workflows/codeql.yml/badge.svg)</a>
  <a href="">![Release](https://badgen.net/github/release/leonidasdeim/goconfig/)</a>
  <a href="">![Releases](https://badgen.net/github/releases/leonidasdeim/goconfig)</a>
  <a href="">![Contributors](https://badgen.net/github/contributors/leonidasdeim/goconfig)</a>
  
</div>

# goconfig

Config tool for Go applications that require configuration changes on-the-fly.

```bash
go get github.com/leonidasdeim/goconfig
```

## Overview

Currently **goconfig** supports **JSON** (default) and **YAML** configuration files with built-in `pkg/handler/filehandler.go`. How to use built-in file handlers find [here](#file-handler-type). You can always write your own handler which would implement `ConfigHandler` interface.

Default config with initial configuration information should be placed in root folder named `<name>.default.<type>`. Name and type of the file could be changed using [custom parameters](#custom-parameters). **Goconfig** also let to you set up default values for entries in configuration with `default:"some_value"` tag. Right now, only *bool*, *int* and *string* is supported.

**Goconfig** uses [validator](https://github.com/go-playground/validator) library for validating loaded configuration. For example you can specify required configuration items with `validate:"required"` tag.

## Getting started

Write config structure of your app. Example of config structure:

```go
type ConfigType struct {
    Version   string `validate:"required"`
    Address   string `validate:"required,ip"`
    Prefork   bool `default:"false"`
}
```

Import main library:

```go
import "github.com/leonidasdeim/goconfig"
```

Initialize and use config:

```go
// creates default goconfig instance with JSON file handler
c, _ := goconfig.Init[ConfigType]()

// access current configuration attributes
cfg := c.GetCfg()

// make some changes to 'cfg' and update current configuration
c.UpdateConfig(cfg)
```

For more examples check out `examples/` folder.

## Change notifications

If you have modules which needs to be notified on config change, add a listener/subscriber:

```go
c.AddSubscriber("name_of_subscriber")
```

Implement waiting goroutine for config change on the fly in your modules:

```go
for {
    <-c.GetSubscriber("name_of_subscriber")
    reconfigureModule()
}
```

You can remove subscriber by given name on the fly as well:

```go
c.RemoveSubscriber("name_of_subscriber")
```

## File handler type

By default **goconfig** initializes with JSON file handler. You can specify type by creating handler instance and providing it during initialization.

Import built-in filehandler
```go
import (
	"github.com/leonidasdeim/goconfig"
	fh "github.com/leonidasdeim/goconfig/pkg/filehandler"
)
```

```go
h, _ := fh.New(fh.WithType(fh.YAML))
c, _ := goconfig.Init[ConfigType](h)
```

## Custom parameters

Handlers also support optional parameters with high order functions.
You can specify custom path, name and file handler (currently JSON, YAML and TOML are supported by default)

```go
h, _ := fh.New(fh.WithPath("./dir"), fh.WithName("name"), fh.WithType(fh.JSON))
c, _ := goconfig.Init[ConfigType](h)
```
