package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/ChimeraCoder/anaconda"
	"github.com/pkg/errors"
	"google.golang.org/appengine/memcache"
	"google.golang.org/appengine/urlfetch"
)

func twAPI(ctx context.Context) *anaconda.TwitterApi {
	api := anaconda.NewTwitterApiWithCredentials(
		os.Getenv("TWITTER_ACCESS_TOKEN"),
		os.Getenv("TWITTER_ACCESS_TOKEN_SECRET"),
		os.Getenv("TWITTER_CONSUMER_KEY"),
		os.Getenv("TWITTER_CONSUMER_SECRET"))
	api.HttpClient.Transport = &urlfetch.Transport{Context: ctx}
	return api
}

func getTweetsFromMemcache(ctx context.Context, cacheID string) ([]anaconda.Tweet, error) {
	item, err := memcache.Get(ctx, cacheID)
	if err != nil {
		return nil, err
	}
	var tweets []anaconda.Tweet
	err = json.Unmarshal(item.Value, &tweets)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal cache data")
	}
	return tweets, nil
}

func getTweetsFromTwitter(ctx context.Context, listID int64) ([]anaconda.Tweet, error) {
	api := twAPI(ctx)
	tweets, err := api.GetListTweets(listID, false, url.Values{"count": []string{"100"}})
	if err != nil {
		return nil, errors.Wrap(err, "could not get tweets from twitter")
	}
	return tweets, nil
}

func updateMemcacheTweets(ctx context.Context, listID int64, cacheID string, expiration time.Duration) ([]anaconda.Tweet, error) {
	tweets, err := getTweetsFromMemcache(ctx, cacheID)
	if err == nil {
		// already updated
		return tweets, nil
	}
	if err != memcache.ErrCacheMiss {
		return nil, err
	}

	tweets, err = getTweetsFromTwitter(ctx, listID)
	if err != nil {
		return nil, err
	}

	b, err := json.Marshal(tweets)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal tweets")
	}

	err = memcache.Set(ctx, &memcache.Item{
		Key:        cacheID,
		Value:      b,
		Expiration: expiration,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to write tweet caches to memcache")
	}
	return tweets, nil
}

func getTweets(ctx context.Context, listID int64, expiration time.Duration) ([]anaconda.Tweet, error) {
	cacheID := fmt.Sprintf("list%d", listID)
	tweets, err := getTweetsFromMemcache(ctx, cacheID)
	if err == nil {
		return tweets, nil
	}
	if err != nil && err != memcache.ErrCacheMiss {
		return nil, errors.Wrap(err, "failed to get tweets from cache")
	}
	var err2 error
	err = transaction(ctx, cacheID+"-mutex", func() {
		tweets, err2 = updateMemcacheTweets(ctx, listID, cacheID, expiration)
	})
	if err != nil {
		return nil, err
	}
	if err2 != nil {
		return nil, errors.Wrap(err2, "failed to update tweet cache")
	}
	return tweets, nil
}
