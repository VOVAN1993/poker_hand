package poker

import (
	"bufio"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

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
		Type           TournamentType
		Free           bool
	}
	TournamentType string
)

const (
	BountyHunter TournamentType = "Bounty Hyper"
	Classic      TournamentType = "Big"
	Turbo        TournamentType = "Turbo"
	Hyper        TournamentType = "Hyper"
	TBuilder     TournamentType = "Builder"
	Freeroll     TournamentType = "Freeroll"
	FlipAndGo    TournamentType = "Flip & Go"
	Satellite    TournamentType = "Satellite"
	DeepStacks   TournamentType = "Deep Stacks"
	Shootout     TournamentType = "Shootout"
	Flipout      TournamentType = "Flipout"
)

var (
	biRegexp         = regexp.MustCompile(`([$¥€])([0-9]+(?:\.[0-9]+)?)`)
	totalPrizeRegexp = regexp.MustCompile(`([$¥€])([0-9,]+(?:\.[0-9]+)?)`)
	placeRegexp      = regexp.MustCompile(`(\d+)(?:st|nd|rd|th)? place`)
	reEntriesRegex   = regexp.MustCompile(`You made (\d+) re-entries`)
	myPrizeRegex     = regexp.MustCompile(`received a total of [T,C]?([$¥€])([0-9,]+(?:\.[0-9]+)?)`)
	dateLayout       = "2006/01/02 15:04:05"
)

func ParseTournament(s *bufio.Scanner) (*Tournament, error) {
	/*
		Tournament #183300341, Bounty Hunters Special $2.50 [7-Max], Hold'em No Limit
		Buy-in: $1.3+$0.2+$1
		2245 Players
		Total Prize Pool: $5,163.5
		Tournament started 2025/01/13 12:30:00
		316th : Hero, $1
		You finished the tournament in 316th place.
		You made 1 re-entries and received a total of $1.
	*/
	var t Tournament
	i := -1
	for s.Scan() {
		i++
		switch i {
		case 0:
			name := s.Text()
			id, ttype, err := parseName(name)
			if err != nil {
				var target *SkipTournamentError
				if errors.As(err, &target) {
					fmt.Println("Skip tournament cause ", err.Error())
					return nil, nil
				}
				return nil, err
			}
			t.ID = id
			t.Type = ttype
			t.Name = name
		case 1:
			bi, err := parseBI(s.Text())
			if err != nil {
				return nil, err
			}
			t.BI = bi
		case 2:
			players, err := parsePlayersCount(s.Text())
			if err != nil {
				return nil, err
			}
			t.Players = players
		case 3:
			totalPrize, err := parseTotalPrize(s.Text())
			if err != nil {
				return nil, err
			}
			t.TotalPrizePool = totalPrize
		case 4:
			startTime, err := parseTime(s.Text())
			if err != nil {
				return nil, err
			}
			t.Started = startTime
		case 6:
			place, err := parsePlace(s.Text())
			if err != nil {
				return nil, err
			}
			t.MyPlace = place
		case 7:
			prize, reentry, notFinished, err := parsePrizeAndReentry(s.Text())
			if err != nil {
				return nil, err
			}
			if notFinished {
				return nil, nil
			}
			t.MyPrize = prize
			t.Reentries = reentry
		}
	}
	return &t, nil
}

func parsePrizeAndReentry(s string) (float32, int, bool, error) {
	//You made 1 re-entries and received a total of $1.
	//You received a total of $1.

	reEntriesMatch := reEntriesRegex.FindStringSubmatch(s)
	reEntries := 0
	if reEntriesMatch != nil {
		var err error
		reEntries, err = strconv.Atoi(reEntriesMatch[1])
		if err != nil {
			return 0, 0, false, fmt.Errorf("failed to parse re-entries: %v", err)
		}
	}

	if strings.Contains(s, "You have advanced to") {
		return 0, 0, true, nil
	}

	if strings.Contains(s, "0 chips") {
		return 0, 0, false, nil
	}
	prizeMatch := myPrizeRegex.FindStringSubmatch(s)
	if prizeMatch == nil {
		return 0, 0, false, errors.New("failed to find my prize")
	}

	currency := prizeMatch[1]
	t, err := strconv.ParseFloat(prizeMatch[2], 32)
	value := float32(t)
	if err != nil {
		return 0, 0, false, fmt.Errorf("failed to parse prize value: %v", err)
	}
	if currency == "¥" {
		value = yuanToDollar(value)
	} else if currency == "€" {
		value = euroToDollar(value)
	}
	return value, reEntries, false, nil
}

func parsePlace(s string) (int, error) {
	//You finished the tournament in 12th place.
	match := placeRegexp.FindStringSubmatch(s)
	if match == nil {
		return 0, fmt.Errorf("no placement found in input")
	}
	placement, err := strconv.Atoi(match[1])
	if err != nil {
		return 0, fmt.Errorf("failed to convert placement to int: %v", err)
	}

	return placement, nil
}

func parseTime(s string) (time.Time, error) {
	//Tournament started 2025/01/13 12:30:00
	var date, t string
	n, err := fmt.Sscanf(s, "Tournament started %s %s", &date, &t)
	if err != nil {
		return time.Time{}, err
	}
	if n != 2 {
		return time.Time{}, errors.New("invalid format date line")
	}
	return time.Parse(dateLayout, strings.Join([]string{date, t}, " "))
}

func parseTotalPrize(s string) (float32, error) {
	//Total Prize Pool: $5,163.5
	match := totalPrizeRegexp.FindStringSubmatch(s)
	if match == nil {
		return 0, errors.New("no valid prize pool found in input")
	}

	currency := match[1]
	numberStr := strings.ReplaceAll(match[2], ",", "")
	value, err := strconv.ParseFloat(numberStr, 32)
	if err != nil {
		return 0, fmt.Errorf("failed to parse number: %v", err)
	}
	amount := float32(value)
	if currency == "¥" {
		amount = yuanToDollar(amount)
	} else if currency == "€" {
		amount = euroToDollar(amount)
	}
	return amount, nil
}
func parsePlayersCount(s string) (int, error) {
	//2245 Players
	var count int
	n, err := fmt.Sscanf(s, "%d Players", &count)
	if err != nil {
		return 0, err
	}
	if n != 1 {
		return 0, errors.New("invalid format players line")
	}
	return count, nil
}

func parseBI(s string) (float32, error) {
	//Buy-in: $1.3+$0.2+$1
	matches := biRegexp.FindAllStringSubmatch(s, -1)

	if matches == nil {
		return 0, fmt.Errorf("no valid numbers found in input")
	}

	var sum float32
	var currency string

	for _, match := range matches {
		if currency == "" {
			currency = match[1]
		}
		value, err := strconv.ParseFloat(match[2], 32)
		if err != nil {
			return 0, fmt.Errorf("failed to parse number: %v", err)
		}
		sum += float32(value)
	}
	if currency == "¥" {
		sum = yuanToDollar(sum)
	} else if currency == "€" {
		sum = euroToDollar(sum)
	}
	return sum, nil
}

func parseName(s string) (string, TournamentType, error) {
	//	Tournament #183300341, Bounty Hunters Special $2.50 [7-Max], Hold'em No Limit
	arr := strings.Split(s, ",")
	if len(arr) < 3 {
		return "", "", fmt.Errorf("invalid tournament name: %s", s)
	}
	id := strings.Split(arr[0], "#")[1]
	if id == "" {
		return "", "", fmt.Errorf("Cannot parse tournament id")
	}

	var ttype TournamentType
	if strings.Contains(arr[1], "Bounty") || strings.Contains(arr[1], "Баунти") {
		ttype = BountyHunter
	} else if strings.Contains(arr[1], "Daily Big") || strings.Contains(arr[1], "Sunday Big") ||
		strings.Contains(arr[1], "Daily Special") || strings.Contains(arr[1], "Weekender") {
		ttype = Classic
	} else if strings.Contains(arr[1], string(Turbo)) {
		ttype = Turbo
	} else if strings.Contains(arr[1], string(Hyper)) {
		ttype = Hyper
	} else if strings.Contains(arr[1], string(TBuilder)) {
		ttype = TBuilder
	} else if strings.Contains(arr[1], "Chat&Play") || strings.Contains(arr[1], "ThanksHoldemPlayers") {
		ttype = Freeroll
	} else if strings.Contains(arr[1], "Шутаут") {
		ttype = Shootout
	} else if strings.Contains(arr[1], string(FlipAndGo)) {
		ttype = FlipAndGo
	} else if strings.Contains(s, string(Flipout)) {
		ttype = Flipout
	} else if strings.Contains(arr[1], string(DeepStacks)) || strings.Contains(arr[1], " Monster Stack") {
		ttype = DeepStacks
	} else if strings.Contains(arr[1], string(Satellite)) || strings.Contains(arr[1], string("Step to")) || strings.Contains(arr[1], string("Road to")) {
		ttype = Satellite
	} else {
		fmt.Println("Cannot parse tournament type(set to Classic) : ", arr)
		ttype = Classic
	}

	if strings.TrimSpace(arr[len(arr)-1]) != "Hold'em No Limit" {
		return "", "", &SkipTournamentError{TournamentType(arr[2])}
	}
	return id, ttype, nil
}

func euroToDollar(y float32) float32 {
	return y * 1.04
}

func yuanToDollar(y float32) float32 {
	return y * 0.14
}
