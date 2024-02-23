package main

import (
	"math/rand"
)

func Usergen() string {
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
		"kun", "chan", "san", "sama", "dono", "sensei", "shisho", "hakase", "shi",
	}
	displaynamemiddle := []string{
		"くん", "ちゃん", "さん", "さま", "どの", "せんせい", "ししょう", "はかせ", "し",
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

	return name + "," + displayname + "," + password + "," + description // TODO: User 構造体に入れる
}
