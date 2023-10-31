package config

import (
	"encoding/json"
	"os"
)

func LoadFromJson(fileName string, config interface{}) error {
	if file, err := os.ReadFile(fileName); err != nil {
		return err
	} else if err := json.Unmarshal(file, &config); err != nil {
		return err
	} else {
		return nil
	}
}
