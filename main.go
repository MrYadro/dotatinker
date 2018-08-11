package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

type liveResponses []struct {
	liveResponse
}

type liveResponse struct {
	ActivateTime   int    `json:"activate_time"`
	DeactivateTime int    `json:"deactivate_time"`
	ServerSteamID  string `json:"server_steam_id"`
	LobbyID        string `json:"lobby_id"`
	LeagueID       int    `json:"league_id"`
	LobbyType      int    `json:"lobby_type"`
	GameTime       int    `json:"game_time"`
	Delay          int    `json:"delay"`
	Spectators     int    `json:"spectators"`
	GameMode       int    `json:"game_mode"`
	AverageMmr     int    `json:"average_mmr"`
	SortScore      int    `json:"sort_score"`
	LastUpdateTime int    `json:"last_update_time"`
	RadiantLead    int    `json:"radiant_lead"`
	RadiantScore   int    `json:"radiant_score"`
	DireScore      int    `json:"dire_score"`
	Players        []struct {
		AccountID   int    `json:"account_id"`
		HeroID      int    `json:"hero_id"`
		Name        string `json:"name,omitempty"`
		CountryCode string `json:"country_code,omitempty"`
		FantasyRole int    `json:"fantasy_role,omitempty"`
		TeamID      int    `json:"team_id,omitempty"`
		TeamName    string `json:"team_name,omitempty"`
		TeamTag     string `json:"team_tag,omitempty"`
		IsLocked    bool   `json:"is_locked,omitempty"`
		IsPro       bool   `json:"is_pro,omitempty"`
		LockedUntil int    `json:"locked_until,omitempty"`
	} `json:"players"`
	BuildingState              int    `json:"building_state"`
	TeamNameRadiant            string `json:"team_name_radiant,omitempty"`
	TeamNameDire               string `json:"team_name_dire,omitempty"`
	TeamLogoRadiant            string `json:"team_logo_radiant,omitempty"`
	TeamLogoDire               string `json:"team_logo_dire,omitempty"`
	WeekendTourneyTournamentID int    `json:"weekend_tourney_tournament_id,omitempty"`
	WeekendTourneyDivision     int    `json:"weekend_tourney_division,omitempty"`
	WeekendTourneySkillLevel   int    `json:"weekend_tourney_skill_level,omitempty"`
	WeekendTourneyBracketRound int    `json:"weekend_tourney_bracket_round,omitempty"`
}

type widgetTeam struct {
	Name string `json:"name"`
}

type widgetScore struct {
	TeamA int `json:"team_a"`
	TeamB int `json:"team_b"`
}

type widgetMatch struct {
	TeamA widgetTeam  `json:"team_a"`
	TeamB widgetTeam  `json:"team_b"`
	Score widgetScore `json:"score"`
}

type widgetJSON struct {
	Title   string        `json:"title"`
	Matches []widgetMatch `json:"matches"`
}

type appConfig struct {
	VkAPIkey  string `json:"vkAPIkey"`
	Whitelist []int  `json:"whitelist"`
}

func loadAppConfig() appConfig {
	buf := bytes.NewBuffer(nil)
	f, _ := os.Open("config/app.json")
	io.Copy(buf, f)
	f.Close()

	var jsonobject appConfig

	err := json.Unmarshal(buf.Bytes(), &jsonobject)

	if err != nil {
		fmt.Println("error:", err)
	}

	return jsonobject
}

var myClient = &http.Client{Timeout: 10 * time.Second}

func getLiveMatches(url string, target interface{}) error {
	r, err := myClient.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(target)
}

func sendPayloadToVK(payload string, widgetType string, accessToken string) (err error) {
	_, err = http.Get("https://api.vk.com/method/" + "appWidgets.update?" +
		url.Values{
			"access_token": {accessToken},
			"v":            {"5.80"},
			"type":         {widgetType},
			"code":         {payload}}.Encode())
	if err != nil {
		log.Fatal(err)
	}
	return nil
}

func main() {
	config := loadAppConfig()

	liveMatches := liveResponses{}
	getLiveMatches("https://api.opendota.com/api/live", &liveMatches)

	var matchesCount int

	vkPayloadMatches := []widgetMatch{}

	for _, match := range liveMatches {
		for _, league := range config.Whitelist {
			if match.LeagueID == league && matchesCount < 5 {
				matchJSON := widgetMatch{
					TeamA: widgetTeam{Name: match.TeamNameRadiant},
					TeamB: widgetTeam{Name: match.TeamNameDire},
					Score: widgetScore{TeamA: match.RadiantScore, TeamB: match.DireScore},
				}
				vkPayloadMatches = append(vkPayloadMatches, matchJSON)
				matchesCount++
			}
		}
	}

	if matchesCount != 0 {
		vkPayload := widgetJSON{
			Title:   "Live Dota 2 Matches",
			Matches: vkPayloadMatches,
		}

		b, err := json.Marshal(vkPayload)
		if err != nil {
			fmt.Println("error:", err)
		}
		s := string(b[:])

		payload := "return" + s + ";"
		sendPayloadToVK(payload, "matches", config.VkAPIkey)
	} else {
		payload := "return{\"title\":\"Live Dota 2 Matches\",\"text\": \"No live matches in progress\"};"
		sendPayloadToVK(payload, "text", config.VkAPIkey)
	}
}
