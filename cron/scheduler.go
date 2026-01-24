package cron

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/go-co-op/gocron/v2"
	db "github.com/stuneak/sopeko/db/sqlc"
)

var slog = log.New(log.Writer(), "[CRON] ", log.Flags())

var subreddits = []string{
	"wallstreetbets",
	"pennystocks",
	"investing",
	"stocks",
}

type Scheduler struct {
	scheduler     gocron.Scheduler
	store         *db.Queries
	redditScraper *RedditScraper
	nasdaqFetcher *NasdaqFetcher
	tickerParser  *TickerParser
}

func NewScheduler(db *db.Queries) (*Scheduler, error) {
	usEastern, err := time.LoadLocation("America/New_York")
	if err != nil {
		return nil, err
	}

	s, err := gocron.NewScheduler(gocron.WithLocation(usEastern))
	if err != nil {
		return nil, err
	}

	return &Scheduler{
		scheduler:     s,
		store:         db,
		redditScraper: NewRedditScraper(),
		nasdaqFetcher: NewNasdaqFetcher(),
		tickerParser:  NewTickerParser(db),
	}, nil
}

func (s *Scheduler) scrapeSubreddit(subreddit string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()
	slog.Printf("Starting Reddit scrape for r/%s", subreddit)

	posts, comments, err := s.redditScraper.ScrapeSubreddit(ctx, subreddit)
	if err != nil {
		slog.Printf("Error scraping r/%s: %v", subreddit, err)
		return
	}

	slog.Printf("Scraped r/%s: %d posts, %d comments", subreddit, len(posts), len(comments))

	// Process posts for ticker mentions
	if err := s.tickerParser.ProcessPosts(ctx, posts); err != nil {
		slog.Printf("Error processing posts for r/%s: %v", subreddit, err)
	}

	// Process comments for ticker mentions
	if err := s.tickerParser.ProcessComments(ctx, comments); err != nil {
		slog.Printf("Error processing comments for r/%s: %v", subreddit, err)
	}

	slog.Printf("Finished processing r/%s", subreddit)
}

func (s *Scheduler) syncTickers() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	slog.Println("syncTickers: job triggered, starting NASDAQ tickers sync")

	stocks, err := s.nasdaqFetcher.FetchStocks(ctx)
	if err != nil {
		slog.Printf("syncTickers: error fetching NASDAQ stocks: %v", err)
		return
	}

	slog.Printf("syncTickers: fetched %d stocks from NASDAQ", len(stocks))

	var synced, skipped int
	for _, stock := range stocks {
		if strings.Contains(stock.Symbol, "^") || strings.Contains(stock.Symbol, "/") {
			skipped++
			continue
		}
		err := s.store.UpsertTicker(ctx, db.UpsertTickerParams{
			Symbol:      stock.Symbol,
			CompanyName: stock.Name,
			Exchange:    "NASDAQ",
		})
		if err != nil {
			slog.Printf("syncTickers: error upserting ticker %s: %v", stock.Symbol, err)
			continue
		}
		synced++
	}

	slog.Printf("syncTickers: done - synced %d, skipped %d (contained ^ or /)", synced, skipped)
}

func (s *Scheduler) fetchAllTickerPrices() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	slog.Println("fetchAllTickerPrices: job triggered, starting ticker prices fetch")

	tickers, err := s.store.ListAllTickers(ctx)
	if err != nil {
		slog.Printf("fetchAllTickerPrices: error fetching tickers from DB: %v", err)
		return
	}

	slog.Printf("fetchAllTickerPrices: fetching prices for %d tickers", len(tickers))

	now := time.Now()
	var fetched int
	var errorSymbols []string

	for i, ticker := range tickers {
		if strings.Contains(ticker.Symbol, "^") || strings.Contains(ticker.Symbol, "/") {
			continue
		}

		if i > 0 && i%100 == 0 {
			slog.Printf("fetchAllTickerPrices: progress %d/%d tickers processed (%d fetched, %d errors so far)", i, len(tickers), fetched, len(errorSymbols))
		}

		price, volume, priceTime, err := s.tickerParser.fetchHistoricalPrice(ctx, ticker.Symbol, now)
		if err != nil {
			if len(errorSymbols) < 10 {
				slog.Printf("fetchAllTickerPrices: error fetching price for %s: %v", ticker.Symbol, err)
			}
			errorSymbols = append(errorSymbols, ticker.Symbol)
			continue
		}

		slog.Printf("fetchAllTickerPrices: inserting price for %s: price=%v, volume=%v, time=%s", ticker.Symbol, price, volume, priceTime.Format(time.RFC3339))

		_, err = s.store.InsertTickerPrice(ctx, db.InsertTickerPriceParams{
			TickerID:   ticker.ID,
			Price:      price,
			Volume:     volume,
			RecordedAt: priceTime,
		})
		if err != nil {
			slog.Printf("fetchAllTickerPrices: error inserting price for %s (tickerID=%d): %v", ticker.Symbol, ticker.ID, err)
			errorSymbols = append(errorSymbols, ticker.Symbol)
			continue
		}
		slog.Printf("fetchAllTickerPrices: successfully inserted price for %s", ticker.Symbol)
		fetched++
	}

	slog.Printf("fetchAllTickerPrices: done - %d fetched, %d errors", fetched, len(errorSymbols))
	if len(errorSymbols) > 0 {
		slog.Printf("fetchAllTickerPrices: error symbols: %v", errorSymbols)
	}
}

func (s *Scheduler) fetchAllSplits() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	slog.Println("fetchAllSplits: job triggered, starting stock splits fetch")

	tickers, err := s.store.ListAllTickers(ctx)
	if err != nil {
		slog.Printf("fetchAllSplits: error fetching tickers from DB: %v", err)
		return
	}

	slog.Printf("fetchAllSplits: processing %d tickers for splits", len(tickers))

	var fetched, fetchErrors, insertErrors int
	for i, ticker := range tickers {
		if strings.Contains(ticker.Symbol, "^") || strings.Contains(ticker.Symbol, "/") {
			continue
		}

		if i > 0 && i%100 == 0 {
			slog.Printf("fetchAllSplits: progress %d/%d tickers processed (%d splits found, %d fetch errors, %d insert errors)", i, len(tickers), fetched, fetchErrors, insertErrors)
		}

		splits, err := s.tickerParser.FetchSplits(ctx, ticker.Symbol)
		if err != nil {
			fetchErrors++
			if fetchErrors <= 10 {
				slog.Printf("fetchAllSplits: error fetching splits for %s: %v", ticker.Symbol, err)
			}
			continue
		}

		slog.Printf("Found %d splits for %s", len(splits), ticker.Symbol)

		for _, split := range splits {
			slog.Printf("fetchAllSplits: inserting split for %s on %s (ratio %.4f)", ticker.Symbol, split.EffectiveDate.Format("2006-01-02"), split.Ratio)
			err = s.store.InsertStockSplit(ctx, db.InsertStockSplitParams{
				TickerID:      ticker.ID,
				Ratio:         fmt.Sprintf("%.4f", split.Ratio),
				EffectiveDate: split.EffectiveDate,
			})
			if err != nil {
				slog.Printf("fetchAllSplits: error inserting split for %s: %v", ticker.Symbol, err)
				insertErrors++
				continue
			}
			fetched++
		}
	}

	slog.Printf("fetchAllSplits: done - %d splits stored, %d fetch errors, %d insert errors (out of %d tickers)", fetched, fetchErrors, insertErrors, len(tickers))
}

func (s *Scheduler) RegisterJobs() error {
	now := time.Now()
	slog.Printf("RegisterJobs: registering all jobs at %s (location: %s)", now.Format(time.RFC3339), now.Location())

	// NASDAQ tickers sync - once daily
	tickerSyncStart := now.Add(5 * time.Second)
	_, err := s.scheduler.NewJob(
		gocron.DurationJob(24*time.Hour),
		gocron.NewTask(func() {
			slog.Println("RegisterJobs: nasdaq-tickers-sync job fired")
			s.syncTickers()
		}),
		gocron.WithName("nasdaq-tickers-sync"),
		gocron.WithStartAt(gocron.WithStartDateTime(tickerSyncStart)),
	)
	if err != nil {
		slog.Printf("RegisterJobs: error registering nasdaq-tickers-sync: %v", err)
		return err
	}

	slog.Printf("RegisterJobs: registered nasdaq-tickers-sync (every 24h, first run at %s)", tickerSyncStart.Format(time.RFC3339))

	// Ticker prices fetch - every 2 hours, first run after 2h
	pricesStart := now.Add(2 * time.Hour)
	_, err = s.scheduler.NewJob(
		gocron.DurationJob(2*time.Hour),
		gocron.NewTask(func() {
			slog.Println("RegisterJobs: ticker-prices job fired")
			s.fetchAllTickerPrices()
		}),
		gocron.WithName("ticker-prices"),
		gocron.WithStartAt(gocron.WithStartDateTime(pricesStart)),
	)
	if err != nil {
		slog.Printf("RegisterJobs: error registering ticker-prices: %v", err)
		return err
	}

	slog.Printf("RegisterJobs: registered ticker-prices (every 2h, first run at %s)", pricesStart.Format(time.RFC3339))

	// Stock splits fetch - every 12h
	splitsStart := now.Add(12 * time.Hour)
	_, err = s.scheduler.NewJob(
		gocron.DurationJob(12*time.Hour),
		gocron.NewTask(func() {
			slog.Println("RegisterJobs: stock-splits job fired")
			s.fetchAllSplits()
		}),
		gocron.WithName("stock-splits"),
		gocron.WithStartAt(gocron.WithStartDateTime(splitsStart)),
	)
	if err != nil {
		slog.Printf("RegisterJobs: error registering stock-splits: %v", err)
		return err
	}

	slog.Printf("RegisterJobs: registered stock-splits (every 12h, first run at %s)", splitsStart.Format(time.RFC3339))

	// Reddit scraping - every 1h, staggered
	startDelays := []time.Duration{1 * time.Hour, 2 * time.Hour, 2 * time.Hour, 2 * time.Hour}
	for i, subreddit := range subreddits {
		sub := subreddit
		subStart := now.Add(startDelays[i])

		_, err := s.scheduler.NewJob(
			gocron.DurationJob(1*time.Hour),
			gocron.NewTask(func() {
				slog.Printf("RegisterJobs: reddit-scrape-%s job fired", sub)
				s.scrapeSubreddit(sub)
			}),
			gocron.WithName("reddit-scrape-"+sub),
			gocron.WithStartAt(gocron.WithStartDateTime(subStart)),
		)
		if err != nil {
			slog.Printf("RegisterJobs: error registering reddit-scrape-%s: %v", sub, err)
			return err
		}

		slog.Printf("RegisterJobs: registered reddit-scrape-%s (every 1h, first run at %s)", sub, subStart.Format(time.RFC3339))
	}

	slog.Printf("RegisterJobs: all %d jobs registered successfully", 3+len(subreddits))
	return nil
}

func (s *Scheduler) Start() {
	slog.Printf("Starting scheduler with %d Reddit scraping jobs...", len(subreddits))
	s.scheduler.Start()
}

func (s *Scheduler) Stop() error {
	slog.Println("Stopping scheduler...")
	return s.scheduler.Shutdown()
}
