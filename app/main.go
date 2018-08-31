package main

import (
	"encoding/json"
	"fmt"
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
	http.HandleFunc("/widget", widget)
	http.HandleFunc("/", index)
	appengine.Main()
}

func index(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	log.Println(r.Header.Get("Accept-Encoding"))
	h := w.Header()
	h.Set("X-XSS-Protection", "1; mode=block")
	h.Set("X-Frame-Options", "DENY")
	h.Set("X-Content-Type-Options", "nosniff")
	h.Set("Content-Type", "text/html; charset=UTF-8")
	fmt.Fprint(w, `<!DOCTYPE html>
<html lang="ja">
<meta charset="UTF-8">
<title>どらっぴ</title>
<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/element-ui@2.4.6/lib/theme-chalk/index.css">
<style>
.description {
	font-size: 75%;
	color: #666;
}
</style>
<audio id="se" preload src="/static/se.opus"></audio>
<div id="app">
	<el-container>
		<el-header>
			<h1>どらっぴ</h1>
		</el-header>
		<el-main>
			<p>どらっぴは <a href="https://vrlive.party/" target="_blank">.LIVE</a> 所属メンバーのツイートをハッシュタグで絞って表示するツールです。</p>
			<h2>ウィジェットの作成</h2>
			<el-form ref="form" :model="form" label-width="120px">
				<el-form-item label="ハッシュタグ">
					<el-autocomplete class="inline-input" v-model="form.hashtag" :fetch-suggestions="suggestHashtags" placeholder="入力するか候補から選択" @select="updateUrl">
						<template slot="prepend">#</template>
					</el-autocomplete>
					<div class="description">
						直接入力することで、候補にないハッシュタグも使えます。
					</div>
				</el-form-item>
				<el-form-item label="最低表示時間">
					<el-slider v-model="form.duration" :min="3" :max="30" :format-tooltip="formatDuration" @change="updateUrl"></el-slider>
					<div class="description">
						ツイートが消えるまでの速さを指定します。実際の速さはツイート内容により可変します。
					</div>
				</el-form-item>
				<el-form-item label="配置">
					<el-radio-group v-model="form.placement" @change="updateUrl">
						<el-radio-button label="0">上寄せ</el-radio-button>
						<el-radio-button label="1">下寄せ</el-radio-button>
					</el-radio-group>
					<div class="description">
						画面上部に配置する場合は上寄せ、下部に配置する場合は下寄せがおすすめです。
					</div>
				</el-form-item>
				<el-form-item label="通知音の音量">
					<el-slider v-model="form.volume" :min="0" :max="100" :format-tooltip="formatPercent" @change="updateVolume"></el-slider>
					<div class="description">
						ツイートの表示時に鳴る効果音の音量を指定します。
					</div>
				</el-form-item>
				<el-form-item label="テストモード">
					<el-checkbox v-model="form.enableTest" border @change="updateUrl">テストモードを有効にする</el-checkbox>
					<div class="description">
						有効にすると一部の設定を無視して適当なツイートを表示します。OBS での配置テスト及び動作確認用です。
					</div>
				</el-form-item>
				<el-form-item label="生成されたURL">
					<a v-bind:href="form.url" target="_blank">{{ form.url }}</a>
					<div class="description">
						この URL をコピーして OBS の「ブラウザ」を使って表示することで、配信画面にツイートを埋め込むことができます。
					</div>
				</el-form-item>
			</el-form>
			<h2>カスタム CSS</h2>
			<p>OBS の「カスタム CSS」を使うことでデザインの変更ができます。</p>
			<p>以下は実際の設定例で、そのまま「カスタム CSS」にコピペして使うこともできます。</p>
			<p>連続して貼り付けることで「黒（半透明）」＋「太字」＋「文字小さめ」のように設定を組み合わせることもできます。</p>
			<template>
				<el-tabs v-model="customCssExamples">
					<el-tab-pane label="黒" name="black"><el-input type="textarea" autosize readonly v-model="css.black"></el-input></el-tab-pane>
					<el-tab-pane label="黒（半透明）" name="blacktrans"><el-input type="textarea" autosize readonly v-model="css.blacktrans"></el-input></el-tab-pane>
					<el-tab-pane label="白（半透明）" name="whitetrans"><el-input type="textarea" autosize readonly v-model="css.whitetrans"></el-input></el-tab-pane>
					<el-tab-pane label="太字" name="bold"><el-input type="textarea" autosize readonly v-model="css.bold"></el-input></el-tab-pane>
					<el-tab-pane label="文字小さめ" name="small"><el-input type="textarea" autosize readonly v-model="css.small"></el-input></el-tab-pane>
					<el-tab-pane label="文字大きめ" name="big"><el-input type="textarea" autosize readonly v-model="css.big"></el-input></el-tab-pane>
					<el-tab-pane label="左からスライド" name="slideleft"><el-input type="textarea" autosize readonly v-model="css.slideleft"></el-input></el-tab-pane>
					<el-tab-pane label="右からスライド" name="slideright"><el-input type="textarea" autosize readonly v-model="css.slideright"></el-input></el-tab-pane>
				</el-tabs>
			</template>
			<h2>動作の仕組み</h2>
			<ol>
				<li><a href="https://twitter.com/YozakuraTama/lists/list" target="_blank">たまちゃんのリスト</a>からツイートを取得</li>
				<li>RT、引用RT、リプライを含むツイートを除外</li>
				<li>表示期間の範囲に含まれるツイートを表示</li>
			</ol>
			<ul>
				<li>画像つきツイートも表示できます</li>
				<li>動画やアンケート付きのツイートは未テストです</li>
				<li>その他細かい動作は未テストです</li>
			</ul>
		</el-main>
		<el-header>
			<p>Author: <a href="https://twitter.com/gezzdo">@gezzdo</a> / <a href="https://github.com/gezzdo/dltwpicker">GitHub</a></p>
		</el-header>
	</el-container>
</div>
<script src="https://cdn.jsdelivr.net/npm/vue@2.5.17/dist/vue.min.js"></script>
<script src="https://cdn.jsdelivr.net/npm/element-ui@2.4.6/lib/index.js"></script>
<script src="/static/index.js"></script>
</html>
`)
}

func widget(w http.ResponseWriter, r *http.Request) {
	h := w.Header()
	h.Set("X-XSS-Protection", "1; mode=block")
	h.Set("X-Frame-Options", "DENY")
	h.Set("X-Content-Type-Options", "nosniff")
	h.Set("Content-Type", "text/html; charset=UTF-8")
	fmt.Fprint(w, `<!DOCTYPE html>
<html lang="ja">
<meta charset="UTF-8">
<title>どらっぴ</title>
<link rel="stylesheet" href="/static/widget.css">
<audio id="se" preload src="/static/se.opus"></audio>
<div id="app" style="display: none">
	<div class="test-warning" :class="{ 'align-top': !alignBottom, 'align-bottom': alignBottom }" v-if="testMode">どらっぴテストモード中</div>
	<transition name="fade" @after-leave="completeFadeOut">
		<div id="tweet" v-if="tweet" :class="{ 'align-top': !alignBottom, 'align-bottom': alignBottom }">
			<tweet-timer :duration="tweetDuration" @on-complete="timerComplete"></tweet-timer>
			<img class="avatar" :src="tweet.user.profile_image_url_https.replace('_normal', '_bigger')">
			<a class="at" :href="'https://twitter.com/' + tweet.user.screen_name + '/status/' + tweet.id_str">{{ formatTime(tweet.created_at) }}</a>
			<div class="user"><tweet-username :name="tweet.user.name"></tweet-username><div class="screen-name">@{{ tweet.user.screen_name }}</div></div>
			<tweet-message :tweet="tweet" @on-parsed="decorateMessage"></tweet-message>
			<span class="logo"></span>
		</div>
	</transition>
</div>
<script src="https://twemoji.maxcdn.com/2/twemoji.min.js?11.0"></script>
<script src="https://cdn.jsdelivr.net/npm/vue@2.5.17/dist/vue.min.js"></script>
<script src="/static/widget.js"></script>
</html>
`)
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
