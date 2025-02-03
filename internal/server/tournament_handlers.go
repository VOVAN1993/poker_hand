package server

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/types"
)

func (s *Server) tournamentsHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		ts, err := s.handManager.ListTournaments(r.Context())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(fmt.Sprintf("Server error: %s", err)))
			return
		}
		RespondJSON(w, http.StatusOK, ts)
	}
}

func (s *Server) freeTournament() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		id := r.PathValue("id")
		if err := s.handManager.FreeTournament(r.Context(), id); err != nil {
			RespondError(w, http.StatusNotFound, err.Error())
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func (s *Server) tournamentHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		id := r.PathValue("id")
		t, err := s.handManager.GetTournament(r.Context(), id)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(fmt.Sprintf("Server error: %s", err)))
			return
		}
		RespondJSON(w, http.StatusOK, t)
	}
}

func (s *Server) roi() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		tournaments, err := s.handManager.ListTournaments(r.Context())
		if err != nil {
			RespondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if len(tournaments) == 0 {
			return
		}
		sort.Slice(tournaments, func(i, j int) bool {
			return tournaments[i].Started.Before(tournaments[j].Started)
		})

		line := charts.NewLine()
		line.SetGlobalOptions(
			charts.WithInitializationOpts(opts.Initialization{Theme: types.ThemeInfographic}),
			charts.WithTitleOpts(opts.Title{
				Title:    "ROI",
				Subtitle: "Изменение ROI от количества турниров",
			}),
			charts.WithYAxisOpts(opts.YAxis{
				Min: opts.Float(-50),
				Max: opts.Float(150),
			}),
		)

		genValues := func() []opts.LineData {
			items := make([]opts.LineData, 0)
			var curBI, curPrize float32
			for i := 0; i < len(tournaments); i++ {
				if tournaments[i].BI < 0.2 {
					continue
				}
				if !tournaments[i].Free {
					curBI += tournaments[i].BI
				}
				fmt.Println(tournaments[i].BI, tournaments[i].MyPrize)
				curPrize += tournaments[i].MyPrize

				pointRoi := 100 * ((curPrize - curBI) / curBI)
				fmt.Println("roi ", pointRoi)
				fmt.Println(curPrize, curBI)
				items = append(items, opts.LineData{Value: strconv.FormatFloat(float64(pointRoi), 'f', 2, 32)})
			}
			return items
		}
		values := genValues()
		xaxis := make([]int, len(values))
		for i := 0; i < len(values); i++ {
			xaxis[i] = i + 1
		}
		line.SetXAxis(xaxis).
			AddSeries("Current ROI", values).
			SetSeriesOptions(charts.WithLineChartOpts(opts.LineChart{Smooth: opts.Bool(true)}),
				charts.WithMarkPointNameTypeItemOpts(
					opts.MarkPointNameTypeItem{Name: "Точка", Type: "circle"},
				))
		line.Render(w)
	}
}

func (s *Server) plot() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		tournaments, err := s.handManager.ListTournaments(r.Context())
		if err != nil {
			RespondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if len(tournaments) == 0 {
			return
		}
		sort.Slice(tournaments, func(i, j int) bool {
			return tournaments[i].Started.Before(tournaments[j].Started)
		})

		cur := tournaments[0]
		curi := 0
		arr := make([]float32, 1)
		dates := make([]string, 0)
		dates = append(dates, cur.Started.Format("01.02.2006"))
		for _, t := range tournaments {
			if !(cur.Started.Year() == t.Started.Year() &&
				cur.Started.YearDay() == t.Started.YearDay()) {
				curi += 1
				cur = t
				arr = append(arr, 0.0)
				dates = append(dates, cur.Started.Format("01.02.2006"))
			}
			if t.Free {
				arr[curi] = arr[curi] + t.MyPrize
			} else {
				arr[curi] = arr[curi] + t.MyPrize - t.BI
			}
		}
		//// Put data into instance
		genValues := func() []opts.LineData {
			items := make([]opts.LineData, len(dates))
			value := float32(0)
			for i := 0; i < len(dates); i++ {
				value += arr[i]
				items[i] = opts.LineData{Value: strconv.FormatFloat(float64(value), 'f', 2, 32)}
			}
			return items
		}
		line := charts.NewLine()
		// set some global options like Title/Legend/ToolTip or anything else
		line.SetGlobalOptions(
			charts.WithInitializationOpts(opts.Initialization{Theme: types.ThemeInfographic}),
			charts.WithTitleOpts(opts.Title{
				Title:    "BR",
				Subtitle: "Изменение BR по датам",
			}),
		)
		line.SetXAxis(dates).
			AddSeries("Current BR", genValues()).
			SetSeriesOptions(charts.WithLineChartOpts(opts.LineChart{Smooth: opts.Bool(true)}),
				charts.WithMarkPointNameTypeItemOpts(
					opts.MarkPointNameTypeItem{Name: "Точка", Type: "circle"},
				))
		line.Render(w)
	}
}
