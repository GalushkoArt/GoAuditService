package config

import (
	"github.com/galushkoart/go-audit-service/internal/utils"
	"github.com/ilyakaznacheev/cleanenv"
	"log"
	"os"
	"path/filepath"
)

type TestConfig struct {
	Mongo struct {
		Version string `yaml:"version" env:"MONGO_VERSION"`
	} `yaml:"mongo"`
	GRPC struct {
		Port int `yaml:"port" env:"GRPC_PORT" env-default:"50051"`
	} `yaml:"grpc_server"`
	MQ struct {
		Implementation string `yaml:"implementation" env:"MQ_IMAGE" env-default:"rabbitmq"`
		Image          utils.Image
		Version        string `yaml:"version" env:"MQ_VERSION"`
	} `yaml:"mq"`
}

var TestConf TestConfig

func InitTest() error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	err = cleanenv.ReadConfig(filepath.Join(wd, "config/test_config.yaml"), &TestConf)
	if err != nil {
		return err
	}
	TestConf.MQ.Image = findMqImage(TestConf.MQ.Implementation)
	return nil
}

func findMqImage(implementation string) utils.Image {
	for _, image := range utils.Images {
		if image.Name == implementation {
			return image
		}
	}
	log.Fatalf("Couldn't find image for implementation: %s in config/test_config.yaml\n", implementation)
	return utils.Image{}
}
