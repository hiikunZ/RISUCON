package bench

import (
	"math/rand"

)

func Usergen() string {
	nameprefix := []string{
		"aa","ii","uu","ee","oo","kaa","kii","kuu","kee","koo","saa","sii","suu","see","soo","taa","tii","tuu","tee","too","naa","nii","nuu","nee","noo","haa","hii","huu","hee","hoo","maa","mii","muu","mee","moo","yaa","yuu","yoo","raa","rii","ruu","ree","roo","waa","woo","gaa","gii","guu","gee","goo","zaa","zii","zuu","zee","zoo","daa","dii","duu","dee","doo","baa","bii","buu","bee","boo","paa","pii","puu","pee","poo",
	}
	displaynameprefix := []string{
		"あー","いー","うー","えー","おー","かー","きー","くー","けー","こー","さー","しー","すー","せー","そー","たー","ちー","つー","てー","とー","なー","にー","ぬー","ねー","のー","はー","ひー","ふー","へー","ほー","まー","みー","むー","めー","もー","やー","ゆー","よー","らー","りー","るー","れー","ろー","わー","をー","がー","ぎー","ぐー","げー","ごー","ざー","じー","ずー","ぜー","ぞー","だー","ぢー","づー","でー","どー","ばー","びー","ぶー","べー","ぼー","ぱー","ぴー","ぷー","ぺー","ぽー",
	}
	namemiddle := []string{
		"kun","chan","san",
	}
	displaynamemiddle := []string{
		"くん","ちゃん","さん",
	}
	namesuffix := []string{
		"A","B","C","D","E","F","G","H","I","J","K","L","M","N","O","P","Q","R","S","T","U","V","W","X","Y","Z","1","2","3","4","5","6","7","8","9","0",
	}
	displaynamesuffix := []string{
		"えー","びー","しー","でぃー","いー","えふ","じー","えいち","あい","じぇー","けー","える","えむ","えぬ","おー","ぴー","きゅー","あーる","えす","てぃー","ゆー","ぶい","だぶりゅー","えっくす","わい","ぜっと","わん","つー","すりー","ふぉー","ふぁいぶ","しっくす","せぶん","えいと","ないん","ぜろ",
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
		"中学1年","中学2年","中学3年","高校1年","高校2年","高校3年",
	}

	gakunenidx := rand.Intn(len(gakunen))

	description := gakunen[gakunenidx] + "の" + displayname + "です。よろしくお願いします。"

	return name + "," + displayname + "," + password + "," + description // TODO: User 構造体に入れる
}