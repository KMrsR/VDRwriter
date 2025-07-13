package main

type Config struct {
	PoolingDelay   int `yaml:"pooling_delay"`
	WriteAllDelay  int `yaml:"write_all_delay"`
	WatchdogPeriod int `yaml:"watchdog_period"`
	// FLAGsStartReg  int `yaml:"FLAGs_start_reg"`
	// FLAGsCntReg    int `yaml:"FLAGs_count_reg"`
	MapSize int `yaml:"map_size"`
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
