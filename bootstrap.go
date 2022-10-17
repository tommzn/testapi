package main

import (
	"context"

	config "github.com/tommzn/go-config"
	log "github.com/tommzn/go-log"
	secrets "github.com/tommzn/go-secrets"
	core "github.com/tommzn/hdb-core"
)

func bootstrap(conf config.Config, ctx context.Context) (*core.Minion, error) {

	secretsManager := newSecretsManager()
	if conf == nil {
		conf = loadConfig()
	}
	logger := newLogger(conf, secretsManager, ctx)
	server := newServer(logger)
	return core.NewMinion(server), nil
}

// loadConfig from config file.
func loadConfig() config.Config {

	conf, err := config.NewConfigSource().Load()
	if err != nil {
		exitOnError(err)
	}
	return conf
}

// newSecretsManager retruns a new secrets manager from passed config.
func newSecretsManager() secrets.SecretsManager {
	secretsManager := secrets.NewDockerecretsManager("/run/secrets/token")
	return secretsManager
}

// newLogger creates a new logger from  passed config.
func newLogger(conf config.Config, secretsMenager secrets.SecretsManager, ctx context.Context) log.Logger {
	logger := log.NewLoggerFromConfig(conf, secretsMenager)
	logger = log.WithNameSpace(logger, "api-tester")
	return log.WithK8sContext(logger)
}
