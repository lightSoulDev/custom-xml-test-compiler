package xmlParser

// Config ...
type Config struct {
	AppData     string `toml:"app_data"`
	CommonPath  string
	ConfigsPath string
}

// NewConfig ...
func NewConfig() *Config {
	return &Config{
		AppData: "./appData",
	}
}
