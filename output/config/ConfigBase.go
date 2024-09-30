package config

type ConfigBase struct {
	Type string `yaml:"type"` // <-- The struct tag: "type" key in YAML loads into Type field
}
