package cron

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

const (
	redditBaseURL = "https://www.reddit.com"
	userAgent     = "Mozilla/5.0 (compatible; StockMentionBot/1.0)"
)

type RedditScraper struct {
	client *http.Client
}

type RedditPost struct {
	ID          string
	Title       string
	Author      string
	Selftext    string
	CreatedAt   time.Time
	URL         string
	Subreddit   string
	NumComments int
}

type RedditComment struct {
	ID        string
	Author    string
	Body      string
	CreatedAt time.Time
	PostID    string
	ParentID  string
	Replies   []RedditComment
}

type RedditListingChildren []struct {
	Data struct {
		ID          string  `json:"id"`
		Title       string  `json:"title"`
		Author      string  `json:"author"`
		Selftext    string  `json:"selftext"`
		CreatedUTC  float64 `json:"created_utc"`
		Permalink   string  `json:"permalink"`
		Subreddit   string  `json:"subreddit"`
		NumComments int     `json:"num_comments"`
	} `json:"data"`
}

type RedditListingResponse struct {
	Data struct {
		Children RedditListingChildren `json:"children"`
		After    string                `json:"after"`
	} `json:"data"`
}

type CommentsChildren []struct {
	Kind string `json:"kind"`
	Data struct {
		ID         string      `json:"id"`
		Author     string      `json:"author"`
		Body       string      `json:"body"`
		CreatedUTC float64     `json:"created_utc"`
		ParentID   string      `json:"parent_id"`
		Replies    interface{} `json:"replies"`
	} `json:"data"`
}

type RedditCommentResponse []struct {
	Data struct {
		Children CommentsChildren `json:"children"`
	} `json:"data"`
}

func NewRedditScraper() *RedditScraper {
	return &RedditScraper{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (r *RedditScraper) makeRequest(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", userAgent)

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 429 {
		return nil, fmt.Errorf("rate limited by Reddit, status: %d", resp.StatusCode)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return body, nil
}

// FetchSubredditPosts fetches all posts from a subreddit within the last 24 hours
func (r *RedditScraper) FetchSubredditPosts(ctx context.Context, subreddit string) ([]RedditPost, error) {
	var allPosts []RedditPost
	cutoffTime := time.Now().Add(-24 * time.Hour)
	after := ""

	for {
		url := fmt.Sprintf("%s/r/%s/new.json?limit=2000", redditBaseURL, subreddit)
		if after != "" {
			url += "&after=" + after
		}

		body, err := r.makeRequest(ctx, url)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch posts from r/%s: %w", subreddit, err)
		}

		var response RedditListingResponse
		if err := json.Unmarshal(body, &response); err != nil {
			return nil, fmt.Errorf("failed to parse posts response: %w", err)
		}

		reachedOldPosts := false
		for _, child := range response.Data.Children {
			postTime := time.Unix(int64(child.Data.CreatedUTC), 0)

			if postTime.Before(cutoffTime) {
				reachedOldPosts = true
				break
			}

			post := RedditPost{
				ID:          child.Data.ID,
				Title:       child.Data.Title,
				Author:      child.Data.Author,
				Selftext:    child.Data.Selftext,
				CreatedAt:   postTime,
				URL:         redditBaseURL + child.Data.Permalink,
				Subreddit:   child.Data.Subreddit,
				NumComments: child.Data.NumComments,
			}
			allPosts = append(allPosts, post)
		}

		if reachedOldPosts || response.Data.After == "" {
			break
		}

		after = response.Data.After

		// Rate limiting: wait between paginated requests
		time.Sleep(2 * time.Second)
	}

	return allPosts, nil
}

// FetchPostComments fetches all comments for a specific post, including nested replies
func (r *RedditScraper) FetchPostComments(ctx context.Context, subreddit, postID string) ([]RedditComment, error) {
	url := fmt.Sprintf("%s/r/%s/comments/%s.json?limit=2000&depth=10", redditBaseURL, subreddit, postID)

	body, err := r.makeRequest(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch comments for post %s: %w", postID, err)
	}

	var response RedditCommentResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse comments response: %w", err)
	}

	var comments []RedditComment
	if len(response) >= 2 {
		comments = r.parseComments(response[1].Data.Children, postID)
	}

	return comments, nil
}

func (r *RedditScraper) parseComments(children CommentsChildren, postID string) []RedditComment {
	var comments []RedditComment

	for _, child := range children {
		if child.Kind != "t1" {
			continue
		}

		comment := RedditComment{
			ID:        child.Data.ID,
			Author:    child.Data.Author,
			Body:      child.Data.Body,
			CreatedAt: time.Unix(int64(child.Data.CreatedUTC), 0),
			PostID:    postID,
			ParentID:  child.Data.ParentID,
		}

		// Parse nested replies
		if child.Data.Replies != nil {
			if repliesMap, ok := child.Data.Replies.(map[string]interface{}); ok {
				if data, ok := repliesMap["data"].(map[string]interface{}); ok {
					if childrenRaw, ok := data["children"].([]interface{}); ok {
						comment.Replies = r.parseNestedReplies(childrenRaw, postID)
					}
				}
			}
		}

		comments = append(comments, comment)
	}

	return comments
}

func (r *RedditScraper) parseNestedReplies(childrenRaw []interface{}, postID string) []RedditComment {
	var replies []RedditComment

	for _, childRaw := range childrenRaw {
		childMap, ok := childRaw.(map[string]interface{})
		if !ok {
			continue
		}

		kind, _ := childMap["kind"].(string)
		if kind != "t1" {
			continue
		}

		dataMap, ok := childMap["data"].(map[string]interface{})
		if !ok {
			continue
		}

		id, _ := dataMap["id"].(string)
		author, _ := dataMap["author"].(string)
		body, _ := dataMap["body"].(string)
		createdUTC, _ := dataMap["created_utc"].(float64)
		parentID, _ := dataMap["parent_id"].(string)

		comment := RedditComment{
			ID:        id,
			Author:    author,
			Body:      body,
			CreatedAt: time.Unix(int64(createdUTC), 0),
			PostID:    postID,
			ParentID:  parentID,
		}

		// Recursively parse nested replies
		if repliesRaw, ok := dataMap["replies"].(map[string]interface{}); ok {
			if repliesData, ok := repliesRaw["data"].(map[string]interface{}); ok {
				if nestedChildren, ok := repliesData["children"].([]interface{}); ok {
					comment.Replies = r.parseNestedReplies(nestedChildren, postID)
				}
			}
		}

		replies = append(replies, comment)
	}

	return replies
}

// FlattenComments converts nested comments into a flat slice for easier processing
func (r *RedditScraper) FlattenComments(comments []RedditComment) []RedditComment {
	var flat []RedditComment

	var flatten func(cs []RedditComment)
	flatten = func(cs []RedditComment) {
		for _, c := range cs {
			flat = append(flat, RedditComment{
				ID:        c.ID,
				Author:    c.Author,
				Body:      c.Body,
				CreatedAt: c.CreatedAt,
				PostID:    c.PostID,
				ParentID:  c.ParentID,
			})
			if len(c.Replies) > 0 {
				flatten(c.Replies)
			}
		}
	}

	flatten(comments)
	return flat
}

// ScrapeSubreddit is the main entry point - fetches all posts and their comments from the last 24 hours
func (r *RedditScraper) ScrapeSubreddit(ctx context.Context, subreddit string) ([]RedditPost, []RedditComment, error) {
	log.Printf("[REDDIT] Starting scrape for r/%s", subreddit)

	posts, err := r.FetchSubredditPosts(ctx, subreddit)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch posts: %w", err)
	}

	log.Printf("[REDDIT] Found %d posts in r/%s from the last 24 hours", len(posts), subreddit)

	var allComments []RedditComment

	for i, post := range posts {
		if post.NumComments == 0 {
			continue
		}

		comments, err := r.FetchPostComments(ctx, subreddit, post.ID)
		if err != nil {
			log.Printf("[REDDIT] Warning: failed to fetch comments for post %s: %v", post.ID, err)
			continue
		}

		flatComments := r.FlattenComments(comments)
		allComments = append(allComments, flatComments...)

		log.Printf("[REDDIT] Post %d/%d (%s): fetched %d comments", i+1, len(posts), post.ID, len(flatComments))

		// Rate limiting between posts
		time.Sleep(2 * time.Second)
	}

	log.Printf("[REDDIT] Completed r/%s: %d posts, %d total comments", subreddit, len(posts), len(allComments))

	return posts, allComments, nil
}
