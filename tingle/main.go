package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
	"github.com/go-redis/redis/v8"
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
var secs_regex *regexp.Regexp = regexp.MustCompile("([0-9]+) secs")
var mins_regex *regexp.Regexp = regexp.MustCompile("([0-9]+) mins")
var hrs_regex *regexp.Regexp = regexp.MustCompile("([0-9]+) hrs")
var remote_regex *regexp.Regexp = regexp.MustCompile("^In .*")


func checkForWar() bool {

	start := time.Now().UnixNano() / int64(time.Millisecond)
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	defer rdb.Close()

	
	_, err := rdb.Get(ctx, "warStartTime").Result()

	end := time.Now().UnixNano() / int64(time.Millisecond)

	diff := end - start
	if diff > 5 {
		fmt.Printf("checkForWar query took %d ms\n", diff)
	}
	


	if err == redis.Nil {
		return false
	}
	if err != nil {
		panic(err)
	}
	return true	

}

func getWarOpponent() (int, bool) {

	start := time.Now().UnixNano() / int64(time.Millisecond)

	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	defer rdb.Close()

	response, err := rdb.Get(ctx, "warOpponent").Result()

	end := time.Now().UnixNano() / int64(time.Millisecond)
	diff := end - start
	if diff > 5 {
		fmt.Printf("getWarrOpponent query took %d ms\n", diff)
	}

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

	start := time.Now().UnixNano() / int64(time.Millisecond)

	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	defer rdb.Close()

	result, err := rdb.Get(ctx, fmt.Sprintf("%d", factionId)).Result()

	end := time.Now().UnixNano() / int64(time.Millisecond)
	diff := end - start
	if diff > 5 {
		fmt.Printf("getOpponentMembers.factionId query took %d ms\n", diff)
	}

	if err == redis.Nil {
		return nil
	}
	if err != nil {
		panic(err)
	}

	var resultMembers FactionMembers
	json.Unmarshal([]byte(result), &resultMembers)

	for _, member := range resultMembers.Members {

		start = time.Now().UnixNano() / int64(time.Millisecond)

		result, err := rdb.Get(ctx, fmt.Sprintf("%d", member)).Result()

		end = time.Now().UnixNano() / int64(time.Millisecond)
		diff = end - start
		if diff > 5 {
			fmt.Printf("getOpponentMembers.member query took %d ms\n", diff)
		}

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

func evalStatus(inputStatus string) int {

	start_evalStatus := time.Now().UnixNano() / int64(time.Millisecond)

	if inputStatus == "Okay" {
		return -1
	}
	if inputStatus == "Fallen" {
		return 99999999
	}

	var calculated_value int = 0

	// start_regexcompile := time.Now().UnixNano() / int64(time.Millisecond)

	// // secs_regex := regexp.MustCompile("([0-9]+) secs")
	// // mins_regex := regexp.MustCompile("([0-9]+) mins")
	// // hrs_regex := regexp.MustCompile("([0-9]+) hrs")
	// // remote_regex := regexp.MustCompile("^In .*")

	// end_regexcompile := time.Now().UnixNano() / int64(time.Millisecond)
	// diff_regexcompile := end_regexcompile - start_regexcompile

	start_locationeval := time.Now().UnixNano() / int64(time.Millisecond)

	switch status := inputStatus; {
	case strings.Contains(status, "Mexico") || strings.Contains(status, "Mexican"):
		calculated_value += 1000000
	case strings.Contains(status, "Cayman Islands") || strings.Contains(status, "Caymanian"):
		calculated_value += 2000000
	case strings.Contains(status, "Canada") || strings.Contains(status, "Canadian"):
		calculated_value += 3000000
	case strings.Contains(status, "Hawaii") || strings.Contains(status, "Hawaiian"):
		calculated_value += 4000000
	case strings.Contains(status, "United Kingdom") || strings.Contains(status, "British"):
		calculated_value += 5000000
	case strings.Contains(status, "Argentina") || strings.Contains(status, "Argentinian"):
		calculated_value += 6000000
	case strings.Contains(status, "Switzerland") || strings.Contains(status, "Swiss"):
		calculated_value += 7000000
	case strings.Contains(status, "Japan") || strings.Contains(status, "Japanese"):
		calculated_value += 8000000
	case strings.Contains(status, "China") || strings.Contains(status, "Chinese"):
		calculated_value += 9000000
	case strings.Contains(status, "UAE") || strings.Contains(status, "Emirati"):
		calculated_value += 10000000
	case strings.Contains(status, "South Africa") || strings.Contains(status, "South African"):
		calculated_value += 11000000
	}

	end_locationeval := time.Now().UnixNano() / int64(time.Millisecond)
	diff_locationeval := end_locationeval - start_locationeval

	start_flighteval := time.Now().UnixNano() / int64(time.Millisecond)

	switch status := inputStatus; {
	case strings.Contains(status, "Returning to Torn from "):
		calculated_value += 1
	case remote_regex.MatchString(status):
		calculated_value += 2
	case strings.Contains(status, "Traveling to "):
		calculated_value += 3
	}
	end_flighteval := time.Now().UnixNano() / int64(time.Millisecond)
	diff_flighteval := end_flighteval - start_flighteval


	start_hospeval := time.Now().UnixNano() / int64(time.Millisecond)

	if strings.Contains(inputStatus, "hospital") || strings.Contains(inputStatus, "jail") {
		var hosp_eval int = 3

		if strings.Contains(inputStatus, "hrs") {
			//hours
			hrs_ticks, _ := strconv.Atoi(hrs_regex.FindStringSubmatch(inputStatus)[1])
			hrs_ticks = hrs_ticks * 3600
			// fmt.Println( hrs_ticks )
			hosp_eval += hrs_ticks
		}
		if strings.Contains(inputStatus, "mins") {
			//minutes
			mins_ticks, _ := strconv.Atoi(mins_regex.FindStringSubmatch(inputStatus)[1])
			mins_ticks = mins_ticks * 60
			// fmt.Println( mins_ticks )
			hosp_eval += mins_ticks
		}
		if strings.Contains(inputStatus, "secs") {
			//seconds
			secs_ticks, _ := strconv.Atoi(secs_regex.FindStringSubmatch(inputStatus)[1])
			// fmt.Println( secs_ticks )
			hosp_eval += secs_ticks
		}
		calculated_value += hosp_eval
	}
	end_hospeval := time.Now().UnixNano() / int64(time.Millisecond)
	diff_hospeval := end_hospeval - start_hospeval

		



	end_evalStatus := time.Now().UnixNano() / int64(time.Millisecond)
	diff_evalStatus := end_evalStatus - start_evalStatus
	if diff_evalStatus > 5 {
		fmt.Printf("evalStatus took %d ms\n", diff_evalStatus)
		// fmt.Printf("evalStatus.regexcompile took %d ms\n", diff_regexcompile)
		fmt.Printf("evalStatus.locationeval took %d ms\n", diff_locationeval)
		fmt.Printf("evalStatus.flighteval took %d ms\n", diff_flighteval)
		fmt.Printf("evalStatus.hosp_eval took %d ms\n", diff_hospeval)
		
		

	}

	return calculated_value
}

func sortMembers(inputMembers map[int]FactionMember, sortBy string, sortDirection string) []FactionMember {

	start := time.Now().UnixNano() / int64(time.Millisecond)

	var factionMembers []FactionMember

	evalLastStatus := map[string]int{"Offline": 0, "Idle": 1, "Online": 2}

	temp := inputMembers
	var highestStats FactionMember
	var highestIndex int = -1
	size := len(temp)
	for i := 1; i < size; i++ {
		for k, m := range temp {
			switch sortBy {
			case "Status":
				if sortDirection == "asc" {
					if (highestIndex == 0) || (evalStatus(highestStats.Status) < evalStatus(m.Status)) {
						highestStats = m
						highestIndex = k
					}
				} else {
					if (highestIndex == 0) || (evalStatus(highestStats.Status) > evalStatus(m.Status)) {
						highestStats = m
						highestIndex = k
					}
				}
				if (evalStatus(highestStats.Status) == evalStatus(m.Status)) && (evalLastStatus[highestStats.LastStatus] < evalLastStatus[m.LastStatus]) {
					highestStats = m
					highestIndex = k
				}
				if (evalStatus(highestStats.Status) == evalStatus(m.Status)) && (highestStats.LastStatus == m.LastStatus) && (highestStats.BattleStatsRaw < m.BattleStatsRaw) {
					highestStats = m
					highestIndex = k
				}
				if (evalStatus(highestStats.Status) == evalStatus(m.Status)) && (highestStats.LastStatus == m.LastStatus) && (highestStats.BattleStatsRaw == m.BattleStatsRaw) && (highestStats.Level < m.Level) {
					highestStats = m
					highestIndex = k
				}
				if (evalStatus(highestStats.Status) == evalStatus(m.Status)) && (highestStats.LastStatus == m.LastStatus) && (highestStats.BattleStatsRaw == m.BattleStatsRaw) && (highestStats.Level == m.Level) && (highestStats.Name < m.Name) {
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
				if (highestStats.LastStatus == m.LastStatus) && (evalStatus(highestStats.Status) > evalStatus(m.Status)) {
					highestStats = m
					highestIndex = k
				}
				if (highestStats.LastStatus == m.LastStatus) && (evalStatus(highestStats.Status) == evalStatus(m.Status)) && (highestStats.BattleStatsRaw < m.BattleStatsRaw) {
					highestStats = m
					highestIndex = k
				}
				if (highestStats.LastStatus == m.LastStatus) && (evalStatus(highestStats.Status) == evalStatus(m.Status)) && (highestStats.BattleStatsRaw == m.BattleStatsRaw) && (highestStats.Level < m.Level) {
					highestStats = m
					highestIndex = k
				}
				if (highestStats.LastStatus == m.LastStatus) && (evalStatus(highestStats.Status) == evalStatus(m.Status)) && (highestStats.BattleStatsRaw == m.BattleStatsRaw) && (highestStats.Level == m.Level) && (highestStats.Name < m.Name) {
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
				if (highestStats.BattleStatsRaw == m.BattleStatsRaw) && (evalStatus(highestStats.Status) > evalStatus(m.Status)) {
					highestStats = m
					highestIndex = k
				}
				if (highestStats.BattleStatsRaw == m.BattleStatsRaw) && (evalStatus(highestStats.Status) == evalStatus(m.Status)) && (evalLastStatus[highestStats.LastStatus] < evalLastStatus[m.LastStatus]) {
					highestStats = m
					highestIndex = k
				}
				if (highestStats.BattleStatsRaw == m.BattleStatsRaw) && (evalStatus(highestStats.Status) == evalStatus(m.Status)) && (highestStats.LastStatus == m.LastStatus) && (highestStats.Level < m.Level) {
					highestStats = m
					highestIndex = k
				}
				if (highestStats.BattleStatsRaw == m.BattleStatsRaw) && (evalStatus(highestStats.Status) == evalStatus(m.Status)) && (highestStats.LastStatus == m.LastStatus) && (highestStats.Level == m.Level) && (highestStats.Name < m.Name) {
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
				if (highestStats.Level == m.Level) && (evalStatus(highestStats.Status) > evalStatus(m.Status)) {
					highestStats = m
					highestIndex = k
				}
				if (highestStats.Level == m.Level) && (evalStatus(highestStats.Status) == evalStatus(m.Status)) && (evalLastStatus[highestStats.LastStatus] < evalLastStatus[m.LastStatus]) {
					highestStats = m
					highestIndex = k
				}
				if (highestStats.Level == m.Level) && (evalStatus(highestStats.Status) == evalStatus(m.Status)) && (highestStats.LastStatus == m.LastStatus) && (highestStats.BattleStatsRaw < m.BattleStatsRaw) {
					highestStats = m
					highestIndex = k
				}
				if (highestStats.Level == m.Level) && (evalStatus(highestStats.Status) == evalStatus(m.Status)) && (highestStats.LastStatus == m.LastStatus) && (highestStats.BattleStatsRaw == m.BattleStatsRaw) && (highestStats.Name < m.Name) {
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
				if (highestStats.Name == m.Name) && (evalStatus(highestStats.Status) > evalStatus(m.Status)) {
					highestStats = m
					highestIndex = k
				}
				if (highestStats.Name == m.Name) && (evalStatus(highestStats.Status) == evalStatus(m.Status)) && (evalLastStatus[highestStats.LastStatus] < evalLastStatus[m.LastStatus]) {
					highestStats = m
					highestIndex = k
				}
				if (highestStats.Name == m.Name) && (evalStatus(highestStats.Status) == evalStatus(m.Status)) && (highestStats.LastStatus == m.LastStatus) && (highestStats.BattleStatsRaw < m.BattleStatsRaw) {
					highestStats = m
					highestIndex = k
				}
				if (highestStats.Name == m.Name) && (evalStatus(highestStats.Status) == evalStatus(m.Status)) && (highestStats.LastStatus == m.LastStatus) && (highestStats.BattleStatsRaw == m.BattleStatsRaw) && (highestStats.Level < m.Level) {
					highestStats = m
					highestIndex = k
				}
			default:
				if (highestIndex == 0) || (evalStatus(highestStats.Status) > evalStatus(m.Status)) {
					highestStats = m
					highestIndex = k
				}
				if (evalStatus(highestStats.Status) == evalStatus(m.Status)) && (evalLastStatus[highestStats.LastStatus] < evalLastStatus[m.LastStatus]) {
					highestStats = m
					highestIndex = k
				}
				if (evalStatus(highestStats.Status) == evalStatus(m.Status)) && (highestStats.LastStatus == m.LastStatus) && (highestStats.BattleStatsRaw < m.BattleStatsRaw) {
					highestStats = m
					highestIndex = k

				}
				if (evalStatus(highestStats.Status) == evalStatus(m.Status)) && (highestStats.LastStatus == m.LastStatus) && (highestStats.BattleStatsRaw == m.BattleStatsRaw) && (highestStats.Level < m.Level) {
					highestStats = m
					highestIndex = k
				}
				if (evalStatus(highestStats.Status) == evalStatus(m.Status)) && (highestStats.LastStatus == m.LastStatus) && (highestStats.BattleStatsRaw == m.BattleStatsRaw) && (highestStats.Level == m.Level) && (highestStats.Name < m.Name) {
					highestStats = m
					highestIndex = k
				}
			}
		}
		factionMembers = append(factionMembers, highestStats)
		delete(temp, highestIndex)
		highestIndex = 0

	}

	end := time.Now().UnixNano() / int64(time.Millisecond)
	diff := end - start
	if diff > 5 {
		fmt.Printf("sortMembers took %d ms\n", diff)
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
