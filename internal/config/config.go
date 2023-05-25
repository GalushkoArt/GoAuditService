package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"log"
	"os"
	"path/filepath"
)

type Config struct {
	Mongo struct {
		URI      string `yaml:"uri" env:"MONGO_URI" env-default:"mongodb://localhost:27017"`
		Username string `yaml:"username" env:"MONGO_USERNAME" env-default:"audit-admin"`
		Password string `yaml:"password" env:"MONGO_PASSWORD"`
		Database string `yaml:"database" env:"MONGO_DATABASE" env-default:"audit"`
	} `yaml:"mongo"`
	MQ   MqConf `yaml:"mq"`
	GRPC struct {
		Enabled bool `yaml:"enabled" env:"GRPC_ENABLED" env-default:"false"`
		Port    int  `yaml:"port" env:"GRPC_PORT" env-default:"50051"`
	} `yaml:"grpc_server"`
	Logs struct {
		Level string `yaml:"level" env:"LOGS_LEVEL" env-default:"INFO"`
		Path  string `yaml:"path" env:"LOGS_PATH" env-default:"logs.txt"`
	} `yaml:"logs"`
}

type MqConf struct {
	Enabled     bool   `yaml:"enabled" env:"MQ_ENABLED" env-default:"false"`
	Concurrency int    `yaml:"concurrency" env:"MQ_CONCURRENCY" env-default:"10"`
	User        string `yaml:"username" env:"MQ_USERNAME" env-default:"rmuser"`
	Password    string `yaml:"password" env:"MQ_PASSWORD"`
	Host        string `yaml:"host" env:"MQ_HOST" env-default:"localhosta"`
	Port        int    `yaml:"port" env:"MQ_PORT" env-default:"5672"`
	QueueName   string `yaml:"queue_name" env:"MQ_QUEUE_NAME" env-default:"audit"`
}

var Conf Config

func Init() {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal("Couldn't open current working directory!", err)
	}
	err = cleanenv.ReadConfig(filepath.Join(wd, "config/config.yaml"), &Conf)
	if err != nil {
		log.Fatal("Error on reading config!", err)
	}
}
