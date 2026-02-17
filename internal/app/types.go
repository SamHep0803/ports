package app

type Config struct {
	Profiles []Profile `yaml:"profiles"`
}

type Profile struct {
	Name     string    `yaml:"name"`
	Host     string    `yaml:"host"`
	User     string    `yaml:"user"`
	KeyPath  string    `yaml:"keyPath"`
	Forwards []Forward `yaml:"forwards"`
}

type Forward struct {
	Bind       string `yaml:"bind"`
	LocalPort  int    `yaml:"localPort"`
	RemoteHost string `yaml:"remoteHost"`
	RemotePort int    `yaml:"remotePort"`
}
