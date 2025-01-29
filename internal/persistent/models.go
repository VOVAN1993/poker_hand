package persistent

import "time"

type (
	Tournament struct {
		ID             string
		BI             float32 //in $
		Players        int
		TotalPrizePool float32 //without rake
		Started        time.Time
		MyPlace        int
		MyPrize        float32
		Reentries      int
		Name           string
		Type           string
	}
)
