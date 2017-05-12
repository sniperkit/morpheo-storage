package main

import "flag"

// StorageConfig holds the configuration variables for the storage API
type StorageConfig struct {
	// API Server settings
	Hostname string
	Port     int
	CertFile string
	KeyFile  string

	// Database configuration
	DBHost string
	DBPort int
	DBUser string
	DBPass string
	DBName string

	// Database migration flags
	DBMigrationsDir string
	DBRollback      bool

	// Local (disk) blob store configuration
	DataDir string
}

// TLSOn returns true if TLS credentials have been provided. The API will then
// serve requests over TLS.
func (c *StorageConfig) TLSOn() bool {
	return c.CertFile != "" && c.KeyFile != ""
}

// NewStorageConfig computes the configuration object parsing CLI flags
func NewStorageConfig() (conf *StorageConfig) {
	var (
		hostname string
		port     int
		certFile string
		keyFile  string

		dbHost string
		dbPort int
		dbUser string
		dbPass string
		dbName string

		dbMigrationsDir string
		dbRollback      bool

		dataDir string
	)

	// CLI Flags
	flag.StringVar(&hostname, "host", "0.0.0.0", "The hostname our server will be listening on")
	flag.IntVar(&port, "port", 8000, "The port our compute API will be listening on")
	flag.StringVar(&certFile, "cert", "", "The TLS certs to serve to clients (leave blank for no TLS)")
	flag.StringVar(&keyFile, "key", "", "The TLS key used to encrypt connection (leave blank for no TLS)")

	flag.StringVar(&dbHost, "db-host", "postgres", "The hostname of the postgres database (default: postgres)")
	flag.IntVar(&dbPort, "db-port", 5432, "The database port")
	flag.StringVar(&dbName, "db-name", "morpheo_storage", "The database name (default: morpheo_storage)")
	flag.StringVar(&dbUser, "db-user", "storage", "The database user (default: storage)")
	flag.StringVar(&dbPass, "db-pass", "tooshort", "The database password to use (default: tooshort)")

	flag.StringVar(&dbMigrationsDir, "db-migrations-dir", "/migrations", "The database migrations directory (default: /migrations)")
	flag.BoolVar(&dbRollback, "db-rollback", false, "if true, rolls back the last migration (default: false)")

	flag.StringVar(&dataDir, "data-dir", "/data", "The directory to store blob data under (default: /data). Note that this only applies when using local storage")

	flag.Parse()

	// Let's create the config structure
	conf = &StorageConfig{
		Hostname: hostname,
		Port:     port,
		CertFile: certFile,
		KeyFile:  keyFile,

		DBHost: dbHost,
		DBPort: dbPort,
		DBUser: dbUser,
		DBPass: dbPass,
		DBName: dbName,

		DBMigrationsDir: dbMigrationsDir,
		DBRollback:      dbRollback,

		DataDir: dataDir,
	}
	return
}