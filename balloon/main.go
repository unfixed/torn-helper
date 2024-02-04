package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	_ "github.com/mattn/go-sqlite3"
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
	UserId      int
	Name        string
	Level       int
	LastStatus  string
	LastSeen    string
	Status      string
	Status_Long string
}
type FactionMembers struct {
	Members []int
}

// func lookupUserBasic(userid, apikey string) {
// 	fmt.Println("GETing user info...")

// 	URL := fmt.Sprintf("https://api.torn.com/user/%s?selections=basic&key=%s", userid, apikey)

// 	response, err := http.Get(URL)
// 	if err != nil {
// 		log.Fatalln(err)
// 	}

// 	responseBody, err := io.ReadAll(response.Body)
// 	if err != nil {
// 		log.Fatalln(err)
// 	}

// 	formattedData := string(responseBody)
// 	fmt.Println("Status: ", response.Status)
// 	fmt.Println("Response body: ", formattedData)

// 	// clean up memory after execution
// 	defer response.Body.Close()
// }

func getFactionBasicInfo(factionId string, apiKey string) {
	fmt.Println("GETing Faction info...")

	URL := fmt.Sprintf("https://api.torn.com/faction/%s?selections=basic&key=%s", factionId, apiKey)
	fmt.Println(URL)
	response, err := http.Get(URL)
	if err != nil {
		log.Fatalln(err)
	}

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println("Status: ", response.Status)
	// fmt.Println(string(responseBody))

	var data FactionBasicInfo
	json.Unmarshal(responseBody, &data)
	fmt.Println(data.Name)

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

	_, err := rdb.Get(ctx, "warStartTime").Result()
	if err == redis.Nil {
		err := rdb.Set(ctx, "warStartTime", fmt.Sprintf("%d", warStart), 123*time.Hour).Err()
		if err != nil {
			panic(err)
		}
	}

	_, err = rdb.Get(ctx, "warOpponent").Result()
	if err == redis.Nil {
		err := rdb.Set(ctx, "warOpponent", fmt.Sprintf("%d", warOpponent), 123*time.Hour).Err()
		if err != nil {
			panic(err)
		}
	}
}

func getFactionMembers(factionId string, apiKey string) {
	fmt.Println("GETing Faction members...")

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
	fmt.Println("Status: ", response.Status)

	var data FactionBasicInfo
	json.Unmarshal(responseBody, &data)
	fmt.Println(data.Name)

	var members []int
	for i, m := range data.Members {
		members = append(members, i)
		// updateMember(factionId, i, m)
		updateMemberRedis(factionId, i, m)
	}
	updateFactionRedis(factionId, members)

	// clean up memory after execution
	defer response.Body.Close()
}
func (f FactionMembers) MarshalBinary() ([]byte, error) {
	return json.Marshal(f)
}
func updateFactionRedis(factionId string, members []int) {
	fmt.Println("running factionredisupdate")
	fmt.Println(factionId)

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

func (f FactionMember) MarshalBinary() ([]byte, error) {
	return json.Marshal(f)
}
func updateMemberRedis(factionId string, userid int, member Member) {

	var facMember FactionMember
	facMember.Name = member.Name
	facMember.Level = member.Level
	facMember.LastStatus = member.LastAction.Status
	facMember.LastSeen = member.LastAction.Relative
	facMember.Status = member.Status.Description
	facMember.Status_Long = member.Status.Details

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

func updateMember(factionId string, userid int, member Member) {
	fmt.Println(userid)

	if _, err := os.Stat(fmt.Sprintf("./%s.sqlite", factionId)); errors.Is(err, os.ErrNotExist) {
		createMembersTable(factionId)
	}

	db, err := sql.Open("sqlite3", fmt.Sprintf("./%s.sqlite", factionId))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer db.Close()

	statement, err := db.Prepare("SELECT `userid` FROM `members` WHERE `userid`=?")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	rows, err := statement.Query(userid)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer rows.Close()

	var rowCount int = 0
	for rows.Next() {
		rowCount++
	}
	rows.Close()

	if rowCount > 0 {
		// UPDATE EXISTING MEMBER IN TABLE
		_, err := db.Exec("UPDATE `members` SET `name`=?,`level`=?,`lastaction_status`=?,`lastaction_relative`=?,`status_description`=?,`status_state`=? WHERE `userid`=?", member.Name, member.Level, member.LastAction.Status, member.LastAction.Relative, member.Status.Description, member.Status.State, userid)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

	} else {
		// INSERT MEMBER INTO TABLE
		_, err := db.Exec("INSERT INTO `members` (`userid`,`name`,`level`,`lastaction_status`,`lastaction_relative`,`status_description`,`status_state`) VALUES (?, ?, ?, ?, ?, ?, ?)", userid, member.Name, member.Level, member.LastAction.Status, member.LastAction.Relative, member.Status.Description, member.Status.State)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

}

func createMembersTable(factionId string) {
	db, err := sql.Open("sqlite3", fmt.Sprintf("./%s.sqlite", factionId))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	_, err = db.Exec("CREATE TABLE `members` (`userid` INTEGER PRIMARY KEY AUTOINCREMENT, `name` VARCHAR(64) NOT NULL, `level` INTEGER NOT NULL, `lastaction_status` VARCHAR(255) NULL, `lastaction_relative` VARCHAR(255) NULL, `status_description` VARCHAR(255) NULL, `status_state` VARCHAR(255))")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	} else {

		fmt.Println("Created members table")
	}

	db.Close()
}
func main() {

	// factionId := "46708"
	factionId := "16628"
	tornApiKey, ok := os.LookupEnv("tornApiKey")
	if !ok {
		fmt.Println("tornApiKey missing")
		// fmt.Println(os.Environ())
		// fmt.Println(factionId, tornApiKey)
		os.Exit(1)
	}

	getFactionBasicInfo(factionId, tornApiKey)
}
