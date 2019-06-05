package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/hashicorp/vault/api"
)

var (
	mountPath   = "mount-path"
	role        = "role"
	saTokenPath = "sa-token-path"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authenticate with Vault and save token to filesystem.",
	Long: `Uses the ServiceAccount token and provided role to authenticate with Vault.
The returned token and accessor will be saved to the filesystem. 

The role flag is required.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := validate(); err != nil {
			cmd.Help()
			os.Exit(1)
		}

		loginData, err := generateLoginData()
		if err != nil {
			log.Fatal(err)
		}

		secret, err := authenticate(loginData)
		if err != nil {
			log.Fatal(err)
		}

		err = save(secret.Auth.ClientToken, viper.GetString(tokenPath))
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("token saved to %s", viper.GetString(tokenPath))

		err = save(secret.Auth.Accessor, viper.GetString(accessorPath))
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("accessor saved to %s", viper.GetString(accessorPath))
	},
}

func init() {
	rootCmd.AddCommand(authCmd)

	stringFlag(
		authCmd,
		mountPath,
		"kubernetes",
		"Mount path for the Kubernetes authentication backend.",
	)

	stringFlag(
		authCmd,
		role,
		"",
		"Role to use for Vault authentication.",
	)

	stringFlag(
		authCmd,
		saTokenPath,
		"/var/run/secrets/kubernetes.io/serviceaccount/token",
		"File system path to the Kubernetes ServiceAccount token.",
	)
}

func authenticate(data map[string]interface{}) (*api.Secret, error) {
	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize client")
	}

	loginPath := fmt.Sprintf("auth/%v/login", viper.GetString(mountPath))

	secret, err := client.Logical().Write(loginPath, data)
	if err != nil {
		return nil, errors.Wrap(err, "failed to login")
	}

	return secret, nil
}

func generateLoginData() (map[string]interface{}, error) {
	loginData := make(map[string]interface{})

	jwt, err := readJwtToken()
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate login data")
	}

	providedRole := viper.GetString(role)
	if role == "" {
		return nil, errors.New("role must be specified")
	}

	loginData["jwt"] = jwt
	loginData["role"] = providedRole

	return loginData, nil
}

func readJwtToken() (string, error) {
	data, err := ioutil.ReadFile(viper.GetString(saTokenPath))
	if err != nil {
		return "", errors.Wrap(err, "failed to read jwt token")
	}

	return string(bytes.TrimSpace(data)), nil
}

func save(token, path string) error {
	err := ioutil.WriteFile(path, []byte(token), 0600)
	if err != nil {
		return errors.Wrap(err, "failed to save token")
	}
	return nil
}

func validate() error {
	if viper.GetString(role) == "" {
		return errors.New("role not set")
	}
	return nil
}
