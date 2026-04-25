package config

import (
	"log"
	"os"
	"strconv"

	"github.com/BurntSushi/toml"
)

type MainConfig struct {
	Port    int    `toml:"port"`
	AppName string `toml:"appName"`
	Host    string `toml:"host"`
}

type EmailConfig struct {
	Authcode string `toml:"authcode"`
	Email    string `toml:"email" `
}

type RedisConfig struct {
	RedisPort     int    `toml:"port"`
	RedisDb       int    `toml:"db"`
	RedisHost     string `toml:"host"`
	RedisPassword string `toml:"password"`
}

type MysqlConfig struct {
	MysqlPort         int    `toml:"port"`
	MysqlHost         string `toml:"host"`
	MysqlUser         string `toml:"user"`
	MysqlPassword     string `toml:"password"`
	MysqlDatabaseName string `toml:"databaseName"`
	MysqlCharset      string `toml:"charset"`
}

type JwtConfig struct {
	ExpireDuration int    `toml:"expire_duration"`
	Issuer         string `toml:"issuer"`
	Subject        string `toml:"subject"`
	Key            string `toml:"key"`
}

type Rabbitmq struct {
	RabbitmqPort     int    `toml:"port"`
	RabbitmqHost     string `toml:"host"`
	RabbitmqUsername string `toml:"username"`
	RabbitmqPassword string `toml:"password"`
	RabbitmqVhost    string `toml:"vhost"`
}

type RagModelConfig struct {
	RagEmbeddingModel string `toml:"embeddingModel"`
	RagChatModelName  string `toml:"chatModelName"`
	RagDocDir         string `toml:"docDir"`
	RagBaseUrl        string `toml:"baseUrl"`
	RagDimension      int    `toml:"dimension"`
}

type VoiceServiceConfig struct {
	VoiceServiceApiKey    string `toml:"voiceServiceApiKey"`
	VoiceServiceSecretKey string `toml:"voiceServiceSecretKey"`
}

type OCRConfig struct {
	APIURL        string `toml:"apiUrl"`
	Token         string `toml:"token"`
	TimeoutSecond int    `toml:"timeoutSecond"`
}

type Config struct {
	EmailConfig        `toml:"emailConfig"`
	RedisConfig        `toml:"redisConfig"`
	MysqlConfig        `toml:"mysqlConfig"`
	JwtConfig          `toml:"jwtConfig"`
	MainConfig         `toml:"mainConfig"`
	Rabbitmq           `toml:"rabbitmqConfig"`
	RagModelConfig     `toml:"ragModelConfig"`
	VoiceServiceConfig `toml:"voiceServiceConfig"`
	OCRConfig          `toml:"ocrConfig"`
}

type RedisKeyConfig struct {
	CaptchaPrefix   string
	IndexName       string
	IndexNamePrefix string
}

var DefaultRedisKeyConfig = RedisKeyConfig{
	CaptchaPrefix:   "captcha:%s",
	IndexName:       "rag_docs:%s:idx",
	IndexNamePrefix: "rag_docs:%s:",
}

var config *Config

// InitConfig 初始化项目配置
func InitConfig() error {
	// 设置配置文件路径（相对于 main.go 所在的目录）
	if _, err := toml.DecodeFile("config/config.toml", config); err != nil {
		log.Fatal(err.Error())
		return err
	}
	applyEnvOverrides(config)
	return nil
}

func applyEnvOverrides(c *Config) {
	if value := os.Getenv("MYSQL_ROOT_PASSWORD"); value != "" {
		c.MysqlPassword = value
	}
	if value := os.Getenv("MYSQL_DATABASE"); value != "" {
		c.MysqlDatabaseName = value
	}
	if value := os.Getenv("RABBITMQ_DEFAULT_USER"); value != "" {
		c.RabbitmqUsername = value
	}
	if value := os.Getenv("RABBITMQ_DEFAULT_PASS"); value != "" {
		c.RabbitmqPassword = value
	}
	if value := os.Getenv("EMAIL_AUTH_CODE"); value != "" {
		c.Authcode = value
	}
	if value := os.Getenv("EMAIL_ADDRESS"); value != "" {
		c.Email = value
	}
	if value := os.Getenv("OCR_API_URL"); value != "" {
		c.APIURL = value
	}
	if value := os.Getenv("OCR_API_TOKEN"); value != "" {
		c.Token = value
	}
	if value := os.Getenv("OCR_TIMEOUT_SECONDS"); value != "" {
		if timeout, err := strconv.Atoi(value); err == nil {
			c.TimeoutSecond = timeout
		}
	}
}

func GetConfig() *Config {
	if config == nil {
		config = new(Config)
		_ = InitConfig()
	}
	return config
}
