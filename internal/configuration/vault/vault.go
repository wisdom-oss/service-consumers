package vault

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"time"

	"github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/api/auth/userpass"
)

type Vault struct {
	c       *api.Client     // acutal api client
	l       *api.Secret     // login information
	dbCreds *api.Secret     // database credentials
	ctx     context.Context // vault context
}

const (
	newLeaseTTL = 4 * time.Hour
)

const (
	EnvVaultUsername = "VAULT_USER"
	EnvVaultPassword = "VAULT_PASSWORD"
	EnvVaultPaths    = "VAULT_PATHS"
)

func (v *Vault) Initialize() error {
	v.ctx = context.Background()
	if _, set := os.LookupEnv(api.EnvVaultAddress); !set {
		return fmt.Errorf("no address set in %s", api.EnvVaultAddress)
	}

	conf := api.DefaultConfig()
	if err := conf.ReadEnvironment(); err != nil {
		return fmt.Errorf("unable to read vault environment configuration: %w", err)
	}

	client, err := api.NewClient(conf)
	if err != nil {
		return fmt.Errorf("unable to create vault client: %w", err)
	}

	v.c = client
	return nil
}

func (v *Vault) Login() error {
	username, set := os.LookupEnv(EnvVaultUsername)
	if !set {
		return fmt.Errorf("no username set in %s", EnvVaultUsername)
	}

	authData, err := userpass.NewUserpassAuth(username, &userpass.Password{FromEnv: EnvVaultPassword})
	if err != nil {
		return fmt.Errorf("unable to construct login data: %w", err)
	}

	s, err := v.c.Auth().Login(context.Background(), authData)
	if err != nil {
		return fmt.Errorf("unable to login into vault: %w", err)
	}

	if err := os.Setenv(api.EnvVaultToken, s.Auth.ClientToken); err != nil {
		return fmt.Errorf("unable to vault token to environment for viper remote config: %w", err)
	}

	v.l = s

	return nil
}

func (v *Vault) AutoLogin() {
	l := slog.New(slog.NewTextHandler(os.Stdout, nil))
	for {
		if err := v.Login(); err != nil {
			l.LogAttrs(v.ctx, slog.LevelWarn, "unable to autologin into vault", slog.String("error", err.Error()))
			time.Sleep(1 * time.Minute)
			continue
		}

		if err := v.manageSecretLifecycle(v.l, "login"); err != nil {
			l.LogAttrs(v.ctx, slog.LevelError, "unable to manage secret lifecycle", slog.String("error", err.Error()))
			return
		}
	}
}

func (v *Vault) AutoRenewDatabaseCredentials() {
	l := slog.New(slog.NewTextHandler(os.Stdout, nil))
	for {
		if err := v.Login(); err != nil {
			l.LogAttrs(v.ctx, slog.LevelWarn, "unable to generate database credentials", slog.String("error", err.Error()))
			time.Sleep(1 * time.Minute)
			continue
		}

		if err := v.manageSecretLifecycle(v.dbCreds, "database-credentials"); err != nil {
			l.LogAttrs(v.ctx, slog.LevelError, "unable to manage secret lifecycle", slog.String("error", err.Error()))
			return
		}
	}
}

func (v *Vault) DatabaseCredentials(mount, role string) (username, password string, err error) {
	if v.dbCreds != nil {
		var ok bool
		username, ok = v.dbCreds.Data["username"].(string)
		if !ok {
			goto request
		}

		password, ok = v.dbCreds.Data["password"].(string)
		if !ok {
			goto request
		}

		return username, password, nil
	}
request:
	path, err := url.JoinPath(mount, "/creds/", role)
	if err != nil {
		return "", "", err
	}

	v.dbCreds, err = v.c.Logical().Read(path)
	if err != nil {
		return "", "", fmt.Errorf("unable to get database credentials: %w", err)
	}

	username, ok := v.dbCreds.Data["username"].(string)
	if !ok {
		goto request
	}

	password, ok = v.dbCreds.Data["password"].(string)
	if !ok {
		goto request
	}

	return username, password, nil
}

func (v *Vault) ServerAddress() string {
	return v.c.Address()
}

func (v *Vault) manageSecretLifecycle(s *api.Secret, label string) error {
	if (s.Auth != nil && !s.Auth.Renewable) && !s.Renewable {
		return errors.New("supplied secret is not renewable")
	}

	w, err := v.c.NewLifetimeWatcher(&api.LifetimeWatcherInput{
		Secret:    s,
		Increment: int(newLeaseTTL),
	})
	if err != nil {
		return fmt.Errorf("unable to instantiate new lifetime watcher: %w", err)
	}

	var logAttrs []slog.Attr
	logAttrs = append(logAttrs, slog.String("watcherLabel", label))
	logAttrs = append(logAttrs, slog.Bool("loginSecret", s.Auth != nil))

	go w.Start()
	defer w.Stop()

	for {
		select {
		case err := <-w.DoneCh():
			if err != nil {
				slog.LogAttrs(context.Background(), slog.LevelError, "failed to renew secret lease", logAttrs...)
			}
			slog.LogAttrs(context.Background(), slog.LevelError, "secret reached max ttl and cannot be renewed", logAttrs...)
			return nil
		case <-w.RenewCh():
			slog.LogAttrs(context.Background(), slog.LevelInfo, "renewed secret", logAttrs...)
		}
	}
}
