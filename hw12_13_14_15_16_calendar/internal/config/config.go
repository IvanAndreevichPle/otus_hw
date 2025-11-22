// Package config содержит структуры и функции для работы с конфигурацией приложения.
package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config описывает структуру основного конфига приложения.
type Config struct {
	Logger  LoggerConf  `yaml:"logger"`  // параметры логирования
	Storage StorageConf `yaml:"storage"` // параметры хранилища
	Server  ServerConf  `yaml:"server"`  // параметры HTTP-сервера
	DB      DBConf      `yaml:"db"`      // параметры БД
}

// LoggerConf содержит параметры логирования.
type LoggerConf struct {
	Level string `yaml:"level"` // error, warn, info, debug
}

// StorageConf описывает тип используемого хранилища.
type StorageConf struct {
	Type string `yaml:"type"` // memory или sql
}

// ServerConf содержит параметры HTTP-сервера.
type ServerConf struct {
	Host string `yaml:"host"` // адрес
	Port int    `yaml:"port"` // порт
}

// DBConf содержит параметры подключения к базе данных.
type DBConf struct {
	Host     string `yaml:"host"`     // адрес БД
	Port     int    `yaml:"port"`     // порт БД
	User     string `yaml:"user"`     // пользователь
	Password string `yaml:"password"` // пароль
	DBName   string `yaml:"dbname"`   // имя базы
}

// NewConfigFromFile читает и парсит YAML-конфиг из файла.
func NewConfigFromFile(path string) (Config, error) {
	var cfg Config
	f, err := os.Open(path)
	if err != nil {
		return cfg, err
	}
	defer func() { _ = f.Close() }()
	dec := yaml.NewDecoder(f)
	if err := dec.Decode(&cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}
