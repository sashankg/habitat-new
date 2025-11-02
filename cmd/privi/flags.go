package main

import (
	"fmt"
	"strings"

	altsrc "github.com/urfave/cli-altsrc/v3"
	yaml "github.com/urfave/cli-altsrc/v3/yaml"
	"github.com/urfave/cli/v3"
)

var (
	cDebug      = "debug"
	cDomain     = "domain"
	cDb         = "db"
	cPort       = "port"
	cHttpsCerts = "httpscerts"
	cKeyFile    = "keyfile"
)
var profiles []string

func getFlags() ([]cli.Flag, []cli.MutuallyExclusiveFlags) {
	return []cli.Flag{
		&cli.BoolFlag{
			Name:    cDebug,
			Usage:   "Enable debug mode",
			Sources: getSources(cDebug),
		},
		&cli.StringSliceFlag{
			Name:        "profile",
			Usage:       "The configuration profile to use.",
			TakesFile:   true,
			Destination: &profiles,
		},
		&cli.StringFlag{
			Name:     cDomain,
			Required: true,
			Usage:    "The publicly available domain at which the server can be found",
			Sources:  getSources(cDomain),
		},
		&cli.StringFlag{
			Name:    cDb,
			Usage:   "The path to the sqlite file to use as the backing database for this server",
			Value:   "./repo.db",
			Sources: getSources(cDb),
		},
		&cli.StringFlag{
			Name:    cPort,
			Usage:   "The port on which to run the server",
			Value:   "8000",
			Sources: getSources(cPort),
		},
		&cli.StringFlag{
			Name:    cHttpsCerts,
			Usage:   "The directory in which TLS certs can be found. Should contain fullchain.pem and privkey.pem",
			Sources: getSources(cHttpsCerts),
		},
		&cli.StringFlag{
			Name:      cKeyFile,
			Usage:     "The path to the key file to use for OAuth client metadata",
			Value:     "./key.jwk",
			TakesFile: true,
			Sources:   getSources(cKeyFile),
		},
	}, []cli.MutuallyExclusiveFlags{}
}

func getSources(name string) cli.ValueSourceChain {
	return cli.NewValueSourceChain(
		cli.EnvVar("HABITAT_"+strings.ToUpper(name)),
		&profilesSource{name: name},
	)
}

type profilesSource struct {
	name string
}

// GoString implements cli.ValueSource.
func (ps *profilesSource) GoString() string {
	return fmt.Sprintf("&profilesSource{name:%[1]q}", ps.name)
}

func (ps *profilesSource) String() string {
	return strings.Join(profiles, ",")
}

func (ps *profilesSource) Lookup() (string, bool) {
	sources := cli.ValueSourceChain{
		Chain: []cli.ValueSource{},
	}
	for range profiles {
		sources.Chain = append(
			sources.Chain,
			yaml.YAML(ps.name, altsrc.NewStringPtrSourcer(&profiles[0])),
		)
	}
	return sources.Lookup()
}
