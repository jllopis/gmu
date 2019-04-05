package conf

import (
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// type Config struct {
// 	NatsConf *messaging.NatsConfig `mapstructure:"nats_conf"`
// 	LogConf  LoggingConfig         `mapstructure:"log_conf"`
// }

// LoadConfig loads the config from a file if specified, otherwise from the environment
func LoadConfig(cmd *cobra.Command) error {
	// func LoadConfig(cmd *cobra.Command) (*Config, error) {
	// viper.SetConfigType("json")

	err := viper.BindPFlags(cmd.Flags())
	if err != nil {
		return err
	}

	viper.SetEnvPrefix("gmu")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	if configFile, _ := cmd.Flags().GetString("config"); configFile != "" {
		viper.SetConfigFile(configFile)
	} else {
		viper.SetConfigName("gmuconfig")
		viper.AddConfigPath("./")
		viper.AddConfigPath("$HOME/.config/gmu/")
		viper.AddConfigPath("$HOME/.gmu/")
	}

	if err := viper.ReadInConfig(); err != nil && !os.IsNotExist(err) {
		return err
	}

	// config := &Config{}

	// if err := viper.Unmarshal(config); err != nil {
	// 	return nil, err
	// }

	// return config, nil
	return nil
}
