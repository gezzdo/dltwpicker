"use strict";

const Main = {
	data() {
		return {
			form: {
				hashtag: '',
				duration: 5,
				placement: 0,
				volume: 10,
				enableTest: true,
				url: 'https://'
			},
			customCssExamples: 'black',
			css: {
				black: `#tweet {
    background-color: #000;
    color: #fff;
}
#tweet .user .screen-name,
#tweet .at {
    color: #aaa;
}
#tweet .message a {
    color: #8bf;
}
`,
				blacktrans: `#tweet {
    background-color: rgba(0, 0, 0, 0.4);
    color: #fff;
    text-shadow:
        rgba(0,0,0,0.5) -2px -2px, #000 -1px -2px, #000 0px -2px, #000 1px -2px, rgba(0,0,0,0.5) 2px -2px,
        #000 -2px -1px, #000 -1px -1px, #000 0px -1px, #000 1px -1px, #000 2px -1px,
        #000 -2px  0px, #000 -1px  0px, #000 0px  0px, #000 1px  0px, #000 2px  0px,
        #000 -2px  1px, #000 -1px  1px, #000 0px  1px, #000 1px  1px, #000 2px  1px,
        rgba(0,0,0,0.5) -2px  2px, #000 -1px  2px, #000 0px  2px, #000 1px  2px, rgba(0,0,0,0.5) 2px  2px;
}
#tweet .user .screen-name,
#tweet .at {
    color: #aaa;
}
#tweet .message a {
    color: #8bf;
}
`,
				whitetrans: `#tweet {
    background-color: rgba(255, 255, 255, 0.66);
    color: #333;
    text-shadow:
        rgba(255,255,255,0.5) -2px -2px, #fff -1px -2px, #fff 0px -2px, #fff 1px -2px, rgba(255,255,255,0.5) 2px -2px,
        #fff -2px -1px, #fff -1px -1px, #fff 0px -1px, #fff 1px -1px, #fff 2px -1px,
        #fff -2px  0px, #fff -1px  0px, #fff 0px  0px, #fff 1px  0px, #fff 2px  0px,
        #fff -2px  1px, #fff -1px  1px, #fff 0px  1px, #fff 1px  1px, #fff 2px  1px,
        rgba(255,255,255,0.5) -2px  2px, #fff -1px  2px, #fff 0px  2px, #fff 1px  2px, rgba(255,255,255,0.5) 2px  2px;
}
#tweet .user .screen-name,
#tweet .at {
    color: #888;
}
#tweet .message a {
    color: #08b;
}
`,
				bold: `@import url("https://fonts.googleapis.com/css?family=M+PLUS+1p:700");
#tweet {
    font-family: 'M PLUS 1p';
    font-weight: 700;
}
#tweet .user .name,
#tweet .user .screen-name,
#tweet .at,
#tweet .message {
    transform: rotate(0.04deg); /* 字が汚く見えるのを回避 */
}
`,
				small: `#tweet {
    font-size: 80%;
}
`,
				big: `#tweet {
    font-size: 120%;
}
`,
				slideleft: `.fade-enter {
    transform: translateX(-32px);
}
`,
				slideright: `.fade-enter {
    transform: translateX(32px);
}
`,
			},
		}
	},
	methods: {
		suggestHashtags(queryString, cb) {
			cb([{
					value: "牛巻りこ"
				},
				{
					value: "花京院ちえり"
				},
				{
					value: "神楽すず"
				},
				{
					value: "カルロピノ"
				},
				{
					value: "木曽あずき"
				},
				{
					value: "北上双葉"
				},
				{
					value: "金剛いろは"
				},
				{
					value: "猫乃木もち"
				},
				{
					value: "もこ田めめめ"
				},
				{
					value: "八重沢なとり"
				},
				{
					value: "ヤマトイオリ"
				},
				{
					value: "夜桜たま"
				},

				{
					value: "シロ生放送"
				},
				{
					value: "ばあちゃる"
				},
			]);
		},
		updateUrl() {
			this.form.url = location.protocol + '//' + location.host + '/widget' +
				'?h=' + encodeURIComponent(this.form.hashtag) +
				'&d=' + this.form.duration +
				'&p=' + this.form.placement +
				'&vol=' + this.form.volume +
				'&test=' + (this.form.enableTest ? 1 : 0);
		},
		updateVolume() {
			const se = document.getElementById('se');
			se.pause();
			se.currentTime = 0;
			se.volume = this.form.volume / 100;
			se.play();
			this.updateUrl();
		},
		formatDuration(val) {
			return val + ' 秒以上';
		},
		formatPercent(val) {
			return val + '%';
		},
	},
	mounted() {
		let lastHash = '';
		setInterval(() => {
			if (this.form.hashtag != lastHash) {
				lastHash = this.form.hashtag;
				this.updateUrl();
			}
		}, 400);
		this.updateUrl();
	},
};
const Ctor = Vue.extend(Main);
new Ctor().$mount('#app');