package main

import (
	"github.com/yfujita/slackutil"
	"os/exec"
	"encoding/json"
	"fmt"
	"strconv"
	"flag"
	"time"
	"sort"
)

const (
 	REGION = "us-east-1"
	TIME_FORMAT = "2006-01-02T15:04:05Z"
)

type Datapoint struct {
	Timestamp		string
	Maximum			float64
	Unit 			string
}

type Datapoints []Datapoint

func (p Datapoints) Len() int {
	return len(p)
}

func (p Datapoints) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (p Datapoints) Less(i, j int) bool {
	iTime, _ := time.Parse(TIME_FORMAT, p[i].Timestamp)
	jTime, _ := time.Parse(TIME_FORMAT, p[j].Timestamp)
	return iTime.After(jTime)
}


type Resp struct {
	Datapoints Datapoints
}

func main() {
	var slackUrl string
	var slackChannel string
	var slackBotName string
	var slackBotIcon string

	flag.StringVar(&slackUrl, "slackUrl", "blank", "config file path")
	flag.StringVar(&slackChannel, "slackChannel", "#bot_test", "slack channel")
	flag.StringVar(&slackBotName, "slackBotName", "aws-billing", "slack bot name")
	flag.StringVar(&slackBotIcon, "slackBotIcon", ":ghost:", "slack bot name")

	flag.Parse()
	if slackUrl == "blank" {
		panic("Invalid url parameter")
	}

	endTime := time.Now().Format(TIME_FORMAT)
	startTime := time.Now().AddDate(0, 0, -1).Format(TIME_FORMAT);

	dataPoints := getMetricStatistics(startTime, endTime)
	if len(dataPoints) == 0 {
		panic("Could not get datapoints.")
	}
	billing := dataPoints[0].Maximum
	title := "今月のAWS利用料金: $" + strconv.FormatFloat(billing, 'g', 5, 64)

	bot := slackutil.NewBot(slackUrl, slackChannel, slackBotName, slackBotIcon)

	fmt.Println("Send message. " + title + "\n" + "");
	err := bot.Message(title, "")
	if err != nil {
		panic(err.Error())
	}
}

func getMetricStatistics(startTime, endTime string) []Datapoint {
	jsonStr := executeCmd("aws", "cloudwatch", "get-metric-statistics", "--region", REGION, "--namespace", "AWS/Billing", "--metric-name",
		"EstimatedCharges", "--start-time", startTime, "--end-time", endTime, "--period", "180", "--statistics", "Maximum",
		"--dimensions", "Name=Currency,Value=USD")
	fmt.Println(jsonStr)
	var resp Resp
	json.Unmarshal([]byte(jsonStr), &resp)
	sort.Sort(resp.Datapoints)
	return resp.Datapoints
}

func executeCmd(name string, args ...string) string {
	cmd := exec.Command(name, args...)
	out, err := cmd.Output()
	if err != nil {
		panic(err.Error())
	}
	return string(out)
}