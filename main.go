package main

import (
	"bytes"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/robfig/cron/v3"
)

// before using the bot, please use this url paste to browser and accept it to channel or guild
// https://discord.com/api/oauth2/authorize?client_id=[put client id inside]&scope=bot&permissions=8

// Variables used for command line parameters
var (
	BotSecret    string
	GuildID      string
	XLSXFile     string
	BackupPeroid string
)

func init() {

	// -t token -g gulidid -f test.xlsx
	flag.StringVar(&BotSecret, "t", "", "Bot Token")
	flag.StringVar(&GuildID, "g", "", "855718269092233247")
	flag.StringVar(&XLSXFile, "f", "", "test.xlsx")
	flag.StringVar(&BackupPeroid, "period", "", "3600")
	flag.Parse()
}

func main() {

	ReadXLSXToTicketMap(XLSXFile, &KeyPairMap)

	// BackupPeroid start require variable larger then 0.
	if BackupPeroid != "0" {

		c := cron.New()
		c.AddFunc("@every "+BackupPeroid+"s", func() {
			SaveUsedCSV()
			log.Println("File Backup Done.")
		})
		go c.Start()
		// close cron
		defer c.Stop()

	}

	RegisterBotFuncAndRun(DiscordAuth{
		BotSecret: BotSecret,
	}, messageCreate)
}

// check the used KKTIX Token.
type Record struct {
	User string
	Time string
}

var usedToken map[string]*Record = make(map[string]*Record)

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	// check message and messages guild role ids
	/*
		fmt.Println("Received msg: ", m.Content)
		fmt.Println("GuildID:", m.GuildID)
		st, _ := s.GuildRoles(m.GuildID)
		for _, v := range st {
			fmt.Println("RoleName and ID: ", v.Name, v.ID)
		}
	*/

	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}

	// https://github-wiki-see.page/m/bwmarrin/discordgo/wiki/FAQ
	isDM, err := ComesFromDM(s, m)
	if err != nil {
		panic(err)
	}

	// comefrom private message
	if isDM {
		fmt.Println(m.Author.Username, m.Author.ID, "進行了驗票流程, 內容:", m.Content)

		// split to command
		commands := strings.Split(m.Content, " ")

		if len(commands) == 0 {
			fmt.Println(m.Author.Username + " / " + m.Author.ID + "找不到這項指令，若您希望註冊您在 MOPCON 2021 的身分，請您輸入: 「ticket [您的票號] [您的 Email]」。")
			s.ChannelMessageSend(m.ChannelID, "找不到這項指令，若您希望註冊您在 MOPCON 2021 的身分，請您輸入: 「ticket [您的票號] [您的 Email]」。")
			return
		}

		if commands[0] == "ticket" || commands[0] == "kktix" || commands[0] == "accupass" {
			// check commands
			if len(commands) != 3 {
				fmt.Println(m.Author.Username + " / " + m.Author.ID + "您還沒輸入正確的票號喔，您需要輸入: 「ticket [您的票號] [您的 Email]」，範例: 「ticket 123456789 test@test.com」。")
				s.ChannelMessageSend(m.ChannelID, "您還沒輸入正確的票號喔，您需要輸入: 「ticket [您的票號] [您的 Email]」，範例: 「ticket 123456789 test@test.com」。")
				return
			}

			// if this user is already registed, then pass it.
			for _, record := range usedToken {
				if record.User == string(m.Author.ID) {
					fmt.Println(m.Author.Username + " / " + m.Author.ID + "您已經有註冊過了，無法再次註冊，若有票務相關問題，請尋求服務台的協助")
					s.ChannelMessageSend(m.ChannelID, "您已經有註冊過了，無法再次註冊，若有票務相關問題，請尋求服務台的協助")
					return
				}
			}

			// block duplicate register
			if _, ok := usedToken[commands[1]]; ok {
				fmt.Println(m.Author.Username + " / " + m.Author.ID + "這個票號已經被註冊過了，請尋求服務台頻道的協助。")
				s.ChannelMessageSend(m.ChannelID, "這個票號已經被註冊過了，請尋求服務台頻道的協助。")
				return
			}

			// given badge and set used token
			if ticket, ok := KeyPairMap[commands[1]]; ok {
				// check email is correct
				if ticket.Email != commands[2] {
					fmt.Println(m.Author.Username + " / " + m.Author.ID + "此票種註冊資訊錯誤，請檢查您的票號或 Email 是否正確，或洽詢服務台。")
					s.ChannelMessageSend(m.ChannelID, "此票種註冊資訊錯誤，請檢查您的票號或 Email 是否正確，或洽詢服務台。")
					return
				}

				usedToken[commands[1]] = &Record{
					User: m.Author.ID,
					Time: time.Now().Format("2006-01-02 15:04:05"),
				}

				err = s.GuildMemberRoleAdd(GuildID, m.Author.ID, ticket.Badge)
				fmt.Println(err)
				fmt.Println(GuildID, m.Author.ID, ticket)

				fmt.Println(m.Author.Username + " / " + m.Author.ID + "您的身分已經完成設定，歡迎您回到 MOPCON Discord 會場!")
				s.ChannelMessageSend(m.ChannelID, "您的身分已經完成設定，歡迎您回到 MOPCON Discord 會場!")
			} else {
				fmt.Println(m.Author.Username + " / " + m.Author.ID + "這個票號不存在，請尋求服務台頻道的協助。")
				s.ChannelMessageSend(m.ChannelID, "這個票號不存在，請尋求服務台頻道的協助。")
				return
			}
		} else {
			fmt.Println(m.Author.Username + " / " + m.Author.ID + "這個指令不存在，您需要輸入: 「ticket [您的票號]」，範例: 「ticket 123456789」 以進行註冊。")
			s.ChannelMessageSend(m.ChannelID, "這個指令不存在，您需要輸入: 「ticket [您的票號]」，範例: 「ticket 123456789」 以進行註冊。")
			return
		}
	} else {
		// receive global message
		//s.ChannelMessageSend(m.ChannelID, "請您使用私訊的方式進行驗票註冊，謝謝您!")
	}
}

func ComesFromDM(s *discordgo.Session, m *discordgo.MessageCreate) (bool, error) {
	channel, err := s.State.Channel(m.ChannelID)
	if err != nil {
		if channel, err = s.Channel(m.ChannelID); err != nil {
			return false, err
		}
	}

	return channel.Type == discordgo.ChannelTypeDM, nil
}

func SaveUsedCSV() {

	csvData := [][]string{
		{"code", "user", "time"},
	}
	for key, val := range usedToken {
		csvData = append(csvData, []string{key, val.User, val.Time})
	}

	b := new(bytes.Buffer)
	w := csv.NewWriter(b)
	w.WriteAll(csvData)

	t := time.Now()
	os.WriteFile(t.Format("20060102150405")+".csv", b.Bytes(), 0644)

}
