package checkmackerelmetric

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/alexflint/go-arg"
	"github.com/mackerelio/checkers"

	"github.com/mackerelio/mackerel-agent/config"

	"github.com/mackerelio/mackerel-client-go"
)

type mackerelMetricOpts struct {
	Host     string `arg:"-H,--host" help:"target host ID" placeholder:"HOST_ID"`
	Service  string `arg:"-s,--service" help:"target service name" placeholder:"SERVICE_NAME"`
	Metric   string `arg:"-n,--name,required" help:"target metric name" placeholder:"METRIC_NAME"`
	Warning  uint   `arg:"-w,--warning,required" help:"minute to be WARNING (MINUTE: 1-1441)" placeholder:"MINUTE"`
	Critical uint   `arg:"-c,--critical,required" help:"minute to be CRITICAL (MINUTE: 1-1441)" placeholder:"MINUTE"`
	StatusAs string `arg:"--status-as" help:"overwrite status=new_status, support multiple comma separates"`
}

var version string
var revision string

func (mackerelMetricOpts) Version() string {
	return fmt.Sprintf("version %s (rev %s)", version, revision)
}

func Do() {
	opts, maps, err := parseArgs(os.Args[1:])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	ckr := opts.run()
	ckr.Name = "MackerelMetric"
	ckr.ExitStatusAs(maps)
}

func parseArgs(args []string) (*mackerelMetricOpts, map[checkers.Status]checkers.Status, error) {
	var mo mackerelMetricOpts
	p, _ := arg.NewParser(arg.Config{}, &mo)
	err := p.Parse(args)

	switch {
	case err == arg.ErrHelp:
		p.WriteHelp(os.Stdout)
		os.Exit(0)
	case err == arg.ErrVersion:
		fmt.Println(mo.Version())
		os.Exit(0)
	case err != nil:
		return &mo, nil, err
	}

	// parse status-as option
	maps, err := checkers.ParseStatusMap(mo.StatusAs)
	if err != nil {
		return &mo, nil, err
	}

	// Set internal limit: 24h1m
	maxMinute := uint(60*24 + 1)
	if mo.Critical < 1 || mo.Critical > maxMinute || mo.Warning < 1 || mo.Warning > maxMinute {
		err = fmt.Errorf("specified minute is out of range (1-%d)", maxMinute)
	}
	if mo.Host != "" && mo.Service != "" {
		err = fmt.Errorf("both --host and --service cannot be specified")
	}
	if mo.Host == "" && mo.Service == "" {
		err = fmt.Errorf("either --host or --service is required")
	}
	if mo.Critical < mo.Warning {
		err = fmt.Errorf("critical minute must be equal or greater than warning minute")
	}
	return &mo, maps, err
}

func (opts *mackerelMetricOpts) run() *checkers.Checker {
	apikey := os.Getenv("MACKEREL_APIKEY")

	conf, err := config.LoadConfig(config.DefaultConfig.Conffile)
	if err != nil {
		if runtime.GOOS == "windows" {
			newpath := filepath.Join(config.DefaultConfig.Conffile, "../../../mackerel-agent.conf")
			conf, err = config.LoadConfig(newpath)
			if err != nil {
				return checkers.Unknown(err.Error())
			}
			conf.Conffile = newpath
			conf.Root = filepath.Dir(newpath)
		} else {
			return checkers.Unknown(err.Error())
		}
	}
	apibase := conf.Apibase
	if apikey == "" {
		apikey = conf.Apikey
	}
	if apibase == "" || apikey == "" {
		return checkers.Unknown("Not found apibase or apikey in " + config.DefaultConfig.Conffile)
	}

	to := time.Now().Unix()
	criticalFrom := to - int64(opts.Critical*60)
	warningFrom := to - int64(opts.Warning*60)

	client, err := mackerel.NewClientWithOptions(apikey, apibase, false)
	if err != nil {
		return checkers.Unknown(err.Error())
	}

	return checkMetric(client, opts, criticalFrom, warningFrom, to)
}

func checkMetric(client *mackerel.Client, opts *mackerelMetricOpts, criticalFrom int64, warningFrom int64, to int64) *checkers.Checker {
	var metricValue []mackerel.MetricValue

	// CRITICAL check
	metricValue, err := fetchMetricValues(client, opts.Host, opts.Service, opts.Metric, criticalFrom, to)
	if err != nil {
		return checkers.Unknown(err.Error())
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

func fetchMetricValues(client *mackerel.Client, hostID string, serviceName string, metricName string, from int64, to int64) ([]mackerel.MetricValue, error) {
	if hostID != "" {
		return client.FetchHostMetricValues(hostID, metricName, from, to)
	} else {
		return client.FetchServiceMetricValues(serviceName, metricName, from, to)
	}
}
