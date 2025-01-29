package poker

import "fmt"

type (
	SkipTournamentError struct {
		TournamentType TournamentType
	}
)

func (err *SkipTournamentError) Error() string {
	return fmt.Sprintf("skip tournament type: %s", err.TournamentType)
}
