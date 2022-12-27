<p align="center">
 <img src="img/goconfig.png" width="450">
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

## How to use

Initial config that will store configuration information should be placed in root folder named `app.default.json`. Write config structure of your app.

`goconfig` uses [validator](https://github.com/go-playground/validator) library for validating config struct. For example you can specify required configuration items with `validate:"required"` tag.

`goconfig` also let to you set up default values for entries in configuration with `default:"some value"` tag. Right now, only bool, int and string is supported.

Example of config structure:

```go
type ConfigType struct {
    Version   string `validate:"required"`
    Address   string `validate:"required,ip"`
    Prefork   bool `default:"false"`
}
```

Initialize and use config:

```go
config := config.Init[ConfigType]()

// access current configuration attributes
cfg := config.GetCfg()

// update current configuration
config.UpdateConfig(cfg)
```

If you have modules which needs to be notified on config change, add a listener/subscriber:

```go
c.AddSubscriber("name_of_subscriber")
```

Implement waiting goroutine for config change on the fly in your modules:

```go
for {
    _ = <-config.GetSubscriber("name_of_subscriber")
    reconfigureModule()
}
```

You can remove subscriber by given name on the fly as well:

```go
c.RemoveSubscriber("name_of_subscriber")
```

Library also support optional parameters with high order functions:

```go
config := config.Init[ConfigType](WithPath("./configuration_dir"), WithName("configuration_name"))
```
