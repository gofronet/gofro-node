package config

import "os"

func LoadXrayConfigFromFile(confPath string) (string, error) {
	conf, err := os.ReadFile(confPath)
	if err != nil {
		return "", err
	}
	return string(conf), err
}

func WriteConfigToFile(confPath, newConf string) error {
	return os.WriteFile(confPath, []byte(newConf), 0644)
}
