package config

import "github.com/spf13/viper"

type PostgresConfig struct {
	DBHost     string `mapstructure:"DB_HOST"`
	DBPort     string `mapstructure:"DB_PORT"`
	DBUser     string `mapstructure:"DB_USER"`
	DBPassword string `mapstructure:"DB_PASSWORD"`
	DBName     string `mapstructure:"DB_NAME"`
}

func LoadPostgresConfig(file_name string) (config *PostgresConfig, err error) {
	// Set default values
	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", "5432")
	viper.SetDefault("DB_USER", "postgres")
	viper.SetDefault("DB_PASSWORD", "postgres")
	viper.SetDefault("DB_NAME", "mini_asm")

	// Try to read from .env file first
	viper.SetConfigFile(file_name)
	err = viper.ReadInConfig()
	if err != nil {
		// If .env file not found, read from OS environment variables
		viper.AutomaticEnv()
	}

	config = &PostgresConfig{}
	err = viper.Unmarshal(config)
	if err != nil {
		return
	}

	return config, err
}
