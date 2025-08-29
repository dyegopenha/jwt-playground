package env

import (
	"bytes"
	_ "embed"
	"fmt"
	"log"
	"os"
	"time"

	root "github.com/dyegopenha/jwt-playground"
	"github.com/dyegopenha/jwt-playground/internal/pkg/validator"
	"github.com/spf13/viper"
)

const defaultEnvFileName = ".env"

type Environment string

const (
	EnvironmentDevelopment Environment = "development"
	EnvironmentProduction  Environment = "production"
	EnvironmentStaging     Environment = "staging"
	EnvironmentTest        Environment = "test"
)

type Env struct {
	v validator.Validator

	Environment      Environment   `mapstructure:"ENVIRONMENT"        validate:"required,oneof=development production staging test"`
	Port             string        `mapstructure:"PORT"`
	RedisDatabaseURL string        `mapstructure:"REDIS_DATABASE_URL" validate:"required"`
	HMACKey          string        `mapstructure:"HMAC_KEY"           validate:"required"`
	AccessTokenTTL   time.Duration `mapstructure:"ACCESS_TOKEN_TTL"   validate:"required"`
	RefreshTokenTTL  time.Duration `mapstructure:"REFRESH_TOKEN_TTL"  validate:"required"`
}

func NewEnv(v validator.Validator) *Env {
	e := &Env{
		v: v,
	}

	if err := e.loadEnv(); err != nil {
		log.Fatalf("failed to load environment variables: %v", err)
	}

	return e
}

type envVariables struct {
	Environment        Environment `mapstructure:"ENVIRONMENT"        validate:"required,oneof=development production staging test"`
	Port               string      `mapstructure:"PORT"`
	RedisDatabaseURL   string      `mapstructure:"REDIS_DATABASE_URL" validate:"required"`
	HMACKey            string      `mapstructure:"HMAC_KEY"           validate:"required"`
	AccessTokenTTLStr  string      `mapstructure:"ACCESS_TOKEN_TTL"   validate:"required"`
	RefreshTokenTTLStr string      `mapstructure:"REFRESH_TOKEN_TTL"  validate:"required"`
}

func (e *Env) loadEnv() error {
	envFile, err := e.getEnvFile()
	if err != nil {
		return fmt.Errorf("failed to load environment variables: %w", err)
	}

	viper.SetConfigType("env")

	if err := viper.ReadConfig(bytes.NewBuffer(envFile)); err != nil {
		return fmt.Errorf("failed to read environment variables: %w", err)
	}

	viper.AutomaticEnv()

	envVariables := envVariables{}
	if err := viper.Unmarshal(&envVariables); err != nil {
		return fmt.Errorf("failed to unmarshal environment variables: %w", err)
	}

	if err := e.parseEnvVariables(envVariables); err != nil {
		return fmt.Errorf("failed to parse environment variables: %w", err)
	}

	if err := e.validate(); err != nil {
		return fmt.Errorf("failed to validate environment variables: %w", err)
	}

	return nil
}

func (e *Env) parseEnvVariables(envVariables envVariables) error {
	e.Environment = envVariables.Environment
	e.Port = envVariables.Port
	e.RedisDatabaseURL = envVariables.RedisDatabaseURL
	e.HMACKey = envVariables.HMACKey

	accessTokenTTL, err := time.ParseDuration(envVariables.AccessTokenTTLStr)
	if err != nil {
		return fmt.Errorf("failed to parse access token ttl: %w", err)
	}
	e.AccessTokenTTL = accessTokenTTL

	refreshTokenTTL, err := time.ParseDuration(envVariables.RefreshTokenTTLStr)
	if err != nil {
		return fmt.Errorf("failed to parse refresh token ttl: %w", err)
	}
	e.RefreshTokenTTL = refreshTokenTTL

	return nil
}

func (e *Env) getEnvFile() (envFile []byte, err error) {
	environment := os.Getenv("ENVIRONMENT")

	if environment != "" {
		envFileName := fmt.Sprintf("%s.%s", defaultEnvFileName, environment)
		envFile, err = root.Env.ReadFile(envFileName)
		if err == nil {
			return envFile, nil
		}
	}

	envFile, err = root.Env.ReadFile(defaultEnvFileName)
	if err != nil {
		return nil, fmt.Errorf("failed to read environment variables: %w", err)
	}

	return envFile, nil
}

func (e *Env) validate() error {
	if err := e.v.Validate(e); err != nil {
		return err
	}
	if e.Environment == "" {
		e.Environment = EnvironmentDevelopment
	}
	if e.Port == "" {
		e.Port = "8080"
	}
	return nil
}
