package modules

import (
	"io"
	"strings"

	"minegram/utils"

	tb "gopkg.in/tucnak/telebot.v2"
)

func handleExtras(data utils.ModuleData, m *tb.Message, dataType string) {
	if len(*data.OnlinePlayers) > 0 {
		sender := strings.TrimSpace(strings.ReplaceAll(m.Sender.FirstName+" "+m.Sender.LastName, "\n", "(nl)"))
		content := dataType
		if m.IsReply() {
			if m.ReplyTo.Text == "" {
				m.ReplyTo.Text = "[unsupported]"
			}
			if m.Caption != "" {
				_, err = io.WriteString(*data.Stdin, "tellraw @a [\"\",{\"text\":\"[TG] "+sender+"\",\"color\":\"aqua\"},{\"text\":\": \"},{\"text\":\"(reply)\",\"color\":\"yellow\",\"hoverEvent\":{\"action\":\"show_text\",\"contents\":\""+m.ReplyTo.Text+"\"}},{\"text\":\" "+content+" - "+m.Caption+"\",\"color\":\"dark_gray\",\"italic\":true}]\n")
			} else {
				_, err = io.WriteString(*data.Stdin, "tellraw @a [\"\",{\"text\":\"[TG] "+sender+"\",\"color\":\"aqua\"},{\"text\":\": \"},{\"text\":\"(reply)\",\"color\":\"yellow\",\"hoverEvent\":{\"action\":\"show_text\",\"contents\":\""+m.ReplyTo.Text+"\"}},{\"text\":\" "+content+"\",\"color\":\"dark_gray\",\"italic\":true}]\n")
			}
		} else {
			if m.Caption != "" {
				_, err = io.WriteString(*data.Stdin, "tellraw @a [\"\",{\"text\":\"[TG] "+sender+"\",\"color\":\"aqua\"},{\"text\":\": \"},{\"text\":\""+content+" - "+m.Caption+"\",\"color\":\"dark_gray\",\"italic\":true}]\n")
			} else {
				_, err = io.WriteString(*data.Stdin, "tellraw @a [\"\",{\"text\":\"[TG] "+sender+"\",\"color\":\"aqua\"},{\"text\":\": \"},{\"text\":\""+content+"\",\"color\":\"dark_gray\",\"italic\":true}]\n")
			}
		}
	}
}

// TgToMc module
// Sends messages from Telegram
// to Minecraft with support
// for replies.
func TgToMc(data utils.ModuleData) {
	(*data.TeleBot).Handle(tb.OnText, func(m *tb.Message) {
		if len(*data.OnlinePlayers) > 0 {
			sender := strings.TrimSpace(strings.ReplaceAll(m.Sender.FirstName+" "+m.Sender.LastName, "\n", "(nl)"))
			content := strings.ReplaceAll(m.Text, "\n", "(nl)")
			if m.IsReply() {
				if m.ReplyTo.Text == "" {
					m.ReplyTo.Text = "[unsupported]"
				}
				_, err = io.WriteString(*data.Stdin, "tellraw @a [\"\",{\"text\":\"[TG] "+sender+"\",\"color\":\"aqua\"},{\"text\":\": \"},{\"text\":\"(reply)\",\"color\":\"yellow\",\"hoverEvent\":{\"action\":\"show_text\",\"contents\":\""+m.ReplyTo.Text+"\"}},{\"text\":\" "+content+"\"}]\n")
			} else {
				_, err = io.WriteString(*data.Stdin, "tellraw @a [\"\",{\"text\":\"[TG] "+sender+"\",\"color\":\"aqua\"},{\"text\":\": "+content+"\"}]\n")
			}
		}
	})

	(*data.TeleBot).Handle(tb.OnPhoto, func(m *tb.Message) {
		handleExtras(data, m, "photo")
	})

	(*data.TeleBot).Handle(tb.OnSticker, func(m *tb.Message) {
		handleExtras(data, m, "sticker")
	})

	(*data.TeleBot).Handle(tb.OnAnimation, func(m *tb.Message) {
		handleExtras(data, m, "gif")
	})

	(*data.TeleBot).Handle(tb.OnVideo, func(m *tb.Message) {
		handleExtras(data, m, "video")
	})

	(*data.TeleBot).Handle(tb.OnVoice, func(m *tb.Message) {
		handleExtras(data, m, "voice")
	})

	(*data.TeleBot).Handle(tb.OnDocument, func(m *tb.Message) {
		handleExtras(data, m, "file")
	})
}
