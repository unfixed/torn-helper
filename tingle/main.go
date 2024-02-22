package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/go-redis/redis/v8"
	_ "github.com/mattn/go-sqlite3"
)

type FactionMember struct {
	UserId             int
	Name               string
	Level              int
	LastStatus         string
	LastSeen           string
	Status             string
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

var ctx = context.Background()

func checkForWar() bool {

	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	defer rdb.Close()

	_, err := rdb.Get(ctx, "warStartTime").Result()
	if err == redis.Nil {
		return false
	}
	if err != nil {
		panic(err)
	}
	return true

}

func getWarOpponent() (int, bool) {

	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	defer rdb.Close()

	response, err := rdb.Get(ctx, "warOpponent").Result()
	if err == redis.Nil {
		return 0, false
	}
	if err != nil {
		panic(err)
	}
	result, err := strconv.Atoi(response)
	if err != nil {
		panic(err)
	}
	return result, true
}

func getOpponentMembers(factionId int) map[int]FactionMember {
	factionMembers := make(map[int]FactionMember)

	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	defer rdb.Close()

	result, err := rdb.Get(ctx, fmt.Sprintf("%d", factionId)).Result()
	if err == redis.Nil {
		return nil
	}
	if err != nil {
		panic(err)
	}

	var resultMembers FactionMembers
	json.Unmarshal([]byte(result), &resultMembers)

	for _, member := range resultMembers.Members {

		result, err := rdb.Get(ctx, fmt.Sprintf("%d", member)).Result()
		if err == redis.Nil {
			return nil
		}
		if err != nil {
			panic(err)
		}
		var resultMember FactionMember
		json.Unmarshal([]byte(result), &resultMember)
		factionMembers[member] = resultMember
	}

	return factionMembers
}

func sortMembers(inputMembers map[int]FactionMember, sortBy string, sortDirection string) []FactionMember {
	var factionMembers []FactionMember

	evalLastStatus := map[string]int{"Offline": 0, "Idle": 1, "Online": 2}
	evalStatus := map[string]int{
		"Okay":                                  -1,
		"Returning to Torn from Mexico":         10000,
		"In Mexico":                             10001,
		"Traveling to Mexico":                   10002,
		"Returning to Torn from Cayman Islands": 20000,
		"In Cayman Islands":                     20001,
		"Traveling to Cayman Islands":           20002,
		"Returning to Torn from Canada":         30000,
		"In Canada":                             30001,
		"Traveling to Canada":                   30002,
		"Returning to Torn from Hawaii":         40000,
		"In Hawaii":                             40001,
		"Traveling to Hawaii":                   40002,
		"Returning to Torn from United Kingdom": 50000,
		"In United Kingdom":                     50001,
		"Traveling to United Kingdom":           50002,
		"Returning to Torn from Argentina":      60000,
		"In Argentina":                          60001,
		"Traveling to Argentina":                60002,
		"Returning to Torn from Switzerland":    70000,
		"In Switzerland":                        70001,
		"Traveling to Switzerland":              70002,
		"Returning to Torn from Japan":          80000,
		"In Japan":                              80001,
		"Traveling to Japan":                    80002,
		"Returning to Torn from China":          90000,
		"In China":                              90001,
		"Traveling to China":                    90002,
		"Returning to Torn from UAE":            100000,
		"In UAE":                                100001,
		"Traveling to UAE":                      100002,
		"Returning to Torn from South Africa":   110000,
		"In South Africa":                       110001,
		"Traveling to South Africa":             110002,
		"Fallen":                                1000000,
	}

	temp := inputMembers
	var highestStats FactionMember
	var highestIndex int = -1
	size := len(temp)
	for i := 1; i < size; i++ {
		for k, m := range temp {
			switch sortBy {
			case "Status":
				if sortDirection == "asc" {
					if (highestIndex == 0) || (evalStatus[highestStats.Status] < evalStatus[m.Status]) {
						highestStats = m
						highestIndex = k
					}
				} else {
					if (highestIndex == 0) || (evalStatus[highestStats.Status] > evalStatus[m.Status]) {
						highestStats = m
						highestIndex = k
					}
				}
				if (evalStatus[highestStats.Status] == evalStatus[m.Status]) && (evalLastStatus[highestStats.LastStatus] < evalLastStatus[m.LastStatus]) {
					highestStats = m
					highestIndex = k
				}
				if (evalStatus[highestStats.Status] == evalStatus[m.Status]) && (highestStats.LastStatus == m.LastStatus) && (highestStats.BattleStatsRaw < m.BattleStatsRaw) {
					highestStats = m
					highestIndex = k
				}
				if (evalStatus[highestStats.Status] == evalStatus[m.Status]) && (highestStats.LastStatus == m.LastStatus) && (highestStats.BattleStatsRaw == m.BattleStatsRaw) && (highestStats.Level < m.Level) {
					highestStats = m
					highestIndex = k
				}
				if (evalStatus[highestStats.Status] == evalStatus[m.Status]) && (highestStats.LastStatus == m.LastStatus) && (highestStats.BattleStatsRaw == m.BattleStatsRaw) && (highestStats.Level == m.Level) && (highestStats.Name < m.Name) {
					highestStats = m
					highestIndex = k
				}
			case "LastStatus":
				if sortDirection == "asc" {
					if (highestIndex == 0) || (evalLastStatus[highestStats.LastStatus] > evalLastStatus[m.LastStatus]) {
						highestStats = m
						highestIndex = k
					}
				} else {
					if (highestIndex == 0) || (evalLastStatus[highestStats.LastStatus] < evalLastStatus[m.LastStatus]) {
						highestStats = m
						highestIndex = k
					}
				}
				if (highestStats.LastStatus == m.LastStatus) && (evalStatus[highestStats.Status] > evalStatus[m.Status]) {
					highestStats = m
					highestIndex = k
				}
				if (highestStats.LastStatus == m.LastStatus) && (evalStatus[highestStats.Status] == evalStatus[m.Status]) && (highestStats.BattleStatsRaw < m.BattleStatsRaw) {
					highestStats = m
					highestIndex = k
				}
				if (highestStats.LastStatus == m.LastStatus) && (evalStatus[highestStats.Status] == evalStatus[m.Status]) && (highestStats.BattleStatsRaw == m.BattleStatsRaw) && (highestStats.Level < m.Level) {
					highestStats = m
					highestIndex = k
				}
				if (highestStats.LastStatus == m.LastStatus) && (evalStatus[highestStats.Status] == evalStatus[m.Status]) && (highestStats.BattleStatsRaw == m.BattleStatsRaw) && (highestStats.Level == m.Level) && (highestStats.Name < m.Name) {
					highestStats = m
					highestIndex = k
				}
			case "BattleStats":
				if sortDirection == "asc" {
					if (highestIndex == 0) || (highestStats.BattleStatsRaw > m.BattleStatsRaw) {
						highestStats = m
						highestIndex = k
					}
				} else {
					if (highestIndex == 0) || (highestStats.BattleStatsRaw < m.BattleStatsRaw) {
						highestStats = m
						highestIndex = k
					}
				}
				if (highestStats.BattleStatsRaw == m.BattleStatsRaw) && (evalStatus[highestStats.Status] > evalStatus[m.Status]) {
					highestStats = m
					highestIndex = k
				}
				if (highestStats.BattleStatsRaw == m.BattleStatsRaw) && (evalStatus[highestStats.Status] == evalStatus[m.Status]) && (evalLastStatus[highestStats.LastStatus] < evalLastStatus[m.LastStatus]) {
					highestStats = m
					highestIndex = k
				}
				if (highestStats.BattleStatsRaw == m.BattleStatsRaw) && (evalStatus[highestStats.Status] == evalStatus[m.Status]) && (highestStats.LastStatus == m.LastStatus) && (highestStats.Level < m.Level) {
					highestStats = m
					highestIndex = k
				}
				if (highestStats.BattleStatsRaw == m.BattleStatsRaw) && (evalStatus[highestStats.Status] == evalStatus[m.Status]) && (highestStats.LastStatus == m.LastStatus) && (highestStats.Level == m.Level) && (highestStats.Name < m.Name) {
					highestStats = m
					highestIndex = k
				}
			case "Level":
				if sortDirection == "asc" {
					if (highestIndex == 0) || (highestStats.Level > m.Level) {
						highestStats = m
						highestIndex = k
					}
				} else {
					if (highestIndex == 0) || (highestStats.Level < m.Level) {
						highestStats = m
						highestIndex = k
					}
				}
				if (highestStats.Level == m.Level) && (evalStatus[highestStats.Status] > evalStatus[m.Status]) {
					highestStats = m
					highestIndex = k
				}
				if (highestStats.Level == m.Level) && (evalStatus[highestStats.Status] == evalStatus[m.Status]) && (evalLastStatus[highestStats.LastStatus] < evalLastStatus[m.LastStatus]) {
					highestStats = m
					highestIndex = k
				}
				if (highestStats.Level == m.Level) && (evalStatus[highestStats.Status] == evalStatus[m.Status]) && (highestStats.LastStatus == m.LastStatus) && (highestStats.BattleStatsRaw < m.BattleStatsRaw) {
					highestStats = m
					highestIndex = k
				}
				if (highestStats.Level == m.Level) && (evalStatus[highestStats.Status] == evalStatus[m.Status]) && (highestStats.LastStatus == m.LastStatus) && (highestStats.BattleStatsRaw == m.BattleStatsRaw) && (highestStats.Name < m.Name) {
					highestStats = m
					highestIndex = k
				}
			case "Name":
				if sortDirection == "asc" {
					if (highestIndex == 0) || (highestStats.Name > m.Name) {
						highestStats = m
						highestIndex = k
					}
				} else {
					if (highestIndex == 0) || (highestStats.Name < m.Name) {
						highestStats = m
						highestIndex = k
					}
				}
				if (highestStats.Name == m.Name) && (evalStatus[highestStats.Status] > evalStatus[m.Status]) {
					highestStats = m
					highestIndex = k
				}
				if (highestStats.Name == m.Name) && (evalStatus[highestStats.Status] == evalStatus[m.Status]) && (evalLastStatus[highestStats.LastStatus] < evalLastStatus[m.LastStatus]) {
					highestStats = m
					highestIndex = k
				}
				if (highestStats.Name == m.Name) && (evalStatus[highestStats.Status] == evalStatus[m.Status]) && (highestStats.LastStatus == m.LastStatus) && (highestStats.BattleStatsRaw < m.BattleStatsRaw) {
					highestStats = m
					highestIndex = k
				}
				if (highestStats.Name == m.Name) && (evalStatus[highestStats.Status] == evalStatus[m.Status]) && (highestStats.LastStatus == m.LastStatus) && (highestStats.BattleStatsRaw == m.BattleStatsRaw) && (highestStats.Level < m.Level) {
					highestStats = m
					highestIndex = k
				}
			default:
				if (highestIndex == 0) || (evalStatus[highestStats.Status] > evalStatus[m.Status]) {
					highestStats = m
					highestIndex = k
				}
				if (evalStatus[highestStats.Status] == evalStatus[m.Status]) && (evalLastStatus[highestStats.LastStatus] < evalLastStatus[m.LastStatus]) {
					highestStats = m
					highestIndex = k
				}
				if (evalStatus[highestStats.Status] == evalStatus[m.Status]) && (highestStats.LastStatus == m.LastStatus) && (highestStats.BattleStatsRaw < m.BattleStatsRaw) {
					highestStats = m
					highestIndex = k

				}
				if (evalStatus[highestStats.Status] == evalStatus[m.Status]) && (highestStats.LastStatus == m.LastStatus) && (highestStats.BattleStatsRaw == m.BattleStatsRaw) && (highestStats.Level < m.Level) {
					highestStats = m
					highestIndex = k
				}
				if (evalStatus[highestStats.Status] == evalStatus[m.Status]) && (highestStats.LastStatus == m.LastStatus) && (highestStats.BattleStatsRaw == m.BattleStatsRaw) && (highestStats.Level == m.Level) && (highestStats.Name < m.Name) {
					highestStats = m
					highestIndex = k
				}
			}
		}
		factionMembers = append(factionMembers, highestStats)
		delete(temp, highestIndex)
		highestIndex = 0

	}
	return factionMembers
}

// func sortOpponentMembers(factionMembers map[int]FactionMember) map[int]FactionMember {
// 	var order []int
// 	for _, member := range factionMembers {

// 	order +=
// 	return factionMembers
// }

func viewIndex(w http.ResponseWriter, r *http.Request) {

	tmpl := template.Must(template.ParseFiles("templates/index.html"))

	sortBy := r.URL.Query().Get("sortby")
	sortDirection := r.URL.Query().Get("sortdirection")
	ctx := map[string]any{"sortBy": sortBy, "sortDirection": sortDirection}

	tmpl.Execute(w, ctx)
}

func viewMemberList(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/memberlist.html"))
	if !checkForWar() {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	warOpponent, ok := getWarOpponent()
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	sortBy := r.URL.Query().Get("sortby")
	sortDirection := r.URL.Query().Get("sortdirection")
	ctx := map[string]any{"members": sortMembers(getOpponentMembers(warOpponent), sortBy, sortDirection), "sortBy": sortBy, "sortDirection": sortDirection}

	tmpl.Execute(w, ctx)
}

func main() {

	http.HandleFunc("/", viewIndex)
	http.HandleFunc("/memberlist", viewMemberList)
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	tingleSSL, ok := os.LookupEnv("tingleSSL")
	if !ok {
		fmt.Println("tingleSSL missing, running as http on port 8000")
		log.Fatal(http.ListenAndServe(":8000", nil))
	}
	tingleSSLCert := tingleSSL + "fullchain.pem"
	tingleSSLkey := tingleSSL + "privkey.pem"
	log.Fatal(http.ListenAndServeTLS(":443", tingleSSLCert, tingleSSLkey, nil))

}
