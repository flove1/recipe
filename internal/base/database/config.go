package database

import (
	"fmt"
	"net/url"
)

type Config struct {
	Host     string
	Port     string
	Name     string
	Username string
	Password string
}

func (c *Config) toDSN() string {
	q := url.Values{}
	q.Add("sslmode", "disable")

	u := url.URL{
		Scheme:   "postgresql",
		User:     url.UserPassword(c.Username, c.Password),
		Host:     fmt.Sprintf("%s:%s", c.Host, c.Port),
		Path:     c.Name,
		RawQuery: q.Encode(),
	}

	return u.String()
}
