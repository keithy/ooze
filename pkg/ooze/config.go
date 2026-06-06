package ooze

const Version = "0.1.0"

type Config struct {
	Pool       string `json:"pool"`
	Dataset    string `json:"dataset"`
	TargetHost string `json:"target_host"`
}

func DefaultConfig() *Config {
	return &Config{
		Pool:    "tank",
		Dataset: "",
	}
}
