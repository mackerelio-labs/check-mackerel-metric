package checkmackerelmetric

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/mackerelio/checkers"
	"github.com/mackerelio/mackerel-client-go"
	"github.com/stretchr/testify/assert"
)

func TestParseArgs(t *testing.T) {
	opts, err := parseArgs(strings.Split("-H HOSTID -n METRIC -w 30 -c 60", " "))
	assert.Equal(t, nil, err, "parameters with a host should be passed")
	assert.Equal(t, "HOSTID", opts.Host, "host is passed correctly")
	assert.Equal(t, "METRIC", opts.Metric, "metric is passed correctly")
	assert.Equal(t, uint(30), opts.Warning, "warning is passed correctly")
	assert.Equal(t, uint(60), opts.Critical, "critical is passed correctly")

	opts, err = parseArgs(strings.Split("-s SERVICE -n METRIC -w 30 -c 60", " "))
	assert.Equal(t, nil, err, "parameters with a service should be passed")
	assert.Equal(t, "SERVICE", opts.Service, "service is passed correctly")

	opts, err = parseArgs(strings.Split("-H HOSTID -n METRIC -w 30 -c 1441", " "))
	assert.Equal(t, nil, err, "parmeters with max minute should be passed")
	assert.Equal(t, uint(1441), opts.Critical, "1441 minutes (= max minute) is passed correctly")

	_, err = parseArgs(strings.Split("-H HOSTID -n METRIC -w 60 -c 60", " "))
	assert.Equal(t, nil, err, "it is acceptable for warning and critical to have the same value")

	_, err = parseArgs(strings.Split("-H HOSTID -n METRIC -w 1441 -c 60", " "))
	assert.Equal(t, fmt.Errorf("critical minute must be equal or greater than warning minute"), err, "warning can't over critical")

	_, err = parseArgs(strings.Split("-H HOSTID", " "))
	assert.Equal(t, fmt.Errorf("--name is required"), err, "needs metric name")

	_, err = parseArgs(strings.Split("-H HOSTID -w 30", " "))
	assert.Equal(t, fmt.Errorf("--name is required"), err, "needs metric name")

	_, err = parseArgs(strings.Split("-H HOSTID -n METRIC", " "))
	assert.Equal(t, fmt.Errorf("--warning is required"), err, "needs warning metric")

	_, err = parseArgs(strings.Split("-H HOSTID -n METRIC -w 30", " "))
	assert.Equal(t, fmt.Errorf("--critical is required"), err, "needs critical metric")

	_, err = parseArgs(strings.Split("-H HOSTID -n METRIC -c 30", " "))
	assert.Equal(t, fmt.Errorf("--warning is required"), err, "needs warning metric")

	_, err = parseArgs(strings.Split("-n METRIC -c 60 -w 30", " "))
	assert.Equal(t, fmt.Errorf("either --host or --service is required"), err, "needs host or service")

	_, err = parseArgs(strings.Split("-s SERVICE", " "))
	assert.Equal(t, fmt.Errorf("--name is required"), err, "needs metric name")

	_, err = parseArgs(strings.Split("-s SERVICE -w 30", " "))
	assert.Equal(t, fmt.Errorf("--name is required"), err, "needs metric name")

	_, err = parseArgs(strings.Split("-s SERVICE -n METRIC", " "))
	assert.Equal(t, fmt.Errorf("--warning is required"), err, "needs warning metric")

	_, err = parseArgs(strings.Split("-s SERVICE -n METRIC -w 30", " "))
	assert.Equal(t, fmt.Errorf("--critical is required"), err, "needs critical metric")

	_, err = parseArgs(strings.Split("-s SERVICE -n METRIC -c 30", " "))
	assert.Equal(t, fmt.Errorf("--warning is required"), err, "needs warning metric")

	_, err = parseArgs(strings.Split("-H HOSTID -s SERVICE -n METRIC -w 30 -c 60", " "))
	assert.Equal(t, fmt.Errorf("both --host and --service cannot be specified"), err, "one of host or service")

	_, err = parseArgs(strings.Split("-H HOSTID -n METRIC -w 0 -c 60", " "))
	assert.Equal(t, fmt.Errorf("specified minute is out of range (1-1441)"), err, "0 minute is invalid")

	_, err = parseArgs(strings.Split("-H HOSTID -n METRIC -w \"-10\" -c 60", " "))
	assert.Equal(t, fmt.Errorf("error processing -w: strconv.ParseUint: parsing \"\\\"-10\\\"\": invalid syntax"), err, "negative minute is invalid")

	_, err = parseArgs(strings.Split("-H HOSTID -n METRIC -w 30.1 -c 60", " "))
	assert.Equal(t, fmt.Errorf("error processing -w: strconv.ParseUint: parsing \"30.1\": invalid syntax"), err, "float minute is invalid")

	_, err = parseArgs(strings.Split("-H HOSTID -n METRIC -w 30 -c \"-10\"", " "))
	assert.Equal(t, fmt.Errorf("error processing -c: strconv.ParseUint: parsing \"\\\"-10\\\"\": invalid syntax"), err, "negative minute is invalid")

	_, err = parseArgs(strings.Split("-H HOSTID -n METRIC -w 30 -c 60.1", " "))
	assert.Equal(t, fmt.Errorf("error processing -c: strconv.ParseUint: parsing \"60.1\": invalid syntax"), err, "float minute is invalid")

	_, err = parseArgs(strings.Split("-H HOSTID -n METRIC -w 30 -c 1442", " "))
	assert.Equal(t, fmt.Errorf("specified minute is out of range (1-1441)"), err, "over 1441 minute is invalid")

	_, err = parseArgs(strings.Split("-H HOSTID -n METRIC -w true -c 60", " "))
	assert.Equal(t, fmt.Errorf("error processing -w: strconv.ParseUint: parsing \"true\": invalid syntax"), err, "string minute is invalid")

	_, err = parseArgs(strings.Split("-H HOSTID -n METRIC -w 30 -c true", " "))
	assert.Equal(t, fmt.Errorf("error processing -c: strconv.ParseUint: parsing \"true\": invalid syntax"), err, "string minute is invalid")
}

func TestCheckMetric(t *testing.T) {
	to := int64(1695006000) // 2023-09-18 12:00:00 +09:00

	ts := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		var respJSON []byte

		query := req.URL.Query()
		if !reflect.DeepEqual(query["name"], []string{"my.metric"}) {
			res.WriteHeader(404)
			respJSON, _ = json.Marshal(map[string]map[string]string{
				"error": {"message": "metric not found"},
			})
		} else {
			switch req.URL.Path {
			case "/api/v0/hosts/test1/metrics", "/api/v0/services/test1/metrics":
				metrics := map[string][]mackerel.MetricValue{
					"metrics": {
						{Time: to - 60*60},
						{Time: to - 40*60},
						{Time: to - 20*60},
					},
				}
				respJSON, _ = json.Marshal(metrics)
			case "/api/v0/hosts/test2/metrics", "/api/v0/services/test2/metrics":
				metrics := map[string][]mackerel.MetricValue{
					"metrics": {},
				}
				respJSON, _ = json.Marshal(metrics)
			case "/api/v0/hosts/test3/metrics", "/api/v0/services/test3/metrics":
				metrics := map[string][]mackerel.MetricValue{
					"metrics": {
						{Time: to - 40*60},
					},
				}
				respJSON, _ = json.Marshal(metrics)
			case "/api/v0/hosts/nohost/metrics":
				res.WriteHeader(404)
				respJSON, _ = json.Marshal(map[string]map[string]string{
					"error": {"message": "host not found"},
				})
			case "/api/v0/services/noservice/metrics":
				res.WriteHeader(404)
				respJSON, _ = json.Marshal(map[string]map[string]string{
					"error": {"message": "service not found"},
				})
			}
		}
		res.Header()["Content-Type"] = []string{"application/json"}
		fmt.Fprint(res, string(respJSON))
	}))
	defer ts.Close()

	client, _ := mackerel.NewClientWithOptions("dummy-key", ts.URL, false)
	var mo = &mackerelMetricOpts{}

	mo.Host = "test1"
	mo.Metric = "my.metric"
	criticalFrom := to - int64(60*60) // 60 minutes before, 2023-09-18 11:00:00 +0900
	warningFrom := to - int64(30*60)  // 30 minutes before, 2023-09-18 11:30:00 +0900

	checker := checkMetric(client, mo, criticalFrom, warningFrom, to)
	assert.Equal(t, checkers.OK, checker.Status, "metric exists")

	checker = checkMetric(client, mo, criticalFrom, warningFrom+15*60, to) // 15 minutes before
	assert.Equal(t, checkers.WARNING, checker.Status, "some metrics returned, but no metric for warning range (by modifing warningFrom)")

	mo.Host = "test2"
	checker = checkMetric(client, mo, criticalFrom, warningFrom, to)
	assert.Equal(t, checkers.CRITICAL, checker.Status, "no metric so critical")

	mo.Host = "test3"
	checker = checkMetric(client, mo, criticalFrom, warningFrom, to)
	assert.Equal(t, checkers.WARNING, checker.Status, "some metrics returned, but no metric for warning range")

	mo.Host = "nohost"
	checker = checkMetric(client, mo, criticalFrom, warningFrom, to)
	assert.Equal(t, checkers.UNKNOWN, checker.Status, "no host")

	mo.Host = "test1"
	mo.Metric = "nometric"
	checker = checkMetric(client, mo, criticalFrom, warningFrom, to)
	assert.Equal(t, checkers.UNKNOWN, checker.Status, "no metric for host")

	mo.Host = ""
	mo.Service = "test1"
	mo.Metric = "my.metric"
	checker = checkMetric(client, mo, criticalFrom, warningFrom, to)
	assert.Equal(t, checkers.OK, checker.Status, "metric exists")

	checker = checkMetric(client, mo, criticalFrom, warningFrom+15*60, to)
	assert.Equal(t, checkers.WARNING, checker.Status, "some metrics returned, but no metric for warning range (by modifing warningFrom)")

	mo.Service = "test2"
	checker = checkMetric(client, mo, criticalFrom, warningFrom, to)
	assert.Equal(t, checkers.CRITICAL, checker.Status, "no metric so critical")

	mo.Service = "test3"
	checker = checkMetric(client, mo, criticalFrom, warningFrom, to)
	assert.Equal(t, checkers.WARNING, checker.Status, "some metrics returned, but no metric for warning range")

	mo.Service = "noservice"
	checker = checkMetric(client, mo, criticalFrom, warningFrom, to)
	assert.Equal(t, checkers.UNKNOWN, checker.Status, "no service")

	mo.Service = "test1"
	mo.Metric = "nometric"
	checker = checkMetric(client, mo, criticalFrom, warningFrom, to)
	assert.Equal(t, checkers.UNKNOWN, checker.Status, "no metric for service")
}
