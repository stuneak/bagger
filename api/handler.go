package api

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/stuneak/sopeko/db/sqlc"
)

// Excluded usernames (mods, bots, special accounts)
var excludedUsernames = []string{
	"OhWowMuchFunYouGuys",
	"miamihausjunkie",
	"AnnArchist",
	"Immersions-",
	"butthoofer",
	"AutoModerator",
	"PennyBotWeekly",
	"PennyPumper",
	"TransSpeciesDog",
	"the_male_nurse",
	"VisualMod",
	"OPINION_IS_UNPOPULAR",
	"zjz",
	"OSRSkarma",
	"Dan_inKuwait",
	"Swiifttx",
	"teddy_riesling",
	"Stylux",
	"Latter-day_weeb",
	"ShopBitter",
	"CHAINSAW_VASECTOMY",
}

type MentionResponse struct {
	Symbol           string    `json:"symbol"`
	MentionPrice     string    `json:"mention_price"`
	CurrentPrice     string    `json:"current_price"`
	CurrentPriceDate time.Time `json:"current_price_date"`
	PercentChange    string    `json:"percent_change"`
	SplitRatio       float64   `json:"split_ratio"`
	MentionedAt      time.Time `json:"mentioned_at"`
}

func (server *Server) getUserMentions(ctx *gin.Context) {
	username := ctx.Param("username")

	mentions, err := server.store.GetFirstMentionPerTickerByUsername(ctx, username)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if len(mentions) == 0 {
		ctx.JSON(http.StatusOK, []MentionResponse{})
		return
	}

	results := []MentionResponse{}

	for _, mention := range mentions {
		adjustedMentionPrice := mention.MentionPrice
		if mention.CalculatedSplitRatio != 1.0 {
			mp, err := strconv.ParseFloat(mention.MentionPrice, 64)
			if err == nil {
				adjustedMentionPrice = fmt.Sprintf("%.2f", mp*mention.CalculatedSplitRatio)
			}
		}

		percentChange := calculatePercentChange(adjustedMentionPrice, mention.CurrentPrice)
		results = append(results, MentionResponse{
			Symbol:           mention.Symbol,
			MentionPrice:     adjustedMentionPrice,
			CurrentPrice:     mention.CurrentPrice,
			CurrentPriceDate: mention.CurrentPriceDate,
			PercentChange:    percentChange,
			SplitRatio:       mention.CalculatedSplitRatio,
			MentionedAt:      mention.MentionedAt,
		})
	}

	ctx.JSON(http.StatusOK, results)
}

func calculatePercentChange(oldPrice, newPrice string) string {
	old, err := strconv.ParseFloat(oldPrice, 64)
	if err != nil || old == 0 {
		return "0.00%"
	}
	new, err := strconv.ParseFloat(newPrice, 64)
	if err != nil {
		return "0.00%"
	}

	change := ((new - old) / old) * 100
	if change >= 0 {
		return fmt.Sprintf("+%.2f%%", change)
	}
	return fmt.Sprintf("%.2f%%", change)
}

func (server *Server) getExcludedUsernames(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, excludedUsernames)
}

type PickDetail struct {
	Symbol       string  `json:"symbol"`
	PickPrice    string  `json:"pick_price"`
	CurrentPrice string  `json:"current_price"`
	PercentGain  float64 `json:"percent_gain"`
	SplitRatio   float64 `json:"split_ratio"`
}

type TopUserResponse struct {
	Username         string       `json:"username"`
	TotalPercentGain float64      `json:"total_percent_gain"`
	Picks            []PickDetail `json:"picks"`
}

type PickPerformanceResponse struct {
	Symbol           string    `json:"symbol"`
	MentionPrice     string    `json:"mention_price"`
	CurrentPrice     string    `json:"current_price"`
	CurrentPriceDate time.Time `json:"current_price_date"`
	PercentChange    float64   `json:"percent_change"`
	SplitRatio       float64   `json:"split_ratio"`
	MentionedAt      time.Time `json:"mentioned_at"`
}

func (server *Server) getTopPerformingPicks(ctx *gin.Context) {
	server.getPerformingPicks(ctx, true)
}

func (server *Server) getWorstPerformingPicks(ctx *gin.Context) {
	server.getPerformingPicks(ctx, false)
}

func (server *Server) getPerformingPicks(ctx *gin.Context, topPerformers bool) {
	// Parse time period filter: "weekly", "monthly", or empty for all time
	period := ctx.Query("period")
	var cutoffTime time.Time
	now := time.Now()
	switch period {
	case "daily":
		cutoffTime = now.AddDate(0, 0, -1)
	case "weekly":
		cutoffTime = now.AddDate(0, 0, -7)
	case "monthly":
		cutoffTime = now.AddDate(0, -1, 0)
	default:
		cutoffTime = time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	}

	picks, err := server.store.GetAllPicksWithPricesAndSplitsSince(ctx, cutoffTime)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	excludedMap := make(map[string]bool)
	for _, u := range excludedUsernames {
		excludedMap[u] = true
	}

	// Track first mention per ticker (unique across all users)
	firstMentionPerTicker := make(map[int64]struct {
		symbol               string
		mentionPrice         string
		currentPrice         string
		currentPriceDate     time.Time
		calculatedSplitRatio float64
		mentionedAt          time.Time
	})

	for _, pick := range picks {
		if excludedMap[pick.Username] {
			continue
		}
		if existing, ok := firstMentionPerTicker[pick.TickerID]; !ok || pick.MentionedAt.Before(existing.mentionedAt) {
			firstMentionPerTicker[pick.TickerID] = struct {
				symbol               string
				mentionPrice         string
				currentPrice         string
				currentPriceDate     time.Time
				calculatedSplitRatio float64
				mentionedAt          time.Time
			}{
				symbol:               pick.Symbol,
				mentionPrice:         pick.MentionPrice,
				currentPrice:         pick.CurrentPrice,
				currentPriceDate:     pick.CurrentPriceDate,
				calculatedSplitRatio: pick.CalculatedSplitRatio,
				mentionedAt:          pick.MentionedAt,
			}
		}
	}

	var results []PickPerformanceResponse

	for _, pick := range firstMentionPerTicker {
		mentionPrice, err := strconv.ParseFloat(pick.mentionPrice, 64)
		if err != nil || mentionPrice <= 0 {
			continue
		}

		currPrice, err := strconv.ParseFloat(pick.currentPrice, 64)
		if err != nil {
			continue
		}

		adjustedMentionPrice := mentionPrice
		if pick.calculatedSplitRatio != 1.0 {
			mp, err := strconv.ParseFloat(pick.mentionPrice, 64)
			if err == nil {
				adjustedMentionPrice = mp * pick.calculatedSplitRatio
			}
		}

		percentChange := ((currPrice - adjustedMentionPrice) / adjustedMentionPrice) * 100
		results = append(results, PickPerformanceResponse{
			Symbol:           pick.symbol,
			MentionPrice:     fmt.Sprintf("%.2f", adjustedMentionPrice),
			CurrentPrice:     pick.currentPrice,
			CurrentPriceDate: pick.currentPriceDate,
			PercentChange:    percentChange,
			SplitRatio:       pick.calculatedSplitRatio,
			MentionedAt:      pick.mentionedAt,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		if topPerformers {
			return results[i].PercentChange > results[j].PercentChange
		}
		return results[i].PercentChange < results[j].PercentChange
	})

	if len(results) > 10 {
		results = results[:10]
	}

	ctx.JSON(http.StatusOK, results)
}

func (server *Server) getTopPerformingUsers(ctx *gin.Context) {
	// Parse time period filter: "weekly", "monthly", or empty for all time
	period := ctx.Query("period")
	var cutoffTime time.Time
	now := time.Now()
	switch period {
	case "daily":
		cutoffTime = now.AddDate(0, 0, -1)
	case "weekly":
		cutoffTime = now.AddDate(0, 0, -7)
	case "monthly":
		cutoffTime = now.AddDate(0, -1, 0)
	default:
		// No filter - use a very old date to include all
		cutoffTime = time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	}

	picks, err := server.store.GetUniqueUserPicksWithLatestPricesSince(ctx, cutoffTime)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	excludedMap := make(map[string]bool)
	for _, u := range excludedUsernames {
		excludedMap[u] = true
	}

	users := make(map[string]*TopUserResponse)

	for _, pick := range picks {
		if excludedMap[pick.Username] {
			continue
		}

		pickPrice, _ := strconv.ParseFloat(pick.MentionPrice, 64)
		currPrice, _ := strconv.ParseFloat(pick.CurrentPrice, 64)

		adjustedPickPrice := pickPrice
		if pick.CalculatedSplitRatio != 1.0 {
			mp, err := strconv.ParseFloat(pick.MentionPrice, 64)
			if err == nil {
				adjustedPickPrice = mp * pick.CalculatedSplitRatio
			}
		}

		percentGain := ((currPrice - adjustedPickPrice) / adjustedPickPrice) * 100

		if percentGain == 0 {
			continue
		}

		if users[pick.Username] == nil {
			users[pick.Username] = &TopUserResponse{
				Username: pick.Username,
				Picks:    []PickDetail{},
			}
		}

		users[pick.Username].TotalPercentGain += percentGain
		users[pick.Username].Picks = append(users[pick.Username].Picks, PickDetail{
			Symbol:       pick.Symbol,
			PickPrice:    fmt.Sprintf("%.2f", adjustedPickPrice),
			CurrentPrice: pick.CurrentPrice,
			PercentGain:  percentGain,
			SplitRatio:   pick.CalculatedSplitRatio,
		})
	}

	results := make([]TopUserResponse, 0, len(users))
	for _, stats := range users {
		results = append(results, *stats)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].TotalPercentGain > results[j].TotalPercentGain
	})

	if len(results) > 10 {
		results = results[:10]
	}

	for i := range results {
		sort.Slice(results[i].Picks, func(a, b int) bool {
			return results[i].Picks[a].PercentGain > results[i].Picks[b].PercentGain
		})
	}

	ctx.JSON(http.StatusOK, results)
}
