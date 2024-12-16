package config

import (
	"os"
)

type Config struct {
   LLamaUrl string `json:"llama_url"`	
}

func NewConfig() (*Config, error) {
	llamaUrl := os.Getenv("LLAMA_URL")
	config := &Config{
		LLamaUrl: llamaUrl,
	} 
	return config, nil
}
