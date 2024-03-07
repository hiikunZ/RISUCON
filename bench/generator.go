package main

import (
	"math/rand"
	"sync"
)

var (
	usernamesetmu, teamnamesetmu = sync.Mutex{}, sync.Mutex{}
	usernameset, teamnameset     = make(map[string]bool, 0), make(map[string]bool, 0)
)

func (s *Scenario) addexistnametoset() {
	for _, u := range s.Users.list {
		usernameset[u.Name] = true
	}
	for _, t := range s.Teams.list {
		teamnameset[t.Name] = true
	}
}

func Usergen() User {
regenerate:
	nameprefix := []string{
		"a", "i", "u", "e", "o", "ka", "ki", "ku", "ke", "ko", "sa", "shi", "su", "se", "so", "ta", "chi", "tsu", "te", "to", "na", "ni", "nu", "ne", "no", "ha", "hi", "fu", "he", "ho", "ma", "mi", "mu", "me", "mo", "ya", "yu", "yo", "ra", "ri", "ru", "re", "ro", "wa", "wo", "ga", "gi", "gu", "ge", "go", "za", "zi", "zu", "ze", "zo", "da", "di", "du", "de", "do", "ba", "bi", "bu", "be", "bo", "pa", "pi", "pu", "pe", "po",
		"aa", "ii", "uu", "ee", "oo", "kaa", "kii", "kuu", "kee", "koo", "saa", "shii", "suu", "see", "soo", "taa", "chii", "tsuu", "tee", "too", "naa", "nii", "nuu", "nee", "noo", "haa", "hii", "fuu", "hee", "hoo", "maa", "mii", "muu", "mee", "moo", "yaa", "yuu", "yoo", "raa", "rii", "ruu", "ree", "roo", "waa", "woo", "gaa", "gii", "guu", "gee", "goo", "zaa", "zii", "zuu", "zee", "zoo", "daa", "dii", "duu", "dee", "doo", "baa", "bii", "buu", "bee", "boo", "paa", "pii", "puu", "pee", "poo",
		"kyaa", "kyuu", "kyoo", "gyaa", "gyuu", "gyoo", "shaa", "shuu", "shoo", "jaa", "juu", "joo", "chaa", "chuu", "choo", "nyaa", "nyuu", "nyoo", "hyaa", "hyuu", "hyoo", "byaa", "byuu", "byoo", "pyaa", "pyuu", "pyoo", "myaa", "myuu", "myoo", "ryaa", "ryuu", "ryoo",
	}
	displaynameprefix := []string{
		"あ", "い", "う", "え", "お", "か", "き", "く", "け", "こ", "さ", "し", "す", "せ", "そ", "た", "ち", "つ", "て", "と", "な", "に", "ぬ", "ね", "の", "は", "ひ", "ふ", "へ", "ほ", "ま", "み", "む", "め", "も", "や", "ゆ", "よ", "ら", "り", "る", "れ", "ろ", "わ", "を", "が", "ぎ", "ぐ", "げ", "ご", "ざ", "じ", "ず", "ぜ", "ぞ", "だ", "ぢ", "づ", "で", "ど", "ば", "び", "ぶ", "べ", "ぼ", "ぱ", "ぴ", "ぷ", "ぺ", "ぽ",
		"あー", "いー", "うー", "えー", "おー", "かー", "きー", "くー", "けー", "こー", "さー", "しー", "すー", "せー", "そー", "たー", "ちー", "つー", "てー", "とー", "なー", "にー", "ぬー", "ねー", "のー", "はー", "ひー", "ふー", "へー", "ほー", "まー", "みー", "むー", "めー", "もー", "やー", "ゆー", "よー", "らー", "りー", "るー", "れー", "ろー", "わー", "をー", "がー", "ぎー", "ぐー", "げー", "ごー", "ざー", "じー", "ずー", "ぜー", "ぞー", "だー", "ぢー", "づー", "でー", "どー", "ばー", "びー", "ぶー", "べー", "ぼー", "ぱー", "ぴー", "ぷー", "ぺー", "ぽー",
		"きゃー", "きゅー", "きょー", "ぎゃー", "ぎゅー", "ぎょー", "しゃー", "しゅー", "しょー", "じゃー", "じゅー", "じょー", "ちゃー", "ちゅー", "ちょー", "にゃー", "にゅー", "にょー", "ひゃー", "ひゅー", "ひょー", "びゃー", "びゅー", "びょー", "ぴゃー", "ぴゅー", "ぴょー", "みゃー", "みゅー", "みょー", "りゃー", "りゅー", "りょー",
	}
	namemiddle := []string{
		"kun", "chan", "san", "sama", "dono", "sensei", "shisho", "hakase", "shi", "cat", "dog", "rabbit", "red", "blue", "green", "yellow", "black", "white", "orange", "pink", "purple", "brown", "gray", "silver", "gold", "platinum", "diamond", "ruby", "sapphire", "emerald", "crystal", "pearl", "opal", "amethyst", "topaz", "garnet", "aquamarine", "peridot", "citrine", "turquoise",
	}
	displaynamemiddle := []string{
		"くん", "ちゃん", "さん", "さま", "どの", "せんせい", "ししょう", "はかせ", "し", "きゃっと", "どっぐ", "らびっと", "れっど", "ぶるー", "ぐりーん", "いえろー", "ぶらっく", "ほわいと", "おれんじ", "ぴんく", "ぱーぷる", "ぶらうん", "ぐれい", "しるばー", "ごーるど", "ぷらちな", "だいあもんど", "るびー", "さふぁいあ", "えめらるど", "くりすたる", "ぱーる", "おぱーる", "あめしすと", "とぱーず", "がーねっと", "あくあまりん", "ぺりどっと", "しとりん", "たーこいず",
	}
	namesuffix := []string{
		"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z", "1", "2", "3", "4", "5", "6", "7", "8", "9", "0",
	}
	displaynamesuffix := []string{
		"えー", "びー", "しー", "でぃー", "いー", "えふ", "じー", "えいち", "あい", "じぇー", "けー", "える", "えむ", "えぬ", "おー", "ぴー", "きゅー", "あーる", "えす", "てぃー", "ゆー", "ぶい", "だぶりゅー", "えっくす", "わい", "ぜっと", "わん", "つー", "すりー", "ふぉー", "ふぁいぶ", "しっくす", "せぶん", "えいと", "ないん", "ぜろ",
	}
	prefixidx := rand.Intn(len(nameprefix))
	middleidx := rand.Intn(len(namemiddle))
	suffixidx := rand.Intn(len(namesuffix))

	name := nameprefix[prefixidx] + namemiddle[middleidx] + namesuffix[suffixidx]

	usernamesetmu.Lock()
	if usernameset[name] {
		goto regenerate
	}
	usernameset[name] = true
	usernamesetmu.Unlock()

	displayname := displaynameprefix[prefixidx] + displaynamemiddle[middleidx] + displaynamesuffix[suffixidx]

	passwordletters := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	password := ""
	for i := 0; i < 8; i++ {
		password += string(passwordletters[rand.Intn(len(passwordletters))])
	}

	gakunen := []string{
		"中学1年", "中学2年", "中学3年", "高校1年", "高校2年", "高校3年",
	}

	gakunenidx := rand.Intn(len(gakunen))

	description := gakunen[gakunenidx] + "の" + displayname + "です。よろしくお願いします。"

	return User{
		ID:            -1,
		Name:          name,
		DisplayName:   displayname,
		Description:   description,
		Password:      password,
		SubmissionIDs: []int{},
	}
}

func Teamgen() Team {
regenerate:
	nameprefix := []string{
		"cho", "sugoi", "kamino", "saikyo", "kyoha", "super", "miracle", "hyper", "special", "ultra", "mysterious", "fantastic", "mystic", "mystical", "tensai", "tenshino", "akumano", "mahouno", "maou", "majo", "yuusha", "kishi", "ninja", "samurai", "shinobi", "pasokon",
	}
	displaynameprefix := []string{
		"超", "すごい", "神の", "最強", "今日は", "スーパー", "ミラクル", "ハイパー", "スペシャル", "ウルトラ", "ミステリアス", "ファンタスティック", "ミスティック", "ミスティカル", "天才", "天使の", "悪魔の", "魔法の", "魔王", "魔女", "勇者", "騎士", "忍者", "侍", "忍び", "パソコン",
	}
	namemiddle := []string{
		"risu", "inu", "inko", "usagi", "oumu", "kaba", "kamonohashi", "kamome", "karasu", "sai", "suzume", "tsubame", "neko", "nezumi", "hagewashi", "hato", "hamusuta", "fukurou", "furamingo", "perikan", "pengin", "mimizuku", "mogura", "momonga", "morumotto", "washi", "ryokuwotakameru",
	}
	displaynamemiddle := []string{
		"リス", "イヌ", "インコ", "ウサギ", "オウム", "カバ", "カモノハシ", "カモメ", "カラス", "サイ", "スズメ", "ツバメ", "ネコ", "ネズミ", "ハゲワシ", "ハト", "ハムスター", "フクロウ", "フラミンゴ", "ペリカン", "ペンギン", "ミミズク", "モグラ", "モモンガ", "モルモット", "ワシ", "力を高める",
	}
	namesuffix := []string{
		"team", "group", "gumi", "dan", "nokai", "club", "circle", "guild", "bu", "sha", "kumiai", "doumei", "renmei", "kyoukai", "rengou", "nonakamatachi", "nominasan", "toyukainanakamatachi", "etal", "nado", "desu", "gundan", "gun", "tai",
	}
	displaynamesuffix := []string{
		"チーム", "グループ", "組", "団", "の会", "クラブ", "サークル", "ギルド", "部", "社", "組合", "同盟", "連盟", "協会", "連合", "の仲間たち", "の皆さん", "と愉快な仲間たち", "et al.", "など", "です", "軍団", "軍", "隊",
	}
	prefixidx := rand.Intn(len(nameprefix))
	middleidx := rand.Intn(len(namemiddle))
	suffixidx := rand.Intn(len(namesuffix))

	name := nameprefix[prefixidx] + namemiddle[middleidx] + namesuffix[suffixidx]

	teamnamesetmu.Lock()
	if teamnameset[name] {
		goto regenerate
	}
	teamnameset[name] = true
	teamnamesetmu.Unlock()
	
	displayname := displaynameprefix[prefixidx] + displaynamemiddle[middleidx] + displaynamesuffix[suffixidx]
	description := "チーム「" + displayname + "」です。よろしくお願いします。"
	invitation_code := ""
	for i := 0; i < 16; i++ {
		invitation_code += string("0123456789abcdef"[rand.Intn(16)])
	}

	return Team{
		ID:             -1,
		Name:           name,
		DisplayName:    displayname,
		LeaderID:       nulluserid,
		Member1ID:      nulluserid,
		Member2ID:      nulluserid,
		Description:    description,
		InvitationCode: invitation_code,
	}
}
