package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ChimeraCoder/anaconda"
	"golang.org/x/text/unicode/norm"
	"google.golang.org/appengine"
)

func main() {
	http.HandleFunc("/tweets", tweets)
	appengine.Main()
}

func hasHashtag(tw anaconda.Tweet, h string) bool {
	for _, v := range tw.Entities.Hashtags {
		if strings.ToLower(norm.NFKC.String(v.Text)) == h {
			return true
		}
	}
	return false
}

func tweets(w http.ResponseWriter, r *http.Request) {
	listID, err := strconv.ParseInt(os.Getenv("TWITTER_TARGET_LIST_ID"), 10, 64)
	if err != nil {
		http.Error(w, "failed to read list ID", 500)
		log.Println(err)
		return
	}

	ctx := appengine.NewContext(r)
	tweets, err := getTweets(ctx, listID, 1*time.Minute)
	if err != nil {
		http.Error(w, "failed to get tweets", 500)
		log.Println(err)
		return
	}

	t, err := strconv.ParseInt(r.FormValue("t"), 10, 64)
	if err != nil {
		t = 0
	}

	hashtag := strings.ToLower(norm.NFKC.String(r.FormValue("h")))

	ret := make([]anaconda.Tweet, 0, len(tweets))
	for i := 0; i < len(tweets); i++ {
		tw := tweets[i]
		if tw.QuotedStatusIdStr != "" || tw.InReplyToStatusIdStr != "" || tw.RetweetedStatus != nil {
			continue
		}
		if hashtag != "" && !hasHashtag(tw, hashtag) {
			continue
		}
		if t != 0 {
			twAt, err := tw.CreatedAtTime()
			if err != nil {
				continue
			}
			if t > twAt.Unix() {
				continue
			}
		}
		ret = append(ret, tw)
		if t == 0 && len(ret) == 20 {
			break
		}
	}

	h := w.Header()
	h.Set("X-XSS-Protection", "1; mode=block")
	h.Set("X-Frame-Options", "DENY")
	h.Set("X-Content-Type-Options", "nosniff")
	h.Set("Cache-Control", "no-cache, no-store")
	h.Set("Pragma", "no-cache")
	h.Set("Content-Type", "application/json; charset=UTF-8")
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		http.Error(w, "failed to marshal tweet data", 500)
		log.Println(err)
		return
	}
}
