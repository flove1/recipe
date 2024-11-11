package database

import (
	"net/url"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func NewNeo4jConnection(connectionString string) (neo4j.DriverWithContext, error) {
	url, err := url.Parse(connectionString)
	if err != nil {
		return nil, err
	}

	user := url.User.Username()
	password, _ := url.User.Password()

	driver, err := neo4j.NewDriverWithContext(connectionString, neo4j.BasicAuth(user, password, ""))
	if err != nil {
		return nil, err
	}

	return driver, nil
}
