package main

import (
	"context"
	"database/sql"
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

func getFactionMembers(factionId int) (map[int]FactionMember, bool) {
	// fmt.Println("running getFactionMembers")
	factionMembers := make(map[int]FactionMember)

	db, err := sql.Open("sqlite3", fmt.Sprintf("../balloon/%d.sqlite", factionId))
	if err != nil {
		fmt.Println(err)
		return nil, false
	}
	defer db.Close()

	statement, err := db.Prepare("SELECT `userid`,`name`,`level`,`lastaction_status`,`lastaction_relative`,`status_description`,`status_state` FROM `members`")
	if err != nil {
		fmt.Println(err)
		return nil, false
	}

	rows, err := statement.Query()
	if err != nil {
		fmt.Println(err)
		return nil, false
	}
	defer rows.Close()

	for rows.Next() {
		var member FactionMember
		rows.Scan(&member.UserId, &member.Name, &member.Level, &member.LastStatus, &member.LastSeen, &member.Status, &member.Status_Long)
		factionMembers[member.UserId] = member
	}
	rows.Close()
	return factionMembers, true
}

func getOpponentMembers(factionId int) (map[int]FactionMember, bool) {
	factionMembers := make(map[int]FactionMember)

	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	// var members FactionMembers
	result, err := rdb.Get(ctx, fmt.Sprintf("%d", factionId)).Result()
	if err == redis.Nil {
		return nil, false
	}
	if err != nil {
		panic(err)
	}

	var resultMembers FactionMembers
	json.Unmarshal([]byte(result), &resultMembers)

	for _, member := range resultMembers.Members {
		// fmt.Println(fmt.Sprintf("%d", member))

		result, err := rdb.Get(ctx, fmt.Sprintf("%d", member)).Result()
		if err == redis.Nil {
			return nil, false
		}
		if err != nil {
			panic(err)
		}
		var resultMember FactionMember
		json.Unmarshal([]byte(result), &resultMember)
		// factionMembers = append(factionMembers, resultMember)
		factionMembers[member] = resultMember
	}

	// fmt.Println(factionMembers)
	return factionMembers, true
}
func viewIndex(w http.ResponseWriter, r *http.Request) {

	tmpl := template.Must(template.ParseFiles("templates/index.html"))

	tmpl.Execute(w, nil)
}

func viewMemberList(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/memberlist.html"))
	// var ctx = nil
	if !checkForWar() {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	warOpponent, ok := getWarOpponent()
	if !ok {
		// fmt.Println("failed to get war opponent")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// fmt.Printf(fmt.Sprintf("%d", warOpponent))
	// getOpponentMembers(warOpponent)
	ctx, ok := getOpponentMembers(warOpponent)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, ctx)
}

func main() {

	fmt.Println("yo")
	// fmt.Println(fmt.Sprintf("%d", time.Now().Unix()))
	http.HandleFunc("/", viewIndex)
	http.HandleFunc("/memberlist", viewMemberList)
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	log.Fatal(http.ListenAndServe(":8000", nil))

}
