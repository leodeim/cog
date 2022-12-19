# goconfig

Config tool for Go applications that require configuration changes on-the-fly.

```bash
go get github.com/leonidasdeim/goconfig
```

## How to use

Write config structure of your app. Actual config should be placed in root folder named `app_config.default.json`
Example:

```go
type ConfigType struct {
 AppName   string `json:"name"`
 Version   string `json:"version"`
 Prefork   bool   `json:"prefork"`
}
```

If you have modules which needs to be notified on config change, create similar enum:

```go
type ConfigSubscriber int
const (
 FIRST_SUB ConfigSubscriber = iota
 SECOND_SUB
 NUMBER_OF_SUBS
)
```

Initialize and use config:

```go
config, err := goconfig.Init[ConfigType](int(NUMBER_OF_SUBS))

updatedConfig := config.Get() // access current config 
updatedConfig.AppName = "NewName"
config.Update(updatedConfig) // update current config on-the-fly
```

Implement waiting goroutine for config change on the fly in your modules:

```go
<-config.GetSubscriber(FIRST_SUB)
```
