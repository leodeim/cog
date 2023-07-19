<p align="center">
 <img src="assets/cog.png" width="450">
</p>

<div align="center">

  <a href="">![Tests](https://github.com/leonidasdeim/cog/actions/workflows/go.yml/badge.svg)</a>
  <a href="">![Code Scanning](https://github.com/leonidasdeim/cog/actions/workflows/codeql.yml/badge.svg)</a>
  <a href="">![Release](https://badgen.net/github/release/leonidasdeim/cog/)</a>
  <a href="">![Releases](https://badgen.net/github/releases/leonidasdeim/cog)</a>
  <a href="">![Contributors](https://badgen.net/github/contributors/leonidasdeim/cog)</a>
  
</div>

# cog

Config tool for Go applications that require configuration changes on-the-fly.

```bash
go get github.com/leonidasdeim/cog
```

## Overview

Currently **cog** supports **JSON**, **YAML** and **TOML** configuration files with built-in `pkg/handler/filehandler.go`. By default it dynamically detects configuration file type. If you want to specify file type, [here](#file-handler-type) you can find how to use built-in file handlers. You can always write your own handler which would implement `ConfigHandler` interface.

Default config with initial configuration information should be placed in root folder named `<name>.default.<type>`. Name and type of the file could be changed using [custom parameters](#custom-parameters). **cog** also let to you set up default values for entries in configuration with `default:"some_value"` tag. Right now, only *bool*, *int* and *string* is supported.

It is possible to load config fields values from **environment variables** using `env:"ENV_VAR_NAME"` tag. With this tag **cog** will take env. variable value and use it if field value not provided in the config file.

**cog** uses [validator](https://github.com/go-playground/validator) library for validating loaded configuration. For example you can specify required configuration items with `validate:"required"` tag.

## Getting started

Write config structure of your app. Example of config structure with different tags:

```go
type ConfigType struct {
    // simple field, will be empty string if not provided
    Name      string 

    // required: will fail if not provided
    Version   string `validate:"required"` 
    
    // tries to load from env. variable "SERVER_IP_ADDRESS" if not provided in the config file
    Address   string `validate:"required,ip" env:"SERVER_IP_ADDRESS"` 
    
    // sets default value "8080" if field not provided in the config file
    Port      string   `default:"8080"` 
}
```

Import main library:

```go
import "github.com/leonidasdeim/cog"
```

Initialize and use config:

```go
// creates default cog instance with JSON file handler
c, _ := cog.Init[ConfigType]()

// access current configuration attributes
cfg := c.GetCfg()

// make some changes to 'cfg' and update current configuration
c.UpdateConfig(cfg)
```

For more examples check out `examples/` folder.

## Change notifications

### Callback

Register a callback function, which will be called on config change:
```go
c.AddCallback(func(cfg ConfigType) {
    // handle config update
})
```

### Bound

You can register another type of callback - **bound**. It will be called on config change just like regular callback, but it have different signature: it can return an error.
In case bound callback returns an error - whole config update is being rolled back.
Example:
```go
c.AddBound(func(cfg ConfigType) error {
    if err := tryConfigUpdate(cfg); err != nil {
        return err
    }
    return nil
})
```

### Subscription

One more option to be notified on config change is through channels.
To add a listener/subscriber:

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

By default **cog** initializes with dynamic file handler. You can specify type (JSON, YAML or TOML) by creating handler instance and providing it during initialization.

Import built-in filehandler
```go
import (
	"github.com/leonidasdeim/cog"
	fh "github.com/leonidasdeim/cog/pkg/filehandler"
)
```

```go
h, _ := fh.New(fh.WithType(fh.YAML))
c, _ := cog.Init[ConfigType](h)
```

## Custom parameters

Handlers also support optional parameters with high order functions.
You can specify custom path, name and file handler.

```go
h, _ := fh.New(fh.WithPath("./dir"), fh.WithName("name"), fh.WithType(fh.JSON))
c, _ := cog.Init[ConfigType](h)
```
