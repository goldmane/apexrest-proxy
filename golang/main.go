package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2/clientcredentials"
)

type OauthConfig struct {
	ClientId     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

type ProxyConfig struct {
	TargetDomain string       `json:"target_domain"`
	InstanceUrl  string       `json:"instance_url,omitempty"`
	LoginUrl     string       `json:"login_url,omitempty"`
	Oauth        *OauthConfig `json:"oauth"`
	IncludePath  bool         `json:"include_path"`
}

var (
	simulate      bool = false
	proxy_configs map[string]ProxyConfig
)

const (
	HEADER_ORIGINAL_DESTINATION string = "x-apexrestproxy-original-destination"
)

func loadConfig() map[string]ProxyConfig {
	var ret map[string]ProxyConfig
	proxy_configs := os.Getenv("PROXY_CONFIGS")
	as_bytes := []byte(proxy_configs)
	err := json.Unmarshal(as_bytes, &ret)
	if err != nil {
		log.Fatalf("Error unmarshalling the proxy configs: %s\n", err.Error())
	}
	return ret
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading the env config: %s\n", err.Error())
	}

	simulate, err = strconv.ParseBool(os.Getenv("SIMULATE"))
	proxy_configs = loadConfig()

	router := chi.NewRouter()
	router.Use(middleware.Logger)

	// create a route handler for each
	for k, config := range proxy_configs {
		url, err := url.Parse(config.TargetDomain)
		if err != nil {
			log.Fatalf("Error parsing target domain into URL: %s\n", err.Error())
		}
		proxy := httputil.NewSingleHostReverseProxy(url)
		original_director := proxy.Director
		proxy.Director = func(r *http.Request) {
			original_director(r)
			if !simulate {
				r.Host = r.URL.Host
				r.URL.Path = strings.TrimSuffix(r.URL.Path, "/")
			} else {
				r.Header.Set(HEADER_ORIGINAL_DESTINATION, r.URL.String())
				r.Host = "localhost:3000"
				r.URL.Host = "localhost:3000"
				r.URL.Path = "/__internal__/simulated"
				r.URL.Scheme = "http"
				log.Printf("Simulated proxy -> %s\n", r.URL.String())
			}
		}
		path_pattern := fmt.Sprintf("/%s", k)

		if config.IncludePath {
			path_pattern += "/*"
		}

		log.Printf("Registering client %s\n", k)

		handler := func(p *httputil.ReverseProxy) func(http.ResponseWriter, *http.Request) {
			return func(w http.ResponseWriter, r *http.Request) {
				// move the auth header (if already set)
				prev_auth_header := r.Header["Authorization"]
				if len(prev_auth_header) > 0 {
					r.Header["X-Original-Authorization"] = prev_auth_header
					r.Header["Authorization"] = []string{}
				}
				// set the oauth parts (if available)
				if config.Oauth != nil {
					ctx := context.Background()
					conf := &clientcredentials.Config{
						ClientID:     config.Oauth.ClientId,
						ClientSecret: config.Oauth.ClientSecret,
						TokenURL:     fmt.Sprintf("%s/services/oauth2/token", config.InstanceUrl),
					}
					token, err := conf.Token(ctx)
					if err != nil {
						log.Fatalf("Error getting client creds token: %s\n", err.Error())
					}

					r.Header["Authorization"] = []string{fmt.Sprintf("Bearer %s", token.AccessToken)}
				}
				if config.IncludePath {
					// strip off key part the path
					r.URL.Path = strings.TrimPrefix(r.URL.Path, fmt.Sprintf("/%s", k))
				} else {
					r.URL.Path = ""
				}
				// serve
				p.ServeHTTP(w, r)
			}
		}
		router.HandleFunc(path_pattern, handler(proxy))
	}

	log.Printf("Registered %d routes\n", len(router.Routes()))

	port := os.Getenv("PORT")
	if simulate {
		router.HandleFunc("/__internal__/simulated", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-ApexRestProxy-Mode", "simulated")
			w.Header().Set(HEADER_ORIGINAL_DESTINATION, r.Header.Get(HEADER_ORIGINAL_DESTINATION))
			w.WriteHeader(200)
		})
		log.Println("Starting in SIMULATE mode")
	}
	log.Printf("Proxy starting on port: %s\n", port)
	err = http.ListenAndServe(fmt.Sprintf(":%s", port), router)
	if err != nil {
		log.Fatalf("Error starting proxy on port %s: %s\n", port, err.Error())
	}
}
