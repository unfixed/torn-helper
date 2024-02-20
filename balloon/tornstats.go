package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
)

type SpyUserResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Compare struct {
		Status  bool   `json:"status"`
		Message string `json:"message"`
	} `json:"compare"`
	Spy struct {
		Type               string  `json:"type"`
		Status             bool    `json:"status"`
		Message            string  `json:"message"`
		PlayerName         string  `json:"player_name"`
		PlayerID           string  `json:"player_id"`
		PlayerLevel        int     `json:"player_level"`
		PlayerFaction      string  `json:"player_faction"`
		TargetScore        float64 `json:"target_score"`
		YourScore          float64 `json:"your_score"`
		FairFightBonus     int     `json:"fair_fight_bonus"`
		Difference         string  `json:"difference"`
		Timestamp          int     `json:"timestamp"`
		Strength           int64   `json:"strength"`
		DeltaStrength      int64   `json:"deltaStrength"`
		StrengthTimestamp  int     `json:"strength_timestamp"`
		Defense            int64   `json:"defense"`
		DeltaDefense       int64   `json:"deltaDefense"`
		DefenseTimestamp   int     `json:"defense_timestamp"`
		Speed              int64   `json:"speed"`
		DeltaSpeed         int64   `json:"deltaSpeed"`
		SpeedTimestamp     int     `json:"speed_timestamp"`
		Dexterity          int64   `json:"dexterity"`
		DeltaDexterity     int64   `json:"deltaDexterity"`
		DexterityTimestamp int     `json:"dexterity_timestamp"`
		Total              int64   `json:"total"`
		DeltaTotal         int64   `json:"deltaTotal"`
		TotalTimestamp     int     `json:"total_timestamp"`
	} `json:"spy"`
	Attacks struct {
		Status  bool   `json:"status"`
		Message string `json:"message"`
	} `json:"attacks"`
}

func (s SpyUserResponse) MarshalBinary() ([]byte, error) {
	return json.Marshal(s)
}
func getSpyReport(userId int) SpyUserResponse {

	apiKey, ok := os.LookupEnv("tornStatsApiKey")
	if !ok {
		fmt.Println("tornStatsApiKey missing")
		os.Exit(1)
	}

	var spyReport SpyUserResponse

	//check redis cache
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	result, err := rdb.Get(ctx, fmt.Sprintf("spyreport_%d", userId)).Result()
	if err == redis.Nil {
		fmt.Println(fmt.Sprintf("get spy report for %d", userId))
		URL := fmt.Sprintf("https://www.tornstats.com/api/v2/%s/spy/user/%d", apiKey, userId)
		time.Sleep(1 * time.Second)
		response, err := http.Get(URL)
		if err != nil {
			log.Fatalln(err)
		}
		responseBody, err := io.ReadAll(response.Body)
		if err != nil {
			log.Fatalln(err)
		}
		json.Unmarshal(responseBody, &spyReport)

		defer response.Body.Close()
		time.Sleep(1 * time.Second)
		//push to cache
		err = rdb.Set(ctx, fmt.Sprintf("spyreport_%d", userId), spyReport, time.Duration(rand.Intn(60)+10080)*time.Minute).Err()
		if err != nil {
			panic(err)
		}

		return spyReport
	}
	//found in cache
	json.Unmarshal([]byte(result), &spyReport)
	return spyReport

}
