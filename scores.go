package main

import (
	"github.com/bytedance/sonic"
)

type Score struct {
	ID       int     `json:"id"`
	PP       float64 `json:"pp"`
	Rank     string  `json:"rank"`
	Accuracy float64 `json:"accuracy"`
	UserID   int     `json:"user_id"`
	Passed   bool    `json:"passed"`
	MaxCombo int     `json:"max_combo"`
	Perfect  bool    `json:"is_perfect_combo"`
}

type ScoresResponse struct {
	Scores       []Score `json:"scores"`
	CursorString string  `json:"cursor_string"`
}

var cursorString string

func GetScores() error {
	data, err := Fetch("/scores?cursor_string=" + cursorString)
	if err != nil {
		return err
	}

	var response ScoresResponse
	if err := sonic.Unmarshal(data, &response); err != nil {
		return err
	}

	cursorString = response.CursorString

	for _, score := range response.Scores {
		session := sessions.Get(score.UserID)
		if session == nil {
			continue
		}

		session.Session.Scores += 1

		session.Session.accuracies += score.Accuracy
		session.Session.AccuracyAvg = session.Session.accuracies / float64(session.Session.Scores)
		if score.Passed {
			session.Session.Passed += 1
		}

		session.Session.RawPP += score.PP
		if score.MaxCombo > session.Session.MaxCombo {
			session.Session.MaxCombo = score.MaxCombo
		}

		if score.Perfect {
			session.Session.FCs += 1
		}

		switch score.Rank {
		case "XH":
			session.Session.Ranks.XH++
		case "X":
			session.Session.Ranks.X++
		case "SH":
			session.Session.Ranks.SH++
		case "S":
			session.Session.Ranks.S++
		case "A":
			session.Session.Ranks.A++
		case "B":
			session.Session.Ranks.B++
		case "C":
			session.Session.Ranks.C++
		case "D":
			session.Session.Ranks.D++
		case "F":
			session.Session.Ranks.F++
		}

		session.Update()
	}

	return nil
}
