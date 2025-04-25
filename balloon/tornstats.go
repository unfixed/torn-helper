package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
)

//https://www.tornstats.com/api/v2/TS_VeflK7cHcYnd1272/spy/faction/8062
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

type TornStatsMember struct {
	Name          string     `json:"name"`
	Level         int        `json:"level"`
	DaysInFaction int        `json:"days_in_faction"`
	LastAction    LastAction `json:"last_action"`
	Status        Status     `json:"status"`
	Position      string     `json:"position"`
	ID            int        `json:"id"`
	Spy           struct {
		Strength  int   `json:"strength"`
		Defense   int   `json:"defense"`
		Speed     int   `json:"speed"`
		Dexterity int   `json:"dexterity"`
		Total     int64 `json:"total"`
		Timestamp int   `json:"timestamp"`
	} `json:"spy"`
}

type SpyFactionResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Faction struct {
		ID         int                     `json:"ID"`
		Name       string                  `json:"name"`
		RankedWars map[int]RankedWar       `json:"ranked_wars"`
		Members    map[int]TornStatsMember `json:"members"`
	} `json:"faction"`
}

func (s SpyUserResponse) MarshalBinary() ([]byte, error) {
	return json.Marshal(s)
}

func (s TornStatsMember) MarshalBinary() ([]byte, error) {
	return json.Marshal(s)
}

func getSpyReport(userId int) SpyUserResponse {

	// apiKey, ok := os.LookupEnv("tornStatsApiKey")
	// if !ok {
	// 	fmt.Println("tornStatsApiKey missing")
	// 	os.Exit(1)
	// }

	var spyReport SpyUserResponse

	//check redis cache
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	defer rdb.Close()

	result, err := rdb.Get(ctx, fmt.Sprintf("spyreport_%d", userId)).Result()
	if err == redis.Nil {
		// fmt.Printf("get spy report for %d\n", userId)
		// URL := fmt.Sprintf("https://www.tornstats.com/api/v2/%s/spy/user/%d", apiKey, userId)
		// time.Sleep(1 * time.Second)
		// response, err := http.Get(URL)
		// if err != nil {
		// 	log.Println(err)
		// 	return spyReport
		// }
		// responseBody, err := io.ReadAll(response.Body)
		// if err != nil {
		// 	log.Println(err)
		// 	return spyReport
		// }
		// json.Unmarshal(responseBody, &spyReport)
		// defer response.Body.Close()

		// if spyReport.Status {
		// 	//push to cache
		// 	err = rdb.Set(ctx, fmt.Sprintf("spyreport_%d", userId), spyReport, time.Duration(1440)*time.Minute).Err()
		// 	if err != nil {
		// 		panic(err)
		// 	}
		// }

		return spyReport
	}
	//found in cache
	json.Unmarshal([]byte(result), &spyReport)
	return spyReport

}

func getFactionSpyReport(factionId string) {

	fmt.Printf("running getFactionSpyReport:%s\n", factionId)

	logFile, err := os.OpenFile("getFactionSpyReport.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	defer rdb.Close()

	if factionId == "" {
		fmt.Println("------------------")
		opponent, err_redisget_WarOpponent := rdb.Get(ctx, "warOpponent").Result()
		if err_redisget_WarOpponent == redis.Nil {
			log.Println(err_redisget_WarOpponent)
			factionId = "8062"
		} else {
			log.Printf("got opponent factionId as %s", opponent)
			factionId = opponent
		}

	}

	apiKey, ok := os.LookupEnv("tornStatsApiKey")
	if !ok {
		fmt.Println("tornStatsApiKey missing")
		os.Exit(1)
	}

	var spyReport SpyFactionResponse

	log.Printf("getting spy report for faction %s\n", factionId)
	URL := fmt.Sprintf("https://www.tornstats.com/api/v2/%s/spy/faction/%s", apiKey, factionId)
	time.Sleep(1 * time.Second)
	response, err := http.Get(URL)
	if err != nil {
		log.Println(err)
		return
	}
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		log.Println(err)
		return
	}
	json.Unmarshal(responseBody, &spyReport)
	defer response.Body.Close()

	if spyReport.Status {

		for _, member := range spyReport.Faction.Members {
			//push to cache

			fmt.Printf("%v \n", member)

			err = rdb.Set(ctx, fmt.Sprintf("spyreport_%d", member.ID), member, time.Duration(1440)*time.Minute).Err()
			if err != nil {
				panic(err)
			}

		}

	}

}
