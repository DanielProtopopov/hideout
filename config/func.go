package config

import (
	"fmt"
	"github.com/shopspring/decimal"
	"os"
	"strconv"
	"time"
)

// GetEnv Helper function to read an environment or return a default value
func GetEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultVal
}

// GetEnvAsInt Helper function to read an environment variable into integer or return a default value
func GetEnvAsInt(name string, defaultVal int) int {
	valueStr := GetEnv(name, "")
	if value, errAtoi := strconv.Atoi(valueStr); errAtoi == nil {
		return value
	}

	return defaultVal
}

// GetEnvAsDecimal Helper function to read an environment variable into integer or return a default value
func GetEnvAsDecimal(name string, defaultVal decimal.Decimal) decimal.Decimal {
	valueStr := GetEnv(name, "")
	if value, errDecimal := decimal.NewFromString(valueStr); errDecimal == nil {
		return value
	}

	return defaultVal
}

// GetEnvAsUInt Helper function to read an environment variable into unsigned integer or return a default value
func GetEnvAsUInt(name string, defaultVal uint) uint {
	valueStr := GetEnv(name, "")
	if value, errAtoi := strconv.Atoi(valueStr); errAtoi == nil {
		return uint(value)
	}

	return defaultVal
}

// GetEnvAsBool Helper to read an environment variable into a bool or return default value
func GetEnvAsBool(name string, defaultVal bool) bool {
	valStr := GetEnv(name, "")
	if val, errBool := strconv.ParseBool(valStr); errBool == nil {
		return val
	}

	return defaultVal
}

// GetEnvAsDuration Helper function to read an environment or return a default value
func GetEnvAsDuration(key string, defaultVal time.Duration) time.Duration {
	valueStr := GetEnv(key, "")
	if value, errAtoi := strconv.Atoi(valueStr); errAtoi == nil {
		return time.Duration(value)
	}

	return defaultVal
}

func (dc *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("%s://%s:%s@%s:%d/%s?sslmode=%s",
		dc.Proto,
		dc.User,
		dc.Pass,
		dc.Host,
		dc.Port,
		dc.Name,
		dc.SSLMode,
	)
}
