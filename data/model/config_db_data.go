package model

type DatabaseConfig struct {
	DBHost     string `json:"DB_HOST"`
	DBUser     string `json:"DB_USER"`
	DBPassword string `json:"DB_PASSWORD"`
	DBName     string `json:"DB_NAME"`
	DBPort     string `json:"DB_PORT"`
}
