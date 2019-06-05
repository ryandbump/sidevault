package cmd

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/hashicorp/vault/api"
)

var (
	frequency = "frequency"
	lease     = "lease"
)

var renewCmd = &cobra.Command{
	Use:   "renew",
	Short: "Renew a provided Vault token once its TTL is below the threshold.",
	Long: `Loops forever and checks the status of the provided Vault token
every 30 seconds unless a different frequency is provided. A renewal attempt 
is made if TTL <= creation TTL / 2. The token lease is extended to the original
creation TTL unless a different lease is provided.`,
	Run: func(cmd *cobra.Command, args []string) {
		token, err := readToken()
		if err != nil {
			log.Fatal(err)
		}

		client, err := api.NewClient(api.DefaultConfig())
		if err != nil {
			log.Fatal(errors.Wrap(err, "failed to intialize client"))
		}

		client.SetToken(token)

		for {
			secret, err := lookup(client)
			if err != nil {
				log.Fatal(err)
			}

			creationTTL, err := secret.Data["creation_ttl"].(json.Number).Int64()
			if err != nil {
				log.Fatal(err)
			}

			currentTTL, err := secret.Data["ttl"].(json.Number).Int64()
			if err != nil {
				log.Fatal(err)
			}

			if currentTTL <= creationTTL/2 {
				log.Print("token ttl below threshold, renewing token")
				renew(client, determineLease(secret))
				log.Print("token renewed successfully")
			}

			time.Sleep(time.Duration(viper.GetInt(frequency)) * time.Second)
		}
	},
}

func init() {
	rootCmd.AddCommand(renewCmd)

	intFlag(
		renewCmd,
		frequency,
		30,
		"Delay between token ttl checks, in seconds.",
	)

	intFlag(
		renewCmd,
		lease,
		0,
		"Request lease increment when renewing, in seconds.",
	)
}

func lookup(client *api.Client) (*api.Secret, error) {
	secret, err := client.Auth().Token().LookupSelf()
	if err != nil {
		return nil, errors.Wrap(err, "failed to lookup token")
	}

	return secret, nil
}

func renew(client *api.Client, lease int) (*api.Secret, error) {
	secret, err := client.Auth().Token().RenewSelf(lease)
	if err != nil {
		return nil, errors.Wrap(err, "failed to renew token")
	}

	return secret, nil
}

func determineLease(secret *api.Secret) int {
	providedLease := viper.GetInt(lease)
	if providedLease == 0 {
		creationTTL, err := secret.Data["creation_ttl"].(json.Number).Int64()
		if err != nil {
			log.Fatal(err)
		}
		return int(creationTTL)
	}

	return providedLease
}

func readToken() (string, error) {
	data, err := ioutil.ReadFile(viper.GetString(tokenPath))
	if err != nil {
		return "", err
	}

	return string(bytes.TrimSpace(data)), nil
}
