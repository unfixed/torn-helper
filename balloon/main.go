package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var ctx = context.Background()

type ApiKey struct {
	key   string
	valid bool
}

type FactionBasicInfo struct {
	ID            int               `json:"ID"`
	Name          string            `json:"name"`
	Tag           string            `json:"tag"`
	TagImage      string            `json:"tag_image"`
	Leader        int               `json:"leader"`
	CoLeader      int               `json:"co-leader"`
	Respect       int               `json:"respect"`
	Age           int               `json:"age"`
	Capacity      int               `json:"capacity"`
	BestChain     int               `json:"best_chain"`
	TerritoryWars TerritoryWars     `json:"territory_wars"`
	RaidWars      RaidWars          `json:"raid_wars"`
	Peace         Peace             `json:"peace"`
	Rank          Rank              `json:"rank"`
	RankedWars    map[int]RankedWar `json:"ranked_wars"`
	Members       map[int]Member    `json:"members"`
}
type TerritoryWars struct {
}
type RaidWars struct {
}
type Peace struct {
}
type Rank struct {
	Level    int    `json:"level"`
	Name     string `json:"name"`
	Division int    `json:"division"`
	Position int    `json:"position"`
	Wins     int    `json:"wins"`
}
type Faction struct {
	Name  string `json:"name"`
	Score int    `json:"score"`
	Chain int    `json:"chain"`
}

type War struct {
	Start  int `json:"start"`
	End    int `json:"end"`
	Target int `json:"target"`
	Winner int `json:"winner"`
}
type RankedWar struct {
	Factions map[int]Faction `json:"factions"`
	War      War             `json:"war"`
}

type LastAction struct {
	Status    string `json:"status"`
	Timestamp int    `json:"timestamp"`
	Relative  string `json:"relative"`
}
type Status struct {
	Description string `json:"description"`
	Details     string `json:"details"`
	State       string `json:"state"`
	Color       string `json:"color"`
	Until       int    `json:"until"`
}
type Member struct {
	Name          string     `json:"name"`
	Level         int        `json:"level"`
	DaysInFaction int        `json:"days_in_faction"`
	LastAction    LastAction `json:"last_action"`
	Status        Status     `json:"status"`
	Position      string     `json:"position"`
}

type FactionMember struct {
	UserId             int
	Name               string
	Level              int
	LastStatus         string
	LastStatusRaw      int
	LastSeen           string
	LastSeenTimestamp  int
	Status             string
	StatusRaw          int
	Status_Long        string
	BattleStats        string
	BattleStatsRaw     int64
	BattleStats_Str    string
	BattleStats_StrRaw int64
	BattleStats_Def    string
	BattleStats_DefRaw int64
	BattleStats_Dex    string
	BattleStats_DexRaw int64
	BattleStats_Spd    string
	BattleStats_SpdRaw int64
}
type FactionMembers struct {
	Members []int
}

func getFactionBasicInfo(factionId string, apiKey string) {
	// function queries torn api and gets basic info on faction
	// from there it looks for ongoing ranked wars and calls
	// functions to update info on war opponent and it's members

	fmt.Println("running FactionBasicInfo")
	fmt.Printf("Current date and time is: %s\n", time.Now().String())

	URL := fmt.Sprintf("https://api.torn.com/faction/%s?selections=basic&key=%s", factionId, apiKey)
	response, err_getUrl := http.Get(URL)
	if err_getUrl != nil {
		log.Fatalln(err_getUrl)
	}

	responseBody, err_read_responseBody := io.ReadAll(response.Body)
	if err_read_responseBody != nil {
		log.Fatalln(err_read_responseBody)
	}

	var data FactionBasicInfo
	json.Unmarshal(responseBody, &data)

	for _, rankedwar := range data.RankedWars {
		for facId, _ := range rankedwar.Factions {
			if fmt.Sprint(facId) != factionId {
				updateWar(facId, rankedwar.War.Start)
				getFactionMembers(fmt.Sprint(facId), apiKey)
			}
		}
	}

	// clean up memory after execution
	defer response.Body.Close()
}

func updateWar(warOpponent int, warStart int) {
	//function updates War Opponent and War Start Time info.
	fmt.Println("updateWar running")

	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	defer rdb.Close()

	startTime, err_redisset_WarStartTime := rdb.Get(ctx, "warStartTime").Result()
	if err_redisset_WarStartTime == redis.Nil {
		fmt.Println("warStartTime missing, setting value...")
		err_redisset_WarStartTime := rdb.Set(ctx, "warStartTime", fmt.Sprintf("%d", warStart), 123*time.Hour).Err()
		if err_redisset_WarStartTime != nil {
			panic(err_redisset_WarStartTime)
		}
	}
	if startTime < fmt.Sprintf("%d", warStart) {
		fmt.Println("warStartTime is wrong, updating value...")
		err_redisset_WarStartTime := rdb.Set(ctx, "warStartTime", fmt.Sprintf("%d", warStart), 123*time.Hour).Err()
		if err_redisset_WarStartTime != nil {
			panic(err_redisset_WarStartTime)
		}
	}
	opponent, err_redisget_WarOpponent := rdb.Get(ctx, "warOpponent").Result()
	if err_redisget_WarOpponent == redis.Nil {
		err_redisset_WarOpponent := rdb.Set(ctx, "warOpponent", fmt.Sprintf("%d", warOpponent), 123*time.Hour).Err()
		if err_redisset_WarOpponent != nil {
			panic(err_redisset_WarOpponent)
		}
	}
	if opponent != fmt.Sprintf("%d", warOpponent) {
		fmt.Println("warOpponent is wrong, updating value...")
		err_redisset_WarOpponent := rdb.Set(ctx, "warOpponent", fmt.Sprintf("%d", warOpponent), 123*time.Hour).Err()
		if err_redisset_WarOpponent != nil {
			panic(err_redisset_WarOpponent)
		}
	}
}

func getFactionMembers(factionId string, apiKey string) {

	URL := fmt.Sprintf("https://api.torn.com/faction/%s?selections=basic&key=%s", factionId, apiKey)
	// fmt.Println(URL)
	response, err_getUrl := http.Get(URL)
	if err_getUrl != nil {
		log.Fatalln(err_getUrl)
	}

	responseBody, err_read_responseBody := io.ReadAll(response.Body)
	if err_read_responseBody != nil {
		log.Fatalln(err_read_responseBody)
	}

	var data FactionBasicInfo
	json.Unmarshal(responseBody, &data)

	var members []int
	for i, m := range data.Members {
		members = append(members, i)
		// updateMember(factionId, i, m)
		// personalStats :=
		updateMemberRedis(i, m, getSpyReport(i))
	}
	updateFactionRedis(factionId, members)

	// clean up memory after execution
	defer response.Body.Close()
}
func (f FactionMembers) MarshalBinary() ([]byte, error) {
	return json.Marshal(f)
}
func updateFactionRedis(factionId string, members []int) {

	var factionMembers FactionMembers
	factionMembers.Members = members

	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	defer rdb.Close()

	err_redisset_factionMembers := rdb.Set(ctx, factionId, factionMembers, 0).Err()
	if err_redisset_factionMembers != nil {
		panic(err_redisset_factionMembers)
	}
	
	err_redisset_fallbackfactionMembers := rdb.Set(ctx, fmt.Sprintf("fallback_%s", factionId), factionMembers, 0).Err()
	if err_redisset_fallbackfactionMembers != nil {
		panic(err_redisset_fallbackfactionMembers)
	}
}

func (f FactionMember) MarshalBinary() ([]byte, error) {
	return json.Marshal(f)
}
func updateMemberRedis(userid int, member Member, spyReport SpyUserResponse) {

	evalLastStatus := map[string]int{"Offline": 0, "Idle": 1, "Online": 2}

	var facMember FactionMember
	facMember.UserId = userid
	facMember.Name = member.Name
	facMember.Level = member.Level
	facMember.LastStatus = member.LastAction.Status
	facMember.LastStatusRaw = evalLastStatus[member.LastAction.Status]
	facMember.LastSeen = member.LastAction.Relative
	facMember.LastSeenTimestamp = member.LastAction.Timestamp
	facMember.Status = member.Status.Description
	facMember.StatusRaw = 0
	facMember.Status_Long = member.Status.Details

	printer := message.NewPrinter(language.English)
	facMember.BattleStatsRaw = spyReport.Spy.Total
	switch total := spyReport.Spy.Total; {
	case total < 1000000:
		facMember.BattleStats = printer.Sprintf("%d", spyReport.Spy.Total)
	case total < 1000000000:
		facMember.BattleStats = printer.Sprintf("%.2fM", (float32(spyReport.Spy.Total) / 1000000))
	case total >= 1000000000:
		facMember.BattleStats = printer.Sprintf("%.2fB", (float32(spyReport.Spy.Total) / 1000000000))
	}

	facMember.BattleStats_StrRaw = spyReport.Spy.Strength
	switch strength := spyReport.Spy.Strength; {
	case strength < 1000000:
		facMember.BattleStats_Str = printer.Sprintf("%d", spyReport.Spy.Strength)
	case strength < 1000000000:
		facMember.BattleStats_Str = printer.Sprintf("%.2fM", (float32(spyReport.Spy.Strength) / 1000000))
	case strength >= 1000000000:
		facMember.BattleStats_Str = printer.Sprintf("%.2fB", (float32(spyReport.Spy.Strength) / 1000000000))
	}
	facMember.BattleStats_DefRaw = spyReport.Spy.Defense
	switch defense := spyReport.Spy.Defense; {
	case defense < 1000000:
		facMember.BattleStats_Def = printer.Sprintf("%d", spyReport.Spy.Defense)
	case defense < 1000000000:
		facMember.BattleStats_Def = printer.Sprintf("%.2fM", (float32(spyReport.Spy.Defense) / 1000000))
	case defense >= 1000000000:
		facMember.BattleStats_Def = printer.Sprintf("%.2fB", (float32(spyReport.Spy.Defense) / 1000000000))
	}
	facMember.BattleStats_DexRaw = spyReport.Spy.Dexterity
	switch dexterity := spyReport.Spy.Dexterity; {
	case dexterity < 1000000:
		facMember.BattleStats_Dex = printer.Sprintf("%d", spyReport.Spy.Dexterity)
	case dexterity < 1000000000:
		facMember.BattleStats_Dex = printer.Sprintf("%.2fM", (float32(spyReport.Spy.Dexterity) / 1000000))
	case dexterity >= 1000000000:
		facMember.BattleStats_Dex = printer.Sprintf("%.2fB", (float32(spyReport.Spy.Dexterity) / 1000000000))
	}
	facMember.BattleStats_SpdRaw = spyReport.Spy.Speed
	switch speed := spyReport.Spy.Speed; {
	case speed < 1000000:
		facMember.BattleStats_Spd = printer.Sprintf("%d", spyReport.Spy.Speed)
	case speed < 1000000000:
		facMember.BattleStats_Spd = printer.Sprintf("%.2fM", (float32(spyReport.Spy.Speed) / 1000000))
	case speed >= 1000000000:
		facMember.BattleStats_Spd = printer.Sprintf("%.2fB", (float32(spyReport.Spy.Speed) / 1000000000))
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	defer rdb.Close()

	err := rdb.Set(ctx, fmt.Sprintf("%d", userid), facMember, 0).Err()
	if err != nil {
		panic(err)
	}

}

func runInstance(factionId string, apiKey ApiKey, waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()
	for {
		getFactionBasicInfo(factionId, apiKey.key)
		if !apiKey.valid {
			return
		}
		time.Sleep(30300 * time.Millisecond)
	}
}

func main() {

	factionId := "46708"

	// factionId := "30085"
	// tornApiKey, ok_tornApiKey := os.LookupEnv("tornApiKey")
	// if !ok_tornApiKey {
	// 	fmt.Println("tornApiKey missing")
	// 	os.Exit(1)
	// }

	tornApiKeysString, ok_tornApiKeysString := os.LookupEnv("tornApiKeys")
	if !ok_tornApiKeysString {
		fmt.Println("tornApiKeys missing")
		os.Exit(1)
	}

	tornApiKeys := strings.Split(tornApiKeysString, ":")
	var keys []ApiKey
	for _, tornKey := range tornApiKeys {
		keys = append(keys, ApiKey{key: tornKey, valid: true})
	}
	// _ = tornApiKey

	var waitGroup sync.WaitGroup

	for _, key := range keys {

		waitGroup.Add(1)
		go runInstance(factionId, key, &waitGroup)
		time.Sleep(10100 * time.Millisecond)
	}
	waitGroup.Wait()
}
