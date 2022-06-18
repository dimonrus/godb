package godb

import "fmt"

// ConnectionConfig Client connection config
type ConnectionConfig struct {
	Host                   string
	Port                   int
	Name                   string
	User                   string
	Password               string
	MaxConnections         int `yaml:"maxConnections"`
	MaxIdleConnections     int `yaml:"maxIdleConnections"`
	ConnectionIdleLifetime int `yaml:"connectionIdleLifetime"`
}

// * disable - No SSL
// * require - Always SSL (skip verification)
// * verify-ca - Always SSL (verify that the certificate presented by the
//		server was signed by a trusted CA)
// * verify-full - Always SSL (verify that the certification presented by
//		the server was signed by a trusted CA and the server host name
//		matches the one in the certificate)

// PostgresConnectionConfig Postgres connection config
type PostgresConnectionConfig struct {
	ConnectionConfig `yaml:",inline"`
	SSLMode          string `yaml:"sslMode"`
	BinaryParameters bool   `yaml:"binaryParameters"`
}

// To string
func (pcc *PostgresConnectionConfig) String() string {
	stringConnection := "host=%s port=%d user=%s password=%s dbname=%s"
	if pcc.SSLMode != "" {
		stringConnection += " sslmode=" + pcc.SSLMode
	}
	if pcc.BinaryParameters {
		stringConnection += " binary_parameters=yes"
	}
	return fmt.Sprintf(stringConnection, pcc.Host, pcc.Port, pcc.User, pcc.Password, pcc.Name)
}

// GetDbType Get database type
func (pcc *PostgresConnectionConfig) GetDbType() string {
	return "postgres"
}

// GetMaxConnection Get Max Connection
func (cc *ConnectionConfig) GetMaxConnection() int {
	return cc.MaxConnections
}

// GetMaxIdleConns Connection max idle connections
func (cc *ConnectionConfig) GetMaxIdleConns() int {
	return cc.MaxIdleConnections
}

// GetConnMaxLifetime Connection idle lifetime
func (cc *ConnectionConfig) GetConnMaxLifetime() int {
	return cc.ConnectionIdleLifetime
}
