package main

type Config struct {
	PoolingDelay     int    `yaml:"pooling_delay"`
	WriteAllDelay    int    `yaml:"write_all_delay"`
	WatchdogPeriod   int    `yaml:"watchdog_period"`
	MapSize          int    `yaml:"map_size"`
	Role             string `yaml:"role"`               // "A" или "B"
	ActiveControlReg uint16 `yaml:"active_control_reg"` // например 1450
}

type Tag struct {
	Name     string `yaml:"name"`
	Register uint16 `yaml:"reg"`
	Type     string `yaml:"type"`
	Length   uint16 `yaml:"length,omitempty"` // для многорегистровых (опционально)
}

type TagConfig struct {
	Tags []Tag `yaml:"tags"`
}
