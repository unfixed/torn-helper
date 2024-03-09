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
	"time"

	"github.com/go-redis/redis/v8"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var ctx = context.Background()

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

	URL := fmt.Sprintf("https://api.torn.com/faction/%s?selections=basic&key=%s", factionId, apiKey)
	response, err := http.Get(URL)
	if err != nil {
		log.Fatalln(err)
	}

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatalln(err)
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
	fmt.Println("updateWar running")
	// timeOffset := int64(warStart) - time.Now().Unix()

	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	startTime, err := rdb.Get(ctx, "warStartTime").Result()
	if err == redis.Nil {
		fmt.Println("warStartTime missing, setting value...")
		err := rdb.Set(ctx, "warStartTime", fmt.Sprintf("%d", warStart), 123*time.Hour).Err()
		if err != nil {
			panic(err)
		}
	}
	if startTime < fmt.Sprintf("%d", warStart) {
		fmt.Println("warStartTime is wrong, updating value...")
		err := rdb.Set(ctx, "warStartTime", fmt.Sprintf("%d", warStart), 123*time.Hour).Err()
		if err != nil {
			panic(err)
		}
	}
	opponent, err := rdb.Get(ctx, "warOpponent").Result()
	if err == redis.Nil {
		err := rdb.Set(ctx, "warOpponent", fmt.Sprintf("%d", warOpponent), 123*time.Hour).Err()
		if err != nil {
			panic(err)
		}
	}
	if opponent != fmt.Sprintf("%d", warOpponent) {
		fmt.Println("warOpponent is wrong, updating value...")
		err := rdb.Set(ctx, "warOpponent", fmt.Sprintf("%d", warOpponent), 123*time.Hour).Err()
		if err != nil {
			panic(err)
		}
	}
}

// func getUserStats(factionId string, apiKey string) {

// 	URL := fmt.Sprintf("https://api.torn.com/faction/%s?selections=basic&key=%s", factionId, apiKey)
// 	// fmt.Println(URL)
// 	response, err := http.Get(URL)
// 	if err != nil {
// 		log.Fatalln(err)
// 	}

// 	responseBody, err := io.ReadAll(response.Body)
// 	if err != nil {
// 		log.Fatalln(err)
// 	}

// 	var data FactionBasicInfo
// 	json.Unmarshal(responseBody, &data)

// 	var members []int
// 	for i, m := range data.Members {
// 		members = append(members, i)
// 		// updateMember(factionId, i, m)
// 		//updateMemberRedis(factionId, i, m)
// 	}
// 	updateFactionRedis(factionId, members)

// 	// clean up memory after execution
// 	defer response.Body.Close()
// }

func getFactionMembers(factionId string, apiKey string) {

	URL := fmt.Sprintf("https://api.torn.com/faction/%s?selections=basic&key=%s", factionId, apiKey)
	// fmt.Println(URL)
	response, err := http.Get(URL)
	if err != nil {
		log.Fatalln(err)
	}

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatalln(err)
	}

	var data FactionBasicInfo
	json.Unmarshal(responseBody, &data)

	var members []int
	for i, m := range data.Members {
		members = append(members, i)
		// updateMember(factionId, i, m)
		// personalStats :=
		updateMemberRedis(factionId, i, m, getSpyReport(i))
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
	err := rdb.Set(ctx, factionId, factionMembers, 0).Err()
	if err != nil {
		panic(err)
	}

}

func evalStatus(inputStatus Status) int {
	value := 1

	switch {
	case inputStatus.Description == "Okay":
		return value
	case strings.Contains(inputStatus.Description, "In hospital for"):
		hosptime := strings.Fields(strings.Replace(inputStatus.Description, "In hospital for", "", 1))
		_ = hosptime
		//fmt.Println(hosptime)
		// value =

	}

	// if inputStatus.Description == "Okay" {
	// 	return value
	// }
	// if strings.Contains(inputStatus.Description,"")

	return value
}

func (f FactionMember) MarshalBinary() ([]byte, error) {
	return json.Marshal(f)
}
func updateMemberRedis(factionId string, userid int, member Member, spyReport SpyUserResponse) {

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

	p := message.NewPrinter(language.English)
	facMember.BattleStatsRaw = spyReport.Spy.Total
	facMember.BattleStats = p.Sprintf("%d", spyReport.Spy.Total)

	switch total := spyReport.Spy.Total; {
	case total < 1000000:
		facMember.BattleStats = p.Sprintf("%d", spyReport.Spy.Total)
	case total <1000000000:
		facMember.BattleStats = p.Sprintf("%fM", (float32(spyReport.Spy.Total/1000000)))
	case total >=1000000000:
		facMember.BattleStats = p.Sprintf("%fB", (float32(spyReport.Spy.Total/1000000000)))


	facMember.BattleStats_StrRaw = spyReport.Spy.Strength
	facMember.BattleStats_Str = p.Sprintf("%d", spyReport.Spy.Strength)
	facMember.BattleStats_DefRaw = spyReport.Spy.Defense
	facMember.BattleStats_Def = p.Sprintf("%d", spyReport.Spy.Defense)
	facMember.BattleStats_DexRaw = spyReport.Spy.Dexterity
	facMember.BattleStats_Dex = p.Sprintf("%d", spyReport.Spy.Dexterity)
	facMember.BattleStats_SpdRaw = spyReport.Spy.Speed
	facMember.BattleStats_Spd = p.Sprintf("%d", spyReport.Spy.Speed)

	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	err := rdb.Set(ctx, fmt.Sprintf("%d", userid), facMember, 0).Err()
	if err != nil {
		panic(err)
	}

}

func main() {

	factionId := "46708"
	// factionId := "45421"
	tornApiKey, ok := os.LookupEnv("tornApiKey")
	if !ok {
		fmt.Println("tornApiKey missing")
		// fmt.Println(os.Environ())
		// fmt.Println(factionId, tornApiKey)
		os.Exit(1)
	}
	tornStatsApiKey, ok := os.LookupEnv("tornStatsApiKey")
	if !ok {
		fmt.Println("tornStatsApiKey missing start")
		// fmt.Println(os.Environ())
		// fmt.Println(factionId, tornApiKey)
		os.Exit(1)
	}
	_ = tornStatsApiKey
	getFactionBasicInfo(factionId, tornApiKey)
}
