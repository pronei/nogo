package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	rateClient "github.com/pronei/nogo/client"
	"github.com/pronei/nogo/internal/enums"
	structs "github.com/pronei/nogo/shared"
	"go.uber.org/zap"
)

type RequestsTest struct {
	RequestMap map[string]structs.LimitRequest `json:"requestMap"`
}

var config *structs.RateLimiterConfig
var ruleFileName, requestsFileName string

func init() {
	config = &structs.RateLimiterConfig{
		StorageType: enums.InMemoryStorage,
		StrategyConfig: structs.StrategyConfig{
			Type:     "rolling_window",
			TimeUnit: "ns",
		},
		RedisConfig: structs.RedisConfig{
			Host:                      "localhost:6379",
			Password:                  "",
			ConnectionTimeoutInMillis: 10000,
			ReadTimeoutInMillis:       10000,
			WriteTimeoutInMillis:      10000,
			PoolSize:                  300,
			DB:                        1,
		},
		InMemoryConfig: structs.InMemoryConfig{
			Expiration:      structs.Duration(time.Duration(0)),
			CleanupInterval: structs.Duration(time.Duration(0)),
		},
	}
	ruleFileName = "rules.json"
	requestsFileName = "requests2.json"
}

func main() {
	ruleBytes, err := ioutil.ReadFile(ruleFileName)
	if err != nil {
		log.Fatalf("cannot read rule file - %s\n", err.Error())
	}

	rules := &structs.RuleImport{}
	if err := json.Unmarshal(ruleBytes, rules); err != nil {
		log.Fatalf("cannot unmarshal rules - %v\n", err.Error())
	}

	logger, _ := zap.NewProduction()
	client, err := rateClient.Create(logger.Sugar(), config, rules)
	if err != nil {
		log.Fatalf("cannot create client - %s\n", err.Error())
	}

	reqBytes, err := ioutil.ReadFile(requestsFileName)
	if err != nil {
		log.Fatalf("cannot read requests main file - %s\n", err.Error())
	}

	requests := &RequestsTest{}
	if err := json.Unmarshal(reqBytes, requests); err != nil {
		log.Fatalf("cannot unmarshal request tests - %s\n", err.Error())
	}

	ctx := context.TODO()

	fmt.Printf("request ID\tallowed?\n")
	for reqId, req := range requests.RequestMap {
		result, err := client.AllowAndUpdate(ctx, &req)
		if err != nil {
			fmt.Printf("%s\t%v\n", reqId, err.Error())
			continue
		}
		fmt.Printf("%s\t%v\n", reqId, result)
	}

	symbols := []rune{'|', '/', '-', '\\'}
	fmt.Println("trying again after 30sec pause...")
	ticker := time.NewTicker(time.Second * 1)
	done := make(chan bool)
	i := 0
	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				fmt.Printf("\r%c", symbols[i])
				i = (i + 1) % len(symbols)
			}
		}
	}()
	time.Sleep(time.Second * 30)
	ticker.Stop()
	done <- true
	fmt.Println()

	for reqId, req := range requests.RequestMap {
		result, err := client.AllowAndUpdate(ctx, &req)
		if err != nil {
			fmt.Printf("%s\t%v\n", reqId, err.Error())
			continue
		}
		fmt.Printf("%s\t%v\n", reqId, result)
	}
}
