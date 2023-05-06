package modules

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"minegram/utils"

	tb "gopkg.in/tucnak/telebot.v2"
)

// Auth module
// Sets up Telegram handles
// for commands related to
// authentication
func Auth(data utils.ModuleData) {
	(*data.TeleBot).Handle("/link", func(m *tb.Message) {
		if !*data.IsAuthEnabled {
			(*data.TeleBot).Reply(m, "The `auth` module has been disabled.", "Markdown")
			return
		}

		if m.Payload == "" {
			(*data.TeleBot).Reply(m, "Enter an IGN to link to your account!")
			return
		}

		plSplit := strings.Split(m.Payload, " ")
		ign := plSplit[0]

		var existing utils.Player
		(*data.GormDb).First(&existing, "mc_ign = ?", ign)
		if existing.McIgn != "" {
			if existing.TgID == m.Sender.ID {
				(*data.TeleBot).Reply(m, "You have already linked this IGN with your account!")
			} else {
				(*data.TeleBot).Reply(m, "This IGN has already been linked to a different Telegram account!")
			}
			return
		}

		var existingID utils.Player
		(*data.GormDb).First(&existingID, "tg_id = ?", m.Sender.ID)
		if existingID.TgID == 0 {
			fmt.Println("Linking '" + strconv.FormatInt(m.Sender.ID, 10) + "' to '" + ign + "'")
			(*data.GormDb).Create(&utils.Player{McIgn: ign, TgID: m.Sender.ID, LastGameMode: "survival", DidUserAuth: false})
			(*data.TeleBot).Reply(m, "The Minecraft IGN `"+ign+"` has been successfully linked to the telegram account with id `"+strconv.FormatInt(m.Sender.ID, 10)+"`!", "Markdown")
			return
		}

		oldIgn := existingID.McIgn
		if len(plSplit) != 2 {
			(*data.TeleBot).Reply(m, "Your account will be un-linked from `"+oldIgn+"` and linked to `"+ign+"`. To confirm this action, use:\n\n`/link "+ign+" confirm`", "Markdown")

			if strings.ToLower(plSplit[1]) == "confirm" {
				(*data.GormDb).Model(&existingID).Update("mc_ign", ign)
				(*data.TeleBot).Reply(m, "Your account has been un-linked from `"+oldIgn+"` and linked to `"+ign+"`.", "Markdown")
			} else {
				(*data.TeleBot).Reply(m, "The second argument must be '`confirm`'!", "Markdown")
			}
		}
	})

	(*data.TeleBot).Handle("/auth", func(m *tb.Message) {
		if !*data.IsAuthEnabled {
			(*data.TeleBot).Reply(m, "The `auth` module has been disabled.", "Markdown")
			return
		}

		if m.Payload != "" {
			(*data.TeleBot).Reply(m, "`auth` does not take any arguments.", "Markdown")
			return
		}

		var linked utils.Player
		(*data.GormDb).First(&linked, "tg_id = ?", m.Sender.ID)
		if linked.McIgn == "" {
			(*data.TeleBot).Reply(m, "You need to use `/link` before trying to `auth`", "Markdown")
			return
		}

		if !utils.ContainsPlayer(*data.OnlinePlayers, linked.McIgn) {
			(*data.TeleBot).Reply(m, "Your IGN (`"+linked.McIgn+"`) must be in-game to `auth`!", "Markdown")
			return
		}

		if linked.TgID != m.Sender.ID {
			(*data.TeleBot).Reply(m, "This Telegram account is not linked to the IGN `"+linked.McIgn+"`!", "Markdown")
			return
		}

		(*data.TeleBot).Reply(m, "You have successfully authenticated yourself as `"+linked.McIgn+"`!", "Markdown")
		utils.AuthOnlinePlayer(linked.McIgn, *data.OnlinePlayers)

		io.WriteString(*data.Stdin, "effect clear "+linked.McIgn+" blindness\n")

		if linked.DidUserAuth {
			// if user is authenticated set gametype to previous game type
			io.WriteString(*data.Stdin, "gamemode "+linked.LastGameMode+" "+linked.McIgn+"\n")
		} else {
			// if user disconnects during un-authenticated mode, set user gametype to survival
			io.WriteString(*data.Stdin, "gamemode survival "+linked.McIgn+"\n")
		}

		(*data.GormDb).Model(&linked).Update("did_user_auth", true)
	})
}
