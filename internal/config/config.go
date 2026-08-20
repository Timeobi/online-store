package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Config хранит все настройки приложения, собранные из переменных окружения.
type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string
	ServerPort string
	JWTSecret  string // добавили
}

// Load читает .env файл (если есть) и переменные окружения,
// затем собирает их в структуру Config.
func Load() (*Config, error) {
	// Пытаемся загрузить .env — если файла нет, это не критическая ошибка
	// (например, в проде переменные обычно задаются напрямую системой)
	if err := godotenv.Load(); err != nil {
		fmt.Println("предупреждение: .env файл не найден, используются системные переменные окружения")
	}

	cfg := &Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBName:     getEnv("DB_NAME", "ecommerce_dev"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),
		ServerPort: getEnv("SERVER_PORT", "8080"),
		JWTSecret:  getEnv("JWT_SECRET", ""), // добавили
	}

	return cfg, nil
}

// DSN формирует connection string для подключения к PostgreSQL.
// DSN = Data Source Name, стандартный термин для строки подключения к БД.
func (c *Config) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName, c.DBSSLMode,
	)
}

// getEnv — вспомогательная функция: берёт переменную окружения,
// а если её нет — возвращает значение по умолчанию.
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
