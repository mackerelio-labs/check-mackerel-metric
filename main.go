package main

import (
	"fmt"
	"os"
	"time"

	"github.com/jessevdk/go-flags"
	"github.com/mackerelio/checkers"

	"github.com/mackerelio/mackerel-agent/config"

	"github.com/mackerelio/mackerel-client-go"
)

type mackerelMetricOpts struct {
	Host     string `short:"H" long:"host" required:"true" description:"target host ID"`
	Metric   string `short:"n" long:"name" required:"true" description:"target metric name"`
	Warning  int    `short:"w" long:"warning" required:"true" description:"minute to be WARNING"`
	Critical int    `short:"c" long:"critical" required:"true" description:"minute to be CRITICAL"`
}

func main() {
	Do()
}

func Do() {
	opts, err := parseArgs(os.Args[1:])
	if err != nil {
		os.Exit(1)
	}
	ckr := opts.run()
	ckr.Name = "MackerelMetric"
	ckr.Exit()
}

func parseArgs(args []string) (*mackerelMetricOpts, error) {
	opts := &mackerelMetricOpts{}
	_, err := flags.ParseArgs(opts, args)
	return opts, err
}

func (opts *mackerelMetricOpts) run() *checkers.Checker {
	apikey := os.Getenv("MACKEREL_APIKEY")
	apibase := LoadApibaseFromConfig(config.DefaultConfig.Conffile)
	if apikey == "" {
		apikey = LoadApikeyFromConfig(config.DefaultConfig.Conffile)
	}
	if apibase == "" || apikey == "" {
		return checkers.Unknown("Not found apibase or apikey in " + config.DefaultConfig.Conffile)
	}

	to := time.Now().Unix()
	criticalFrom := to - int64(opts.Critical*60)
	warningFrom := to - int64(opts.Warning*60)

	client, err := mackerel.NewClientWithOptions(apikey, apibase, false)
	if err != nil {
		return checkers.Unknown(fmt.Sprintf("%v", err))
	}

	// CRITICAL
	metricValue, err := fetchHostMetricValues(client, opts.Host, opts.Metric, criticalFrom, to)
	if err != nil {
		return checkers.Unknown(fmt.Sprintf("%v", err))
	}
	if len(metricValue) == 0 {
		return checkers.Critical(fmt.Sprintf("no metric for %s has been posted since at least %d minutes ago", opts.Metric, opts.Critical))
	} else {
		// WARNING check
		last := metricValue[len(metricValue)-1].Time // newest metric
		if last < warningFrom {
			return checkers.Warning(fmt.Sprintf("no metric for %s has been posted since at least %d minutes ago", opts.Metric, opts.Warning))
		}
	}

	return checkers.Ok(fmt.Sprintf("metric for %s continues to post", opts.Metric))
}

func fetchHostMetricValues(client *mackerel.Client, hostID string, metricName string, from int64, to int64) ([]mackerel.MetricValue, error) {
	metricValue, err := client.FetchHostMetricValues(hostID, metricName, from, to)
	if err != nil {
		return nil, err
	}
	return metricValue, nil
}

func LoadApibaseFromConfig(conffile string) string {
	conf, err := config.LoadConfig(conffile)
	if err != nil {
		return ""
	}
	return conf.Apibase
}

func LoadApikeyFromConfig(conffile string) string {
	conf, err := config.LoadConfig(conffile)
	if err != nil {
		return ""
	}
	return conf.Apikey
}
