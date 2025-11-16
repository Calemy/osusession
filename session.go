package main

import (
	"encoding/json"
	"sync"
	"time"
)

type Session struct {
	Username string        `json:"username"`
	Start    Stats         `json:"start"`
	Last     Stats         `json:"last"`
	Session  ActiveSession `json:"session"`
}

type Result struct {
	ID         int        `json:"id"`
	Username   string     `json:"username"`
	Online     bool       `json:"is_online"`
	Statistics Statistics `json:"statistics"`
}

type Stats struct {
	Time       time.Time  `json:"time"`
	Statistics Statistics `json:"statistics"`
}

type ActiveSession struct {
	Time        int     `json:"time"`
	GlobalRank  int     `json:"global_rank"`
	CountryRank int     `json:"country_rank"`
	PP          float64 `json:"pp"`
	RawPP       float64 `json:"raw_pp"`
	RankedScore int     `json:"ranked_score"`
	TotalScore  int     `json:"total_score"`
	Accuracy    float64 `json:"hit_accuracy"`
	Playcount   int     `json:"play_count"`
	Playtime    int     `json:"play_time"`
	TotalHits   int     `json:"total_hits"`
	Level       int     `json:"level"`
	Scores      int     `json:"scores"`
}

/*
choke detection (4/5 max_combo)
track scores
*/

type Statistics struct {
	GlobalRank  int     `json:"global_rank"`
	CountryRank int     `json:"country_rank"`
	PP          float64 `json:"pp"`
	RankedScore int     `json:"ranked_score"`
	TotalScore  int     `json:"total_score"`
	Accuracy    float64 `json:"hit_accuracy"`
	Playcount   int     `json:"play_count"`
	Playtime    int     `json:"play_time"`
	TotalHits   int     `json:"total_hits"`
	Level       struct {
		Current  int `json:"current"`
		Progress int `json:"progress"`
	} `json:"level"`
}

var Sessions = map[string]*Session{}
var SessionMutex sync.Mutex

func GetSession(username string) *Session {
	SessionMutex.Lock()
	session, ok := Sessions[username]
	if !ok {
		session = &Session{
			Username: username,
		}
		Sessions[username] = session
	}
	SessionMutex.Unlock()
	return session
}

func (s *Session) Destroy() {
	SessionMutex.Lock()
	delete(Sessions, s.Username)
	SessionMutex.Unlock()
}

func (s *Session) Fetch() error {
	data, err := Fetch("/users/" + s.Username)
	if err != nil {
		return err
	}

	var result Result

	if err := json.Unmarshal(data, &result); err != nil {
		return err
	}

	if !result.Online {
		s.Destroy()
		return ErrOffline
	}

	s.Last.Statistics = result.Statistics
	s.Last.Time = time.Now()

	if s.Start.Time.IsZero() {
		s.Start.Statistics = result.Statistics
		s.Start.Time = time.Now()
	}

	s.Session.Time = int(s.Last.Time.Sub(s.Start.Time).Seconds())
	s.Session.GlobalRank = s.Start.Statistics.GlobalRank - s.Last.Statistics.GlobalRank
	s.Session.CountryRank = s.Start.Statistics.CountryRank - s.Last.Statistics.CountryRank
	s.Session.PP = s.Last.Statistics.PP - s.Start.Statistics.PP
	s.Session.RankedScore = s.Last.Statistics.RankedScore - s.Start.Statistics.RankedScore
	s.Session.TotalScore = s.Last.Statistics.TotalScore - s.Start.Statistics.TotalScore
	s.Session.Accuracy = s.Last.Statistics.Accuracy - s.Start.Statistics.Accuracy
	s.Session.Playcount = s.Last.Statistics.Playcount - s.Start.Statistics.Playcount
	s.Session.Playtime = s.Last.Statistics.Playtime - s.Start.Statistics.Playtime
	s.Session.TotalHits = s.Last.Statistics.TotalHits - s.Start.Statistics.TotalHits
	s.Session.Level = s.Last.Statistics.Level.Current - s.Start.Statistics.Level.Current

	// scores, err := Fetch("/users/" + s.Username + "/scores/recent")
	// if err != nil {
	// 	return err
	// }

	return nil
}
