"use strict";

function parseQuery(qs) {
    return qs.substr(1).split('&').reduce((o, v) => {
        const pos = v.indexOf('=');
        if (pos != -1) {
            o[decodeURIComponent(v.substring(0, pos))] = decodeURIComponent(v.substring(pos + 1));
        } else {
            o[decodeURIComponent(v)] = '';
        }
        return o;
    }, {});
}

const unescaper = document.createElement('span');

function decorateEmoji(s, createElement) {
    unescaper.innerHTML = s;
    twemoji.parse(unescaper);
    const r = [];
    for (let i = 0; i < unescaper.childNodes.length; ++i) {
        const e = unescaper.childNodes[i];
        if (e.nodeType == 3) {
            r.push(e.textContent);
        } else if (e.nodeType == 1 && e.tagName == 'IMG') {
            r.push(createElement('img', {
                class: 'emoji',
                attrs: {
                    src: e.src,
                    draggable: 'false',
                    alt: e.alt
                }
            }));
        }
    }
    return r;
}

function indexOfTweet(tweets, id_str) {
    for (let i = 0; i < tweets.length; ++i) {
        if (tweets[i].id_str == id_str) {
            return i;
        }
    }
    return -1;
}

function indexOfHashtag(t, hashtag) {
    if (!t.entities) {
        return -1;
    }
    if (!t.entities.hashtags) {
        return -1;
    }
    let h = t.entities.hashtags;
    for (let i = 0; i < h.length; ++i) {
        if (h[i].text.normalize('NFKC').toLowerCase() == hashtag) {
            return i;
        }
    }
    return -1;
}

Vue.component('tweet-timer', {
    template: '<div class="timer"><div class="step" :style="{ width: width + \'%\' }"></div></div>',
    data() {
        return {
            timer: null,
            width: 0,
            startAt: 0,
        }
    },
    methods: {
        update() {
            const d = Date.now();
            if (this.startAt + this.duration < d) {
                this.width = 100;
                this.$emit('on-complete');
                clearInterval(this.timer);
                return;
            }
            this.width = (Date.now() - this.startAt) / this.duration * 100;
        },
    },
    props: {
        duration: {
            type: Number,
            required: true
        },
    },
    created() {
        this.startAt = Date.now();
        this.timer = setInterval(() => this.update(), 128);
    },
});

Vue.component('tweet-username', {
    render(createElement) {
        return createElement('div', {
            class: 'name'
        }, decorateEmoji(this.name, createElement));
    },
    props: {
        name: {
            type: String,
            required: true
        }
    },
});

Vue.component('tweet-message', {
    render(createElement) {
        const t = this.tweet;
        const links = [],
            media = [];

        const e = t.entities;
        let chars = 0;
        for (let i = 0; i < e.hashtags.length; ++i) {
            const h = e.hashtags[i];
            chars += h.text.length;
            links.push([h.indices, createElement('a', {
                class: 'hashtag',
                attrs: {
                    href: 'https://twitter.com/search?q=' + encodeURIComponent('#' + h.text)
                }
            }, '#' + h.text)]);
        }
        for (let i = 0; i < e.urls.length; ++i) {
            const u = e.urls[i];
            links.push([u.indices, createElement('a', {
                class: 'extlink',
                attrs: {
                    href: u.url
                }
            }, u.display_url)]);
        }
        for (let i = 0; i < e.user_mentions.length; ++i) {
            const m = e.user_mentions[i];
            links.push([m.indices, createElement('a', {
                class: 'mention',
                attrs: {
                    href: 'https://twitter.com/' + encodeURIComponent(m.screen_name)
                }
            }, '@' + m.screen_name)]);
        }

        const ee = t.extended_entities;
        if (ee.media) {
            for (let i = 0; i < ee.media.length; ++i) {
                const m = ee.media[i];
                media.push(m);
                links.push([m.indices, null]);
            }
        }

        const text = t.full_text.match(/[\uD800-\uDBFF][\uDC00-\uDFFF]|[\s\S]/g) || [];
        for (let i = 0; i < text.length; ++i) {
            if (text[i] === '\n') {
                links.push([
                    [i, i + 1], createElement('br')
                ]);
            }
        }
        links.sort((a, b) => a[0][0] == b[0][0] ? 0 : a[0][0] > b[0][0] ? 1 : -1);

        let pos = 0;
        const r = [];
        for (let i = 0; i < links.length; ++i) {
            const l = links[i];
            const partText = text.slice(pos, l[0][0]);
            chars += partText.length;
            Array.prototype.push.apply(r, decorateEmoji(partText.join(''), createElement))
            if (l[1] != null) {
                r.push(l[1]);
            }
            pos = l[0][1];
        }
        if (pos < text.length) {
            const partText = text.slice(pos, text.length);
            chars += partText.length;
            Array.prototype.push.apply(r, decorateEmoji(partText.join(''), createElement));
        }

        if (media.length) {
            const mc = [];
            for (let i = 0; i < media.length; ++i) {
                const m = media[i];
                mc.push(createElement('img', {
                    class: 'thumb',
                    attrs: {
                        src: m.media_url_https + ':small'
                    }
                }));
            }
            r.push(createElement('div', {
                class: 'media'
            }, mc));
        }
        this.$emit('on-parsed', [chars, media.length]);
        return createElement('div', {
            class: 'message'
        }, r);
    },
    props: {
        tweet: {
            type: Object,
            required: true
        }
    },
});

const Main = {
    data() {
        return {
            hashtag: '',
            durationBase: 5,
            alignBottom: false,
            volume: 10,
            testMode: true,
            url: 'https://',
            obs: {
                active: true,
                visible: true,
            },

            tweet: null,
            latestTweetId: null,
            tweetDuration: 0,
            tweets: [],

            RECEIVE_INTERVAL: 60 * 1000,
        }
    },
    methods: {
        formatTime(t) {
            const d = Date.now() - t.getTime();
            if (60 * 1000 > d) {
                return '今';
            }
            if (60 * 60 * 1000 > d) {
                return (d / (60 * 1000) | 0) + '分';
            }
            if (24 * 60 * 60 * 1000 > d) {
                return (d / (60 * 60 * 1000) | 0) + '時間';
            }
            return (t.getMonth() + 1) + '月' + t.getDate() + '日';
        },

        initOptions() {
            const qs = parseQuery(location.search);

            this.hashtag = qs.h.normalize('NFKC').toLowerCase();

            let v = parseInt(qs.d, 10);
            if (!isNaN(v)) {
                this.durationBase = Math.max(3, Math.min(30, v));
            }

            this.alignBottom = qs.p == '1';
            this.testMode = qs.test == '1';

            v = parseInt(qs.vol, 10);
            if (!isNaN(v)) {
                this.volume = Math.max(0, Math.min(100, v));
            }
        },

        initOBSEvents() {
            obsstudio.onVisibilityChange = v => {
                this.obs.visible = v;
            };
            obsstudio.onActiveChange = v => {
                this.obs.active = v;
            };
        },

        filter(t) {
            if (t.in_reply_to_status_id_str || t.quoted_status_id_str || t.retweeted_status) {
                return false;
            }
            if (t.user.id_str == "953079145335988224") { // @dotLIVEyoutuber
                return false;
            }
            if (this.testMode || this.hashtag === '') {
                return true;
            }
            return indexOfHashtag(t, this.hashtag) != -1;
        },

        findNextTweet(id) {
            let prevIdx = -1;
            for (let i = 0; i < this.tweets.length; ++i) {
                const t = this.tweets[i];
                if (!this.filter(t)) {
                    continue;
                }
                if (t.id_str == id) {
                    if (prevIdx == -1) {
                        return null;
                    }
                    return this.tweets[prevIdx];
                }
                prevIdx = i;
            }
            return this.tweets[prevIdx];
        },

        findNext() {
            this.tweet = this.findNextTweet(this.latestTweetId);
            if (this.tweet == null) {
                return;
            }
            this.latestTweetId = this.tweet.id_str;

            if (this.obs.visible && this.obs.active) {
                const se = document.getElementById('se');
                se.pause();
                se.currentTime = 0;
                se.volume = this.volume / 100;
                se.play();
            }
        },

        decorateMessage(v) {
            const [chars, media] = v;
            this.tweetDuration = (chars / 140 * 15 + media / 4 * 10 + this.durationBase) * 1000;
        },

        timerComplete() {
            this.tweet = null;
            this.tweetDuration = 0;
        },

        completeFadeOut() {
            this.findNext();
        },

        receiveTweets() {
            let r = '/tweets?t=';
            if (this.tweets.length) {
                r += this.tweets[0].created_at.getTime() / 1000 | 0;
            } else {
                if (this.testMode) {
                    r += 0;
                } else {
                    r += (Date.now() - this.RECEIVE_INTERVAL) / 1000 | 0;
                }
            }
            if (!this.testMode && this.hashtag != '') {
                r += '&h=' + encodeURIComponent(this.hashtag);
            }
            return fetch(r).then(resp => resp.json()).then(newTweets => {
                for (let i = newTweets.length - 1; i >= 0; --i) {
                    if (indexOfTweet(this.tweets, newTweets[i].id_str) !== -1) {
                        continue;
                    }
                    const t = newTweets[i];
                    t.created_at = new Date(t.created_at);
                    this.tweets.unshift(t);
                }
                this.tweets.sort((a, b) => {
                    const at = a.created_at.getTime(),
                        bt = b.created_at.getTime();
                    return at == bt ? 0 : at < bt ? 1 : -1;
                });
                if (!this.testMode) {
                    // delete overly old tweets
                    const n = Date.now() - 5 * 60 * 1000;
                    this.tweets = this.tweets.filter(t => t.created_at.getTime() > n);
                }
            });
        },

        receiveLoop() {
            this.receiveTweets().then(() => {
                if (!this.tweet) {
                    this.findNext();
                }
            });
            setTimeout(() => this.receiveLoop(), this.RECEIVE_INTERVAL);
        },
    },
    created() {
        this.initOptions();
        if (window.obsstudio) {
            this.initOBSEvents();
        }
    },
    mounted() {
        setTimeout(() => this.receiveLoop(), this.testMode ? 0 : this.RECEIVE_INTERVAL);
    },
};
const Ctor = Vue.extend(Main);
document.getElementById('app').removeAttribute('style');
var m = new Ctor().$mount('#app');