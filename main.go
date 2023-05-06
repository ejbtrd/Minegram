package main

import (
	"io"
	"os/exec"
	"reflect"
	"runtime"
	"sync"

	"minegram/modules"
	"minegram/utils"

	"github.com/fatih/color"

	tb "gopkg.in/tucnak/telebot.v2"
	"gorm.io/gorm"
)

/* configuration options */
var (
	cmd         string
	tok         string
	admUsers    []string
	authEnabled = true
)

/* shared vars */
var (
	online     = []utils.OnlinePlayer{}
	cliOutput  = make(chan string)
	needResult = false
	db         *gorm.DB
	b          *tb.Bot
	execCmd    *exec.Cmd
	stdin      io.WriteCloser
	stdout     io.ReadCloser
	targetChat tb.Recipient
	wg         sync.WaitGroup
)

/* shared error */
var err error

func plugModule(module utils.ModuleFunction) {
	color.Blue("LOADING MODULE: " + runtime.FuncForPC(reflect.ValueOf(module).Pointer()).Name())
	module(utils.ModuleData{
		CmdToRun: &cmd, TgBotToken: &tok, AdminUsers: &admUsers,
		IsAuthEnabled: &authEnabled, OnlinePlayers: &online,
		ConsoleOut: &cliOutput, NeedResult: &needResult,
		GormDb: &db, TeleBot: &b, ExecCmd: &execCmd, Stdin: &stdin,
		Stdout: &stdout, TargetChat: &targetChat, Waitgroup: &wg,
	})
}

func main() {
	plugModule(modules.Core)
	plugModule(modules.Parser)
	plugModule(modules.TgUtilCommands)
	plugModule(modules.TgToMc)
	plugModule(modules.Auth)
	plugModule(modules.Logger)
	plugModule(modules.Init)
}
