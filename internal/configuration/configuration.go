package configuration

import (
	"errors"
	"fmt"
	"os"
	"strings"

	_ "github.com/dr4hcu5-jan/viper-vault/remote"
	_ "github.com/dr4hcu5-jan/viper-vault/remote/vault"

	"github.com/spf13/viper"

	"microservice/internal/configuration/vault"
)

var Default configuration

type configuration struct {
	i           *viper.Viper // viper instance
	t           string       // type of configuration
	vaultClient *vault.Vault // client used to access a hashicorp vault
	s           []string     // paths to the secrets to be read from the vault
	dbRole      string
	dbMount     string
}

func (c *configuration) Initialize() error {
	c.i = viper.New()
	c.setupDefaults()

	configurationType, set := os.LookupEnv(EnvConfigurationType)
	if !set {
		c.t = ConfigurationType_Local
	}

	switch configurationType {
	case ConfigurationType_Local:
		c.t = configurationType
		return c.initializeLocalReading()
	case ConfigurationType_Vault:
		c.t = configurationType
		return c.initializeVaultReading()
	default:
		return fmt.Errorf("unsupported value set in %s", EnvConfigurationType)
	}

}

func (c *configuration) Viper() *viper.Viper {
	return c.i
}

func (c *configuration) Read() error {
	switch c.t {
	case ConfigurationType_Local:
		return c.readLocalConfiguration()
	case ConfigurationType_Vault:
		return c.readVaultConfiguration()
	default:
		return errors.New("unsupported configuration type for reading configuration")
	}
}

func (c *configuration) initializeVaultReading() error {
	c.vaultClient = &vault.Vault{}
	if err := c.vaultClient.Initialize(); err != nil {
		return fmt.Errorf("unable to initialize vault client: %w", err)
	}

	if err := c.vaultClient.Login(); err != nil {
		return fmt.Errorf("unable to authenticate with configured vault: %w", err)
	}

	secretPaths, set := os.LookupEnv(vault.EnvVaultPaths)
	if !set {
		return fmt.Errorf("no paths to read secrets from set in %s", vault.EnvVaultPaths)
	}
	c.s = strings.Split(secretPaths, ",")

	for _, s := range c.s {
		if err := c.i.AddRemoteProvider("vault", c.vaultClient.ServerAddress(), s); err != nil {
			return err
		}
	}
	c.i.SetConfigType("json")

	return nil
}

func (c *configuration) readVaultConfiguration() error {
	if err := c.i.ReadRemoteConfig(); err != nil {
		return fmt.Errorf("unable to read remote configuration: %w", err)
	}

	if c.i.GetString(ConfigurationKey_DatabaseCredentialType) == DatabaseCredentialType_Static {
		return nil
	}

	c.dbRole = c.i.GetString(ConfigurationKey_DatabaseCredentialRole)

	username, password, err := c.vaultClient.DatabaseCredentials("databases", c.dbRole)
	if err != nil {
		return fmt.Errorf("unable to set up dynamic database credentials: %w", err)
	}

	// automatically renew the database credentials if the lease expires or
	// the credentials are expiring
	go c.vaultClient.AutoRenewDatabaseCredentials()

	c.i.Set(ConfigurationKey_DatabaseUser, username)
	c.i.Set(ConfigurationKey_DatabasePassword, password)

	return nil
}

func (c *configuration) RefreshDatabaseCredentials() error {
	if c.t != ConfigurationType_Vault {
		return errors.New("refreshing database credentials is only supported for vault configuration")
	}

	c.dbMount = c.i.GetString(ConfigurationKey_DatabaseCredentialMount)

	username, password, err := c.vaultClient.DatabaseCredentials(c.dbMount, c.dbRole)
	if err != nil {
		return fmt.Errorf("unable to set up dynamic database credentials: %w", err)
	}

	c.i.Set(ConfigurationKey_DatabaseUser, username)
	c.i.Set(ConfigurationKey_DatabasePassword, password)

	return nil
}

func (c *configuration) initializeLocalReading() error {
	c.i.SetConfigName("config")
	c.i.AddConfigPath("/etc/award/")
	c.i.AddConfigPath("/run/secrets/")
	c.i.AddConfigPath("$HOME/.config/award/")
	c.i.AddConfigPath(".")
	c.i.SetEnvPrefix("")
	c.i.SetEnvKeyReplacer(strings.NewReplacer(".", "__", "-", "_"))
	c.i.AutomaticEnv()

	if err := c.setupEnvironmentAliases(); err != nil {
		return fmt.Errorf("unable to setup environment variable aliases: %w", err)

	}

	return nil
}

func (c *configuration) setupEnvironmentAliases() error {
	for key, envVars := range environmentVariables {
		args := make([]string, len(envVars)+1)
		args[0] = key
		for idx, envVar := range envVars {
			args[idx+1] = envVar
		}
		if err := c.i.BindEnv(args...); err != nil {
			return err
		}
	}
	return nil
}

func (c *configuration) readLocalConfiguration() error {
	_ = c.i.ReadInConfig()

	return nil
}

func (c *configuration) setupDefaults() {
	for key, defaultValue := range defaults {
		c.i.SetDefault(key, defaultValue)
	}

}
