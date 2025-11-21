package main

import (
	"sync"
	"time"

	"github.com/bytedance/sonic"
)

type Session struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Online   bool   `json:"online"`

	StartTime  time.Time `json:"start_time"`
	LastUpdate time.Time `json:"last_update"`
	LastScore  time.Time `json:"last_score"`

	Start   Statistics    `json:"start"`
	Recent  Statistics    `json:"recent"`
	Session ActiveSession `json:"session"`

	mu sync.Mutex
}

type Sessions struct {
	mu sync.RWMutex
	m  map[int]*Session
}

var sessions = &Sessions{
	m: make(map[int]*Session),
}

func (s *Sessions) Set(key int, value *Session) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[key] = value
}

func (s *Sessions) Get(key int) *Session {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.m[key]
}

func (s *Sessions) Delete(key int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.m, key)
}

func CreateSession(username string) (*Session, error) {
	data, err := Fetch("/users/" + username)
	if err != nil {
		return nil, err
	}

	var result UserResponse
	if err := sonic.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	now := time.Now()

	session := &Session{
		ID:         result.ID,
		Username:   username,
		Online:     result.Online,
		StartTime:  now,
		LastUpdate: now,
		Start:      result.Statistics,
		Recent:     result.Statistics,
	}

	sessions.Set(result.ID, session)
	users.Set(username, result.ID)
	return session, nil
}

func (s *Session) Update() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := Fetch("/users/" + s.Username)
	if err != nil {
		return err
	}

	var result UserResponse
	if err := sonic.Unmarshal(data, &result); err != nil {
		return err
	}

	s.Recent = result.Statistics
	s.Online = result.Online

	if !result.Online {
		return nil
	}

	s.LastUpdate = time.Now()
	s.Session.Time = int(s.LastUpdate.Sub(s.StartTime).Seconds())
	s.Session.GlobalRank = s.Start.GlobalRank - s.Recent.GlobalRank
	s.Session.CountryRank = s.Start.CountryRank - s.Recent.CountryRank
	s.Session.PP = s.Recent.PP - s.Start.PP
	s.Session.RankedScore = s.Recent.RankedScore - s.Start.RankedScore
	s.Session.TotalScore = s.Recent.TotalScore - s.Start.TotalScore
	s.Session.Accuracy = s.Recent.Accuracy - s.Start.Accuracy
	s.Session.Playcount = s.Recent.Playcount - s.Start.Playcount
	s.Session.Playtime = s.Recent.Playtime - s.Start.Playtime
	s.Session.TotalHits = s.Recent.TotalHits - s.Start.TotalHits
	s.Session.Level = s.Recent.Level.Current - s.Start.Level.Current

	return nil
}

type UserResponse struct {
	ID         int        `json:"id"`
	Username   string     `json:"username"`
	Online     bool       `json:"is_online"`
	LastVisit  time.Time  `json:"last_visit"`
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
	Passed      int     `json:"passed"`
	Ranks       Ranks   `json:"ranks"`
	MaxCombo    int     `json:"max_combo"`
	FCs         int     `json:"fcs"`
	AccuracyAvg float64 `json:"avg_accuracy"`
	accuracies  float64
}

type Ranks struct {
	XH int `json:"xh"`
	X  int `json:"x"`
	SH int `json:"sh"`
	S  int `json:"s"`
	A  int `json:"a"`
	B  int `json:"b"`
	C  int `json:"c"`
	D  int `json:"d"`
	F  int `json:"f"`
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

func ClearSessions() {
	for k, v := range sessions.m {
		if !v.Online {
			if time.Since(v.LastUpdate) > time.Minute*15 {
				sessions.Delete(k)
				users.Delete(v.Username)
			}
		} else {
			if time.Since(v.LastScore) > time.Hour*6 && time.Since(v.LastUpdate) > time.Minute*15 {
				sessions.Delete(k)
				users.Delete(v.Username)
			}
		}
	}
}
