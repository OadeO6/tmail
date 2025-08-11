package config

struct Config {
  SmtpPort string `env:"SMTP_PORT required:true"`
}

BasicConfigT := &Config{};
LoadConfig(BasicConfig)
BasicConfig = *BasicConfigT
