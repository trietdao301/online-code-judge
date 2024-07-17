package configs

import "time"

type Token struct {
	ExpiresIn string `yaml:"expires_in"`
}

func (t Token) GetExpiresInDuration() (time.Duration, error) {
	return time.ParseDuration(t.ExpiresIn)
}
