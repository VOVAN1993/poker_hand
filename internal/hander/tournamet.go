package hander

import (
	"context"
	"fmt"

	"github.com/VOVAN1993/poker_hand/internal/persistent"
	"github.com/VOVAN1993/poker_hand/internal/poker"
)

func (h *hander) GetTournament(ctx context.Context, id string) (poker.Tournament, error) {
	tournaments, err := h.ps.ListTournaments(ctx, []persistent.WhereOpt{persistent.WithID(id)}...)
	if err != nil {
		return poker.Tournament{}, err
	}
	if len(tournaments) == 0 {
		return poker.Tournament{}, fmt.Errorf("not found tournament #%s", id)
	}
	if len(tournaments) > 1 {
		return poker.Tournament{}, fmt.Errorf("found some tournaments with id #%s", id)
	}
	return castTournamentFromDB(&tournaments[0]), nil
}

func (h *hander) ListTournaments(ctx context.Context) ([]poker.Tournament, error) {
	tournaments, err := h.ps.ListTournaments(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]poker.Tournament, 0, len(tournaments))
	for _, t := range tournaments {
		res = append(res, castTournamentFromDB(&t))
	}
	return res, nil
}

func (h *hander) FreeTournament(ctx context.Context, id string) error {
	ok, err := h.ps.FreeTournament(ctx, id)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("not found tournament #%s", id)
	}
	return nil
}
