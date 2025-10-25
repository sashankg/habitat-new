package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"net/http/httputil"
	"os"

	jose "github.com/go-jose/go-jose/v3"

	"github.com/bluesky-social/indigo/atproto/identity"
	"github.com/eagraf/habitat-new/internal/auth"
	"github.com/eagraf/habitat-new/internal/oauthserver"
	"github.com/eagraf/habitat-new/internal/permissions"
	"github.com/eagraf/habitat-new/internal/privi"
	"github.com/gorilla/sessions"
	"github.com/rs/zerolog/log"
)

const (
	defaultPort = "443"
)

var (
	domainPtr = flag.String(
		"domain",
		"",
		"The publicly available domain at which the server can be found",
	)
	repoPathPtr = flag.String(
		"path",
		"./repo.db",
		"The path to the sqlite file to use as the backing database for this server",
	)
	portPtr = flag.String(
		"port",
		defaultPort,
		"The port on which to run the server. Default 9000",
	)
	certsFilePtr = flag.String(
		"certs",
		"/etc/letsencrypt/live/habitat.network/",
		"The directory in which TLS certs can be found. Should contain fullchain.pem and privkey.pem",
	)
	helpFlag = flag.Bool("help", false, "Display this menu.")

	keyFilePtr = flag.String(
		"key",
		"privi.jwk",
		"The path to the JWK file to use for signing tokens",
	)
)

func main() {
	flag.Parse()

	if helpFlag != nil && *helpFlag {
		flag.PrintDefaults()
		os.Exit(0)
	}

	if domainPtr == nil || *domainPtr == "" {
		fmt.Println("domain flag is required; -h to see help menu")
		os.Exit(1)
	} else if repoPathPtr == nil || *repoPathPtr == "" {
		fmt.Println("No repo path specifiedl using default value ./repo.db")
	} else if portPtr == nil || *portPtr == "" {
		fmt.Printf("No port specified; using default %s\n", defaultPort)
		*portPtr = defaultPort
	}

	fmt.Printf(
		"Using %s as domain and %s as repo path; starting private data server\n",
		*domainPtr,
		*repoPathPtr,
	)

	// Create database file if it does not exist
	// TODO: this should really be taken in as an argument or env variable
	priviRepoPath := *repoPathPtr
	_, err := os.Stat(priviRepoPath)
	if errors.Is(err, os.ErrNotExist) {
		fmt.Println("Privi repo file does not exist; creating...")
		_, err := os.Create(priviRepoPath)
		if err != nil {
			log.Err(err).Msgf("unable to create privi repo file at %s", priviRepoPath)
		}
	} else if err != nil {
		log.Err(err).Msgf("error finding privi repo file")
	}

	priviDB, err := sql.Open("sqlite3", priviRepoPath)
	if err != nil {
		log.Fatal().Err(err).Msg("unable to open sqlite file backing privi server")
	}

	repo, err := privi.NewSQLiteRepo(priviDB)
	if err != nil {
		log.Fatal().Err(err).Msg("unable to setup privi sqlite db")
	}

	adapter, err := permissions.NewSQLiteStore(priviDB)
	if err != nil {
		log.Fatal().Err(err).Msg("unable to setup permissions store")
	}
	priviServer := privi.NewServer(adapter, repo)

	mux := http.NewServeMux()

	loggingMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			x, err := httputil.DumpRequest(r, true)
			if err != nil {
				http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
				return
			}
			fmt.Println("Got a request: ", string(x))
			next.ServeHTTP(w, r)
		})
	}

	// privi routes
	mux.HandleFunc("/xrpc/com.habitat.putRecord", priviServer.PutRecord)
	mux.HandleFunc("/xrpc/com.habitat.getRecord", priviServer.GetRecord)
	mux.HandleFunc("/xrpc/network.habitat.uploadBlob", priviServer.UploadBlob)
	mux.HandleFunc("/xrpc/com.habitat.listPermissions", priviServer.ListPermissions)
	mux.HandleFunc("/xrpc/com.habitat.addPermission", priviServer.AddPermission)
	mux.HandleFunc("/xrpc/com.habitat.removePermission", priviServer.RemovePermission)

	mux.HandleFunc("/.well-known/did.json", func(w http.ResponseWriter, r *http.Request) {
		template := `{
  "id": "did:web:%s",
  "@context": [
    "https://www.w3.org/ns/did/v1",
    "https://w3id.org/security/multikey/v1", 
    "https://w3id.org/security/suites/secp256k1-2019/v1"
  ],
  "service": [
    {
      "id": "#habitat",
      "serviceEndpoint": "https://%s",
      "type": "HabitatServer"
    }
  ]
}`
		domain := *domainPtr
		_, err := fmt.Fprintf(w, template, domain, domain)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	oauthServer, oauthClient := setupOAuthServer(*domainPtr, *keyFilePtr)

	// auth routes
	mux.HandleFunc("/oauth-callback", func(w http.ResponseWriter, r *http.Request) {
		if err := oauthServer.HandleCallback(w, r); err != nil {
			log.Error().Err(err).Msg("error handling oauth callback")
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
	mux.HandleFunc("/client-metadata.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(oauthClient.ClientMetadata())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
	mux.HandleFunc("/oauth/authorize", func(w http.ResponseWriter, r *http.Request) {
		if err := oauthServer.HandleAuthorize(w, r); err != nil {
			log.Error().Err(err).Msg("error handling oauth authorize request")
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
	mux.HandleFunc("/oauth/token", oauthServer.HandleToken)

	s := &http.Server{
		Handler: loggingMiddleware(mux),
		Addr:    fmt.Sprintf(":%s", *portPtr),
	}

	fmt.Println("Starting server on port :" + *portPtr)
	err = s.ListenAndServeTLS(
		fmt.Sprintf("%s%s", *certsFilePtr, "fullchain.pem"),
		fmt.Sprintf("%s%s", *certsFilePtr, "privkey.pem"),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("error serving http")
	}
}

func setupPriviServer() *privi.Server {
	// Create database file if it does not exist
	// TODO: this should really be taken in as an argument or env variable
	priviRepoPath := *repoPathPtr
	_, err := os.Stat(priviRepoPath)
	if errors.Is(err, os.ErrNotExist) {
		fmt.Println("Privi repo file does not exist; creating...")
		_, err := os.Create(priviRepoPath)
		if err != nil {
			log.Err(err).Msgf("unable to create privi repo file at %s", priviRepoPath)
		}
	} else if err != nil {
		log.Err(err).Msgf("error finding privi repo file")
	}

	priviDB, err := sql.Open("sqlite3", priviRepoPath)
	if err != nil {
		log.Fatal().Err(err).Msg("unable to open sqlite file backing privi server")
	}

	repo, err := privi.NewSQLiteRepo(priviDB)
	if err != nil {
		log.Fatal().Err(err).Msg("unable to setup privi sqlite db")
	}

	adapter, err := permissions.NewSQLiteStore(priviDB)
	if err != nil {
		log.Fatal().Err(err).Msg("unable to setup permissions store")
	}
	return privi.NewServer(adapter, repo)
}

func setupOAuthServer(
	domain string,
	keyFile string,
) (*oauthserver.OAuthServer, auth.OAuthClient) {
	// Read JWK from file
	jwkBytes, err := os.ReadFile(keyFile)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			log.Fatal().Err(err).Msgf("unable to read key file at %s", keyFile)
		}
		// Generate ECDSA key using P-256 curve with crypto/rand for secure randomness
		key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			log.Fatal().Err(err).Msgf("failed to generate key")
		}
		// Create JWK from the generated key
		jwk := jose.JSONWebKey{
			Key:       key,
			KeyID:     "habitat",
			Algorithm: string(jose.ES256),
			Use:       "sig",
		}
		jwkBytes, err = json.MarshalIndent(jwk, "", "  ")
		if err != nil {
			log.Fatal().Err(err).Msgf("failed to marshal JWK")
		}
		if err := os.WriteFile(keyFile, jwkBytes, 0o600); err != nil {
			log.Fatal().Err(err).Msgf("failed to write key to file")
		}
		log.Info().Msgf("created key file at %s", keyFile)
	}

	oauthProvider := oauthserver.NewProvider()

	oauthClient, err := auth.NewOAuthClient(
		"https://"+domain+"/client-metadata.json", /*clientId*/
		"https://"+domain,                         /*clientUri*/
		"https://"+domain+"/oauth-callback",       /*redirectUri*/
		jwkBytes,                                  /*secretJwk*/
	)

	return oauthserver.NewOAuthServer(
		oauthProvider,
		oauthClient,
		sessions.NewCookieStore([]byte("my super secret signing password")),
		identity.DefaultDirectory(),
	), oauthClient
}
