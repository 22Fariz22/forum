package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// App config struct
type Config struct {
	Server     ServerConfig
	Middleware MiddlewareConfig
	Storage    StorageConfig
	Postgres   PostgresConfig
	Logger     Logger
}

// Server config struct
type ServerConfig struct {
	AppVersion        string
	BaseUrl           string
	Port              string
	Mode              string
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	CtxDefaultTimeout time.Duration
	MaxHeaderBytes    int
	CtxTimeout        time.Duration
	Debug             bool
}

// Middleware config struct
type MiddlewareConfig struct {
	MiddlewareStackSize         int
	MiddlewareDisablePrintStack bool
	MiddlewareDisableStackAll   bool
	MiddlewareLevel             int
	MiddlewarebodyLimit         string
}

// Logger config
type Logger struct {
	Development       bool
	DisableCaller     bool
	DisableStacktrace bool
	Encoding          string
	Level             string
}

// Storage type of storage
type StorageConfig struct {
	StorageType string
}

// Postgresql config
type PostgresConfig struct {
	PostgresqlHost     string
	PostgresqlPort     string
	PostgresqlUser     string
	PostgresqlPassword string
	PostgresqlDbname   string
	PostgresqlSSLMode  bool
	PgDriver           string
}

// LoadConfig reads environment variables into a Config struct
func LoadConfig() (*Config, error) {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found. Falling back to environment variables.")
	}

	return &Config{
		Server: ServerConfig{
			AppVersion:        getEnv("APP_VERSION", "1.0.0"),
			BaseUrl:           getEnv("SERVER_BASE_URL", "localhost"),
			Port:              getEnv("SERVER_PORT", "8080"),
			Mode:              getEnv("MODE", "Development"),
			ReadTimeout:       getEnvAsDuration("READ_TIMEOUT", 10*time.Second),
			WriteTimeout:      getEnvAsDuration("WRITE_TIMEOUT", 10*time.Second),
			CtxDefaultTimeout: getEnvAsDuration("CTX_DEFAULT_TIMEOUT", 12*time.Second),
			MaxHeaderBytes:    getEnvAsInt("MAX_HEADER_BYTES", 1<<20),
			CtxTimeout:        getEnvAsDuration("CTX_TIMEOUT", 5*time.Second),
			Debug:             getEnvAsBool("DEBUG", false),
		},
		Middleware: MiddlewareConfig{
			MiddlewareStackSize:         getEnvAsInt("MIDDLEWARE_STACK_SIZE", 1024), // Default to 1024 (1 << 10)
			MiddlewareDisablePrintStack: getEnvAsBool("MIDDLEWARE_DISABLE_PRINT_STACK", true),
			MiddlewareDisableStackAll:   getEnvAsBool("MIDDLEWARE_DISABLE_STACK_ALL", true),
			MiddlewareLevel:             getEnvAsInt("MIDDLEWARE_LEVEL", 5),
			MiddlewarebodyLimit:         getEnv("MIDDLEWARE_BODY_LIMIT", "1000M"),
		},
		Logger: Logger{
			Development:       getEnvAsBool("LOGGER_DEVELOPMENT", true),
			DisableCaller:     getEnvAsBool("LOGGER_DISABLE_CALLER", false),
			DisableStacktrace: getEnvAsBool("LOGGER_DISABLE_STACKTRACE", false),
			Encoding:          getEnv("LOGGER_ENCODING", "console"),
			Level:             getEnv("LOGGER_LEVEL", "info"),
		},
		Storage: StorageConfig{
			StorageType: getEnv("STORAGE_TYPE", "inmemory"),
		},
		Postgres: PostgresConfig{
			PostgresqlHost:     getEnv("POSTGRES_HOST", "localhost"),
			PostgresqlPort:     getEnv("POSTGRES_PORT", "5432"),
			PostgresqlUser:     getEnv("POSTGRES_USER", "postgres"),
			PostgresqlPassword: getEnv("POSTGRES_PASSWORD", "postgres"),
			PostgresqlDbname:   getEnv("POSTGRES_DBNAME", "forum_db"),
			PostgresqlSSLMode:  getEnvAsBool("POSTGRES_SSLMODE", false),
			PgDriver:           getEnv("POSTGRES_DRIVER", "pgx"),
		},
	}, nil
}

// Helper functions to parse environment variables
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	valStr := getEnv(key, "")
	if val, err := strconv.ParseBool(valStr); err == nil {
		return val
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	valStr := getEnv(key, "")
	if val, err := time.ParseDuration(valStr); err == nil {
		return val
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valStr := getEnv(key, "")
	if val, err := strconv.Atoi(valStr); err == nil {
		return val
	}
	return defaultValue
}
