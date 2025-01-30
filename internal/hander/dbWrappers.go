package hander

import (
	"github.com/VOVAN1993/poker_hand/internal/persistent"
	"github.com/VOVAN1993/poker_hand/internal/poker"
)

func castTournamentToDB(t *poker.Tournament) persistent.Tournament {
	return persistent.Tournament{
		ID:             t.ID,
		BI:             t.BI,
		Players:        t.Players,
		TotalPrizePool: t.TotalPrizePool,
		Started:        t.Started,
		MyPlace:        t.MyPlace,
		MyPrize:        t.MyPrize,
		Reentries:      t.Reentries,
		Name:           t.Name,
		Type:           string(t.Type),
	}
}

func castTournamentFromDB(t *persistent.Tournament) poker.Tournament {
	return poker.Tournament{
		ID:             t.ID,
		BI:             t.BI,
		Players:        t.Players,
		TotalPrizePool: t.TotalPrizePool,
		Started:        t.Started,
		MyPlace:        t.MyPlace,
		MyPrize:        t.MyPrize,
		Reentries:      t.Reentries,
		Name:           t.Name,
		Type:           poker.TournamentType(t.Type),
	}
}
