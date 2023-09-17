package checkmackerelmetric

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseArgs_success(t *testing.T) {
	_, err := parseArgs(strings.Split("-H HOSTID -n METRIC -w 30 -c 60", " "))
	assert.Equal(t, err, nil, "err should be nil")

	_, err = parseArgs(strings.Split("-s SERVICE -n METRIC -w 30 -c 60", " "))
	assert.Equal(t, err, nil, "err should be nil")

	_, err = parseArgs(strings.Split("-H HOSTID -n METRIC -w 1441 -c 60", " "))
	assert.Equal(t, err, nil, "err should be nil")

	_, err = parseArgs(strings.Split("-H HOSTID -n METRIC -w 30 -c 1441", " "))
	assert.Equal(t, err, nil, "err should be nil")
}

func TestParseArgs_fail(t *testing.T) {
	_, err := parseArgs(strings.Split("-H HOSTID", " "))
	assert.Equal(t, err, fmt.Errorf("--name is required"), "needs metric name")

	_, err = parseArgs(strings.Split("-H HOSTID -w 30", " "))
	assert.Equal(t, err, fmt.Errorf("--name is required"), "needs metric name")

	_, err = parseArgs(strings.Split("-H HOSTID -n METRIC", " "))
	assert.Equal(t, err, fmt.Errorf("--warning is required"), "needs warning metric")

	_, err = parseArgs(strings.Split("-H HOSTID -n METRIC -w 30", " "))
	assert.Equal(t, err, fmt.Errorf("--critical is required"), "needs critical metric")

	_, err = parseArgs(strings.Split("-H HOSTID -n METRIC -c 30", " "))
	assert.Equal(t, err, fmt.Errorf("--warning is required"), "needs warning metric")

	_, err = parseArgs(strings.Split("-n METRIC -c 30 -w 60", " "))
	assert.Equal(t, err, fmt.Errorf("either --host or --service is required"), "needs host or service")

	_, err = parseArgs(strings.Split("-s SERVICE", " "))
	assert.Equal(t, err, fmt.Errorf("--name is required"), "needs metric name")

	_, err = parseArgs(strings.Split("-s SERVICE -w 30", " "))
	assert.Equal(t, err, fmt.Errorf("--name is required"), "needs metric name")

	_, err = parseArgs(strings.Split("-s SERVICE -n METRIC", " "))
	assert.Equal(t, err, fmt.Errorf("--warning is required"), "needs warning metric")

	_, err = parseArgs(strings.Split("-s SERVICE -n METRIC -w 30", " "))
	assert.Equal(t, err, fmt.Errorf("--critical is required"), "needs critical metric")

	_, err = parseArgs(strings.Split("-s SERVICE -n METRIC -c 30", " "))
	assert.Equal(t, err, fmt.Errorf("--warning is required"), "needs warning metric")

	_, err = parseArgs(strings.Split("-H HOSTID -s SERVICE -n METRIC -w 30 -c 60", " "))
	assert.Equal(t, err, fmt.Errorf("both --host and --service cannot be specified"), "one of host or service")

	_, err = parseArgs(strings.Split("-H HOSTID -n METRIC -w 0 -c 60", " "))
	assert.Equal(t, err, fmt.Errorf("specified minute is out of range (1-1441)"), "0 minute is invalid")

	_, err = parseArgs(strings.Split("-H HOSTID -n METRIC -w \"-10\" -c 60", " "))
	assert.Equal(t, err, fmt.Errorf("error processing -w: strconv.ParseUint: parsing \"\\\"-10\\\"\": invalid syntax"), "negative minute is invalid")

	_, err = parseArgs(strings.Split("-H HOSTID -n METRIC -w 1442 -c 60", " "))
	assert.Equal(t, err, fmt.Errorf("specified minute is out of range (1-1441)"), "over 1441 minute is invalid")

	_, err = parseArgs(strings.Split("-H HOSTID -n METRIC -w 30 -c 0", " "))
	assert.Equal(t, err, fmt.Errorf("specified minute is out of range (1-1441)"), "0 minute is invalid")

	_, err = parseArgs(strings.Split("-H HOSTID -n METRIC -w 30 -c \"-10\"", " "))
	assert.Equal(t, err, fmt.Errorf("error processing -c: strconv.ParseUint: parsing \"\\\"-10\\\"\": invalid syntax"), "negative minute is invalid")

	_, err = parseArgs(strings.Split("-H HOSTID -n METRIC -w 30 -c 1442", " "))
	assert.Equal(t, err, fmt.Errorf("specified minute is out of range (1-1441)"), "over 1441 minute is invalid")

	_, err = parseArgs(strings.Split("-H HOSTID -n METRIC -w true -c 60", " "))
	assert.Equal(t, err, fmt.Errorf("error processing -w: strconv.ParseUint: parsing \"true\": invalid syntax"), "string minute is invalid")

	_, err = parseArgs(strings.Split("-H HOSTID -n METRIC -w 30 -c true", " "))
	assert.Equal(t, err, fmt.Errorf("error processing -c: strconv.ParseUint: parsing \"true\": invalid syntax"), "string minute is invalid")
}
