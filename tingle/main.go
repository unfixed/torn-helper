package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/go-redis/redis/v8"
	_ "github.com/mattn/go-sqlite3"
)

type FactionMember struct {
	UserId         int
	Name           string
	Level          int
	LastStatus     string
	LastSeen       string
	Status         string
	Status_Long    string
	BattleStats    string
	BattleStatsRaw int64
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

func sortMembers(inputMembers map[int]FactionMember) []FactionMember {
	var factionMembers []FactionMember

	temp := inputMembers
	var highestStats FactionMember
	var highestIndex int = -1
	size := len(temp)
	for i := 1; i < size; i++ {
		for k, m := range temp {
			if (highestIndex == 0) || (highestStats.Status < m.Status) {
				highestStats = m
				highestIndex = k

			}

			if (highestStats.Status == m.Status) && (highestStats.BattleStatsRaw < m.BattleStatsRaw) {
				highestStats = m
				highestIndex = k

			}
			if (highestStats.Status < m.Status) && (highestStats.BattleStatsRaw == m.BattleStatsRaw) && (highestStats.Level < m.Level) {
				highestStats = m
				highestIndex = k
			}
			if (highestStats.Status < m.Status) && (highestStats.BattleStatsRaw == m.BattleStatsRaw) && (highestStats.Level == m.Level) && (highestStats.Name < m.Name) {
				highestStats = m
				highestIndex = k
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

	tmpl.Execute(w, nil)
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

	ctx := sortMembers(getOpponentMembers(warOpponent))

	// ctx, ok := getOpponentMembers(warOpponent)

	// if !ok {
	// 	w.WriteHeader(http.StatusInternalServerError)
	// 	return
	// }
	tmpl.Execute(w, ctx)
}

func main() {

	http.HandleFunc("/", viewIndex)
	http.HandleFunc("/memberlist", viewMemberList)
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	log.Fatal(http.ListenAndServe(":8000", nil))

}
