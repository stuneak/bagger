package cron

import (
	"context"
	"log"
	"time"

	"github.com/go-co-op/gocron/v2"
	db "github.com/stuneak/bagger/db/sqlc"
)

var subreddits = []string{
	// "pennystocks",
	// "investing",
	// "stocks",
	"wallstreetbets",
}

type Scheduler struct {
	scheduler     gocron.Scheduler
	store         *db.Queries
	redditScraper *RedditScraper
}

func NewScheduler(db *db.Queries) (*Scheduler, error) {
	s, err := gocron.NewScheduler()
	if err != nil {
		return nil, err
	}

	return &Scheduler{
		scheduler:     s,
		store:         db,
		redditScraper: NewRedditScraper(),
	}, nil
}

func (s *Scheduler) scrapeSubreddit(subreddit string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	log.Printf("[CRON] Starting Reddit scrape for r/%s", subreddit)

	posts, comments, err := s.redditScraper.ScrapeSubreddit(ctx, subreddit)
	if err != nil {
		log.Printf("[CRON] Error scraping r/%s: %v", subreddit, err)
		return
	}

	log.Printf("[CRON] Scraped r/%s: %d posts, %d comments", subreddit, len(posts), len(comments))

	// TODO: Process posts and comments - extract ticker mentions, save to DB, etc.
	// Example of what could be done here:
	// - Parse post titles and content for stock ticker symbols ($AAPL, $GME, etc.)
	// - Parse comment bodies for ticker mentions
	// - Save comments to database using s.store.CreateComment()
	// - Create ticker mentions using s.store.CreateTickerMention()
}

func (s *Scheduler) RegisterJobs() error {
	// Reddit scraping jobs - one for each subreddit
	// Stagger them to avoid hitting rate limits
	for i, subreddit := range subreddits {
		sub := subreddit // capture for closure
		offset := time.Duration(i*15)*time.Minute + time.Second

		_, err := s.scheduler.NewJob(
			gocron.DurationJob(1*time.Hour),
			gocron.NewTask(func() {
				s.scrapeSubreddit(sub)
			}),
			gocron.WithName("reddit-scrape-"+sub),
			gocron.WithStartAt(gocron.WithStartDateTime(time.Now().Add(offset))),
		)
		if err != nil {
			return err
		}

		log.Printf("[CRON] Registered Reddit scrape job for r/%s (starts in %v)", sub, offset)
	}

	return nil
}

func (s *Scheduler) Start() {
	log.Printf("[CRON] Starting scheduler with %d Reddit scraping jobs...", len(subreddits))
	s.scheduler.Start()
}

func (s *Scheduler) Stop() error {
	log.Println("[CRON] Stopping scheduler...")
	return s.scheduler.Shutdown()
}
