package config

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/ethereum/go-ethereum/common"
	"github.com/kelseyhightower/envconfig"
)

const (
	CONFIG_FILE string = "config.toml"
)

var (
	GlobalConfigCallback ConfigCallback[GlobalConfig] = ConfigCallback[GlobalConfig]{}
)

type GlobalConfig interface {
	LoggerConfig() LoggerConfig
	ChainConfig() ChainConfig
}

type LoggerLevel string

type LoggerConfig struct {
	Level       string `toml:"level"` // valid values are: DEBUG, INFO, WARN, ERROR, DPANIC, PANIC, FATAL (zap)
	File        string `toml:"file"`
	MaxFileSize int    `toml:"max_file_size"` // In megabytes
	Console     bool   `toml:"console"`
}

type DBConfig struct {
	Host       string `toml:"host" envconfig:"DB_HOST"`
	Port       int    `toml:"port" envconfig:"DB_PORT"`
	Database   string `toml:"database" envconfig:"DB_DATABASE"`
	Username   string `toml:"username" envconfig:"DB_USERNAME"`
	Password   string `toml:"password" envconfig:"DB_PASSWORD"`
	LogQueries bool   `toml:"log_queries"`
}

type ChainConfig struct {
	NodeURL         string `toml:"node_url" envconfig:"CHAIN_NODE_URL"`
	ChainAddressHRP string `toml:"address_hrp" envconfig:"CHAIN_ADDRESS_HRP"`
	ChainID         int    `toml:"chain_id" envconfig:"CHAIN_ID"`
	EthRPCURL       string `toml:"eth_rpc_url" envconfig:"ETH_RPC_URL"`
	ApiKey          string `toml:"api_key" envconfig:"API_KEY"`
	// setting the private key in config file is deprecated, except in development and testing
	// use private_key_file instead
	PrivateKey     string `toml:"private_key" envconfig:"PRIVATE_KEY"`
	PrivateKeyFile string `toml:"private_key_file" envconfig:"PRIVATE_KEY_FILE"`
}

func (cfg ChainConfig) GetPrivateKey() (string, error) {
	if cfg.PrivateKeyFile == "" {
		log.Print("WARNING: using private_key is deprecated, use private_key_file instead")
		return cfg.PrivateKey, nil
	} else {
		content, err := os.ReadFile(cfg.PrivateKeyFile)
		if err != nil {
			return "", fmt.Errorf("error opening private key file: %w", err)
		}
		return strings.TrimSpace(string(content)), nil
	}
}

type EpochConfig struct {
	First int64 `toml:"first" envconfig:"EPOCH_FIRST"`
}

type ContractAddresses struct {
	Voting common.Address `toml:"voting" envconfig:"VOTING_CONTRACT_ADDRESS"`
}

func ParseConfigFile(cfg interface{}, fileName string, allowMissing bool) error {
	content, err := os.ReadFile(fileName)
	if err != nil {
		if allowMissing {
			return nil
		} else {
			return fmt.Errorf("error opening config file: %w", err)
		}
	}

	_, err = toml.Decode(string(content), cfg)
	if err != nil {
		return fmt.Errorf("error parsing config file: %w", err)
	}
	return nil
}

func ReadEnv(cfg interface{}) error {
	err := envconfig.Process("", cfg)
	if err != nil {
		return fmt.Errorf("error reading env config: %w", err)
	}
	return nil
}
