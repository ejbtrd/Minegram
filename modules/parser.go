package modules

import (
	"bufio"
	"io"
	"regexp"
	"strings"
	"time"

	"minegram/utils"
)

var (
	chatRegex            = regexp.MustCompile(`(?:\[Not Secure\] )?<(.+)> (.+)`)
	joinRegex            = regexp.MustCompile(`: (.+) joined the game`)
	joinRegexSpigotPaper = regexp.MustCompile(`: UUID of player (.+) is .*`)
	leaveRegex           = regexp.MustCompile(`: (.+) left the game`)
	advancementRegex     = regexp.MustCompile(`: (.+) has made the advancement (.+)`)
)

/* death regex taken from https://github.com/trgwii/TeMiCross/blob/master/client/parser/default/messages/death.js */
var deathRegex = regexp.MustCompile(`: (.+) (was (shot by .+|shot off (some vines|a ladder) by .+|pricked to death|stabbed to death|squished too much|blown up by .+|killed by .+|doomed to fall by .+|blown from a high place by .+|squashed by .+|burnt to a crisp whilst fighting .+|roasted in dragon breath( by .+)?|struck by lightning( whilst fighting .+)?|slain by .+|fireballed by .+|killed trying to hurt .+|impaled by .+|speared by .+|poked to death by a sweet berry bush( whilst trying to escape .+)?|pummeled by .+)|hugged a cactus|walked into a cactus whilst trying to escape .+|drowned( whilst trying to escape .+)?|suffocated in a wall( whilst fighting .+)?|experienced kinetic energy( whilst trying to escape .+)?|removed an elytra while flying( whilst trying to escape .+)?|blew up|hit the ground too hard( whilst trying to escape .+)?|went up in flames|burned to death|walked into fire whilst fighting .+|went off with a bang( whilst fighting .+)?|tried to swim in lava(( while trying)? to escape .+)?|discovered floor was lava|walked into danger zone due to .+|got finished off by .+|starved to death|didn't want to live in the same world as .+|withered away( whilst fighting .+)?|died( because of .+)?|fell (from a high place( and fell out of the world)?|off a ladder|off to death( whilst fighting .+)?|off some vines|out of the water|into a patch of fire|into a patch of cacti|too far and was finished by .+|out of the world))$`)

var (
	timeRegex          = regexp.MustCompile(`: The time is (.+)`)
	entityPosRegex     = regexp.MustCompile(`: .+ has the following entity data: \[(.+)d, (.+)d, (.+)d\]`)
	simplifiedEPRegex  = regexp.MustCompile(`: .+ has the following entity data: \[(.+)\..*d, (.+)\..*d, (.+)\..*d\]`)
	simpleOutputRegex  = regexp.MustCompile(`.*: (.+)`)
	dimensionRegex     = regexp.MustCompile(`.*has the following entity data: "(minecraft:.+)"`)
	gameTypeRegex      = regexp.MustCompile(`.*has the following entity data: (.+)`)
	genericOutputRegex = regexp.MustCompile(`(\[.+\]) (\[.+\]): (.+)`)
)

// Parser module
// Parses Minecraft server cli
// log and acts as necessary
func Parser(data utils.ModuleData) {
	scanner := bufio.NewScanner(*data.Stdout)
	go func() {
		defer (*data.Waitgroup).Done()
		for scanner.Scan() {
			message := scanner.Text()
			logFeed <- message

			if *data.NeedResult {
				*data.ConsoleOut <- message
				*data.NeedResult = false
			} else {
				go func() {
					if !strings.Contains(message, "INFO") {
						return
					}

					if chatRegex.MatchString(message) {
						result := chatRegex.FindStringSubmatch(message)

						if len(result) != 3 {
							return
						}

						_, _ = (*data.TeleBot).Send(*data.TargetChat, "`"+result[1]+"`"+"**:** "+result[2], "Markdown")
					} else if joinRegex.MatchString(message) || joinRegexSpigotPaper.MatchString(message) {
						result := joinRegex.FindStringSubmatch(message)

						if len(result) != 2 {
							return
						}

						user := result[1]

						if utils.ContainsPlayer(*data.OnlinePlayers, user) {
							return
						}

						newPlayer := utils.OnlinePlayer{InGameName: user, IsAuthd: false}
						*data.OnlinePlayers = append(*data.OnlinePlayers, newPlayer)
						toSend := "`" + user + "`" + " joined the server."
						if *data.IsAuthEnabled {
							toSend += "\nUse /auth to authenticate."
						}
						_, _ = (*data.TeleBot).Send(*data.TargetChat, toSend, "Markdown")
						if !*data.IsAuthEnabled {
							return
						}
						var currentUser utils.Player
						(*data.GormDb).First(&currentUser, "mc_ign = ?", user)

						startCoords := utils.CliExec(*data.Stdin, "data get entity "+user+" Pos", data.NeedResult, *data.ConsoleOut)
						coords := entityPosRegex.FindStringSubmatch(startCoords)

						dimensionStr := utils.CliExec(*data.Stdin, "data get entity "+user+" Dimension", data.NeedResult, *data.ConsoleOut)
						dimension := dimensionRegex.FindStringSubmatch(dimensionStr)

						gameTypeStr := utils.CliExec(*data.Stdin, "data get entity "+user+" playerGameType", data.NeedResult, *data.ConsoleOut)
						rGameType := gameTypeRegex.FindStringSubmatch(gameTypeStr)

						gameType := "survival"
						if len(rGameType) > 0 {
							gameType = utils.GetGameType(rGameType[1])
						}

						(*data.GormDb).Model(&currentUser).Update("last_game_mode", gameType)
						(*data.GormDb).Model(&currentUser).Update("did_user_auth", false)

						_, _ = io.WriteString(*data.Stdin, "effect give "+user+" minecraft:blindness 999999\n")
						_, _ = io.WriteString(*data.Stdin, "gamemode spectator "+user+"\n")
						_, _ = io.WriteString(*data.Stdin, "tellraw "+user+" [\"\",{\"text\":\"If you haven't linked before, send \"},{\"text\":\"/link "+newPlayer.InGameName+" \",\"color\":\"green\"},{\"text\":\"to \"},{\"text\":\"@"+(*data.TeleBot).Me.Username+"\",\"color\":\"yellow\"},{\"text\":\"\\nIf you have \"},{\"text\":\"linked \",\"color\":\"green\"},{\"text\":\"your account, send \"},{\"text\":\"/auth \",\"color\":\"aqua\"},{\"text\":\"to \"},{\"text\":\"@"+(*data.TeleBot).Me.Username+"\",\"color\":\"yellow\"}]\n")

						if len(coords) != 4 || len(dimension) != 2 {
							return
						}

						for {
							player := utils.GetOnlinePlayer(user, *data.OnlinePlayers)

							if player.IsAuthd || player.InGameName == "" {
								break
							}

							command := "execute in " + dimension[1] + " run tp " + user + " " + coords[1] + " " + coords[2] + " " + coords[3] + "\n"
							_, _ = io.WriteString(*data.Stdin, command)
							time.Sleep(400 * time.Millisecond)
						}
					} else if leaveRegex.MatchString(message) {
						result := leaveRegex.FindStringSubmatch(message)

						if len(result) != 2 {
							return
						}

						*data.OnlinePlayers = utils.RemovePlayer(*data.OnlinePlayers, result[1])
						_, _ = (*data.TeleBot).Send(*data.TargetChat, "`"+result[1]+"`"+" has left the server.", "Markdown")
					} else if advancementRegex.MatchString(message) {
						result := advancementRegex.FindStringSubmatch(message)

						if len(result) != 3 {
							return
						}

						_, _ = (*data.TeleBot).Send(*data.TargetChat, "`"+result[1]+"`"+" has made the advancement `"+result[2]+"`.", "Markdown")
					} else if deathRegex.MatchString(message) {
						result := simpleOutputRegex.FindStringSubmatch(message)

						if len(result) != 2 {
							return
						}

						sep := strings.Split(result[1], " ")
						startCoords := utils.CliExec(*data.Stdin, "data get entity "+sep[0]+" Pos", data.NeedResult, *data.ConsoleOut)
						coords := simplifiedEPRegex.FindStringSubmatch(startCoords)
						toSend := "`" + sep[0] + "` " + strings.Join(sep[1:], " ")

						if len(coords) == 4 {
							toSend += " at (`" + coords[1] + " " + coords[2] + " " + coords[3] + "`)"
						}

						_, _ = (*data.TeleBot).Send(*data.TargetChat, toSend+".", "Markdown")
					} else if strings.Contains(message, "For help, type") {
						utils.CliExec(*data.Stdin, "say Server initialised!", data.NeedResult, *data.ConsoleOut)
						_, _ = (*data.TeleBot).Send(*data.TargetChat, "**Server started!**", "Markdown")
					} else if strings.Contains(message, "All dimensions are saved") {
						_, _ = (*data.TeleBot).Send(*data.TargetChat, "**Server stopped!**", "Markdown")
					}
				}()
			}
		}
	}()
}
