package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/fsnotify/fsnotify"
	whoisparser "github.com/likexian/whois-parser"
	"github.com/numero33/domain_exporter/whois"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"net/http"
	"os"
	"time"
)

var (
	version = "development"
	// How often to check domains
	checkRate = 12 * time.Hour

	configFile *string
	httpBind   *string

	domainExpiration = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "domain_expiration_seconds",
			Help: "UNIX timestamp when the WHOIS record states this domain will expire",
		},
		[]string{"domain"},
	)
	domainLastChange = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "domain_last_change_seconds",
			Help: "UNIX timestamp when the WHOIS record states this domain will expire",
		},
		[]string{"domain"},
	)
	parsedDomain = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "domain_parsed",
			Help: "That the domain was parsed",
		},
		[]string{"domain"},
	)
)

func init() {
	configFile = flag.String("config", "domains.yml", "domain_exporter configuration file.")
	httpBind = flag.String("bind", ":9203", "The address to listen on for HTTP requests.")
	debug := flag.Bool("debug", false, "sets log level to debug")
	versionFlg := flag.Bool("version", false, "prints the version")

	flag.Parse()

	if *versionFlg {
		fmt.Println("domain_exporter " + version)
		os.Exit(0)
	}

	// log
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})

	// config
	viper.SetConfigFile(*configFile)
	viper.AddConfigPath(".")
	viper.OnConfigChange(func(e fsnotify.Event) {
		fmt.Println("Config file changed:", e.Name, viper.GetStringSlice("domains"))
	})
	viper.WatchConfig()
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		panic(fmt.Errorf("fatal error config file: %w", err))
	}

}

func main() {

	log.Debug().Str("config", *configFile).Any("domains", viper.GetStringSlice("domains")).Msg("Configuration file")

	prometheus.Register(domainExpiration)
	prometheus.Register(domainLastChange)
	prometheus.Register(parsedDomain)

	ctx := context.Background()

	go func() {
		for {
			for _, domain := range viper.GetStringSlice("domains") {
				info, err := lookup(ctx, domain)
				if err != nil {
					log.Warn().Err(err).Str("domain", domain).Msg("Error looking up domain")
					parsedDomain.WithLabelValues(domain).Set(0)
					continue
				}
				parsedDomain.WithLabelValues(domain).Set(1)

				if info.Domain.ExpirationDateInTime != nil {
					domainExpiration.WithLabelValues(domain).Set(float64((*info.Domain.ExpirationDateInTime).Unix()))
					log.Info().Str("exp", info.Domain.ExpirationDate).Msg("Domain expiration date")
				}

				if info.Domain.UpdatedDateInTime != nil {
					domainLastChange.WithLabelValues(domain).Set(float64((*info.Domain.UpdatedDateInTime).Unix()))
					log.Info().Str("exp", info.Domain.UpdatedDate).Msg("Domain last change date")
				}

				continue
			}
			time.Sleep(checkRate)
		}
	}()

	http.Handle("/metrics", promhttp.Handler())
	log.Info().Str("httpBind", *httpBind).Msg("Listening on port")
	if err := http.ListenAndServe(*httpBind, nil); err != nil {
		log.Error().Err(err).Msg("Error starting HTTP server")
		os.Exit(1)
	}
}

func lookup(ctx context.Context, domain string) (whoisparser.WhoisInfo, error) {
	log.Debug().Str("domain", domain).Msg("Looking up domain.")
	result, _, err := whois.WhoIs(ctx, domain)
	if err != nil {
		return whoisparser.WhoisInfo{}, err
	}

	return whoisparser.Parse(string(result))
}
