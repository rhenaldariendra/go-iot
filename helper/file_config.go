package helper

import (
	"Websocket_Service/data/model"
	"encoding/json"
	"log"
	"os"
)

func ReadConfigBaseServer() model.BaseConfig {
	fileContent, err := os.ReadFile("config/base_config.json")
	if err != nil {
		log.Panic(err)
	}

	var config model.BaseConfig
	err = json.Unmarshal(fileContent, &config)
	if err != nil {
		log.Panic(err)
		return model.BaseConfig{}
	}

	return config
}

func ReadConfigDB() model.DatabaseConfig {
	fileContent, err := os.ReadFile("config/db_config.json")
	if err != nil {
		log.Panic(err)
	}

	var config model.DatabaseConfig
	err = json.Unmarshal(fileContent, &config)
	if err != nil {
		log.Panic(err)
		return model.DatabaseConfig{}
	}

	return config
}
