package hander

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/VOVAN1993/poker_hand/internal/persistent"
	"github.com/VOVAN1993/poker_hand/internal/poker"
)

type (
	HandManager interface {
		Start(ctx context.Context) error
		Stop()

		ListTournaments(ctx context.Context) ([]poker.Tournament, error)
		GetTournament(ctx context.Context, id string) (poker.Tournament, error)
		FreeTournament(ctx context.Context, id string) error
	}
	hander struct {
		ps persistent.Persistent
	}
)

func NewHandManager() HandManager {
	db := persistent.NewPersistent()
	return &hander{ps: db}
}

func (h *hander) parseTournament(path string) (*poker.Tournament, error) {
	if !strings.HasSuffix(path, ".txt") {
		return nil, nil
	}
	readFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	fileScanner := bufio.NewScanner(readFile) //default delimiter - ScanLines
	t, err := poker.ParseTournament(fileScanner)
	return t, err
}

func (h *hander) parseTournaments(ctx context.Context) error {
	tournamentDir := os.Getenv("DB_TOURNAMENT_DIR")
	baseDir := os.Getenv("DB_BASE_DIR")
	if tournamentDir == "" {
		return errors.New("DB_TOURNAMENT_DIR environment variable not set")
	}

	tournaments := make([]*poker.Tournament, 0)
	err := filepath.Walk(path.Join(baseDir, tournamentDir), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Println(err)
			return err
		}
		if info.IsDir() || !strings.HasSuffix(path, ".txt") {
			return nil
		}
		//fmt.Println("Found Tournament ", path, info.Name())
		t, err := h.parseTournament(path)
		if err != nil {
			fmt.Println("got error during parsing tournament:", err)
			return err
		}
		if t == nil {
			return nil
		}
		tournaments = append(tournaments, t)
		return nil
	})
	if err != nil {
		fmt.Println(err)
	}
	newTournaments := 0
	for _, t := range tournaments {
		ok, err := h.ps.SaveTournaments(ctx, castTournamentToDB(t))
		if err != nil {
			return err
		}
		if ok {
			newTournaments++
		}
	}
	fmt.Printf("Saved %d tournamets\n", newTournaments)
	return nil
}

func (h *hander) Start(ctx context.Context) error {
	if err := h.ps.Start(ctx); err != nil {
		return err
	}

	if err := h.ps.CreateTournamentsTable(ctx); err != nil {
		return err
	}
	return h.parseTournaments(ctx)
}

func (h *hander) Stop() {
	h.ps.Stop()
}
