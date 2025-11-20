package main

import (
	"fmt"
	"runtime/debug"
	"sync"
	"time"

	"github.com/bytedance/sonic"
	"github.com/google/uuid"
)

type Session struct {
	ID       string
	Username string

	StartTime  time.Time
	LastUpdate time.Time
	Failed     int

	Start   Statistics
	Recent  Statistics
	Session ActiveSession

	mu sync.Mutex
}

type Sessions struct {
	_  noCopy // <<< compile-time safety
	mu sync.RWMutex
	m  map[string]*Session
}

type noCopy struct{}

func (*noCopy) Lock()   {}
func (*noCopy) Unlock() {}

var sessions = NewSessions()

func NewSessions() *Sessions {
	return &Sessions{
		m: make(map[string]*Session),
	}
}

func (s *Sessions) Set(key string, value *Session) {
	s.mu.Lock()
	defer s.mu.Unlock()
	fmt.Printf("SET key=[%s] len=%d\n", key, len(key))
	println("sessions pointer in Set:", sessions)
	println("set pointer:", s)
	fmt.Println("SET callee:", key, "STACK:")
	debug.PrintStack()
	s.m[key] = value
}

func (s *Sessions) Get(key string) *Session {
	s.mu.RLock()
	defer s.mu.RUnlock()
	fmt.Printf("GET key=[%s] len=%d\n", key, len(key))
	println("sessions pointer:", sessions)
	println("get pointer:", s)
	println("nil: ", s.m[key] == nil)

	fmt.Printf("KEY GET raw bytes: %v\n", []byte(key))
	fmt.Printf("MAP HAS KEYS:\n")
	for k := range s.m {
		fmt.Printf("  key=%q bytes=%v\n", k, []byte(k))
	}

	return s.m[key]
}

func (s *Sessions) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	fmt.Printf("DELETE key=[%s] len=%d\n", key, len(key))
	delete(s.m, key)
}

// var scoresSeen sync.Map

func CreateSession(username string) (*Session, error) {
	data, err := Fetch("/users/" + username)
	if err != nil {
		return nil, err
	}

	var result UserResponse
	if err := sonic.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	if !result.Online {
		return nil, ErrOffline
	}

	now := time.Now()

	session := &Session{
		ID:         uuid.NewString(),
		Username:   username,
		StartTime:  now,
		LastUpdate: now,
		Start:      result.Statistics,
		Recent:     result.Statistics,
	}

	fmt.Println("SETTING USERNAME IN CreateSession: ", username)

	sessions.Set(username, session)
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

	if !result.Online {
		//TODO: Change this behaviour
		fmt.Println("Not online")
		if s.Failed >= 10 {
			sessions.Delete(s.Username)
		}
		return ErrOffline
	}

	s.Recent = result.Statistics
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

type Score struct {
	ID       int     `json:"id"`
	PP       float64 `json:"pp"`
	Rank     string  `json:"rank"`
	Accuracy float64 `json:"accuracy"`
}
