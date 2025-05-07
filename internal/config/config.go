package config

type AppConfig struct {
	Redis    RedisConfig    `yaml:"redis"`
	Postgres PostgresConfig `yaml:"postgres"`
	Tasks    []TaskConfig   `yaml:"tasks"`
}

type RedisConfig struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type PostgresConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
	SSLMode  string `yaml:"sslmode"`
}

type TaskConfig struct {
	Name       string            `yaml:"name"`                 // Logical task name
	RedisKey   string            `yaml:"redis_key"`            // Redis key to read from
	RedisType  string            `yaml:"redis_type"`           // list, stream, sorted_set
	DBTable    string            `yaml:"db_table"`             // Target Postgres table
	FieldMap   map[string]string `yaml:"field_map"`            // Redis field -> DB column
	Schedule   string            `yaml:"schedule"`             // Cron or @every interval
	Checkpoint *CheckpointConfig `yaml:"checkpoint,omitempty"` // Optional
}

type CheckpointConfig struct {
	Enabled      bool   `yaml:"enabled"`        // Enable tracking
	Field        string `yaml:"field"`          // Field to track (e.g. stream ID, timestamp)
	LastValueKey string `yaml:"last_value_key"` // Redis key to store the last checkpoint value
}
