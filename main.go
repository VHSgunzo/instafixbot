package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var err error
var bot *tgbotapi.BotAPI

func send_msg(msg tgbotapi.MessageConfig) bool {
	if _, err := bot.Send(msg); err != nil {
		log.Panic(err)
	}
	return true
}

func main() {

	token := os.Getenv("TGBOT_TOKEN")
	if token == "" {
		log.Fatal("TGBOT_TOKEN переменная окружения не задана!")
	}

	// bot, err = tgbotapi.NewBotAPI(token)
	bot, err = tgbotapi.NewBotAPIWithClient(token, "https://api.telegram.org/bot%s/%s", &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	})
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = false

	log.Printf("Бот авторизован в учетной записи: %s", bot.Self.UserName)

	_, err = bot.Request(tgbotapi.DeleteWebhookConfig{})
	if err != nil {
		log.Panic(err)
	}

	update := tgbotapi.NewUpdate(0)
	update.Timeout = 60

	updates := bot.GetUpdatesChan(update)

	// for render.com
	go func() {
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Бот работает!"))
		})
		http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(fmt.Sprintf("Бот работает! Очередь обновлений: %d", len(updates))))
		})
		log.Fatal(http.ListenAndServe(":7860", nil))
	}()

	ticker := time.NewTicker(300 * time.Second)
	go func() {
		client := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		}
		for range ticker.C {
			// resp, err := client.Get("http://localhost:7860/status")
			resp, err := client.Get("https://instafixbot.onrender.com/status")
			if err != nil {
				log.Println(err)
				continue
			}
			if resp.StatusCode != http.StatusOK {
				log.Println("Ошибка проверки статуса:", resp.StatusCode)
			} else {
				log.Println("Статус бота:", resp.Status)
			}
			if err := resp.Body.Close(); err != nil {
				log.Println(err)
			}
		}
	}()
	// for render.com

	var msg tgbotapi.MessageConfig
	for update := range updates {
		if update.Message != nil {
			updID := update.Message.From.ID
			updChatID := update.Message.Chat.ID
			updMsgText := update.Message.Text
			updUserName := update.Message.From.UserName
			log.Printf("[%d]:[%s]: %s", updID, updUserName, updMsgText)

			if strings.HasPrefix(updMsgText, "https://www.instagram.com") {
				updMsgText := strings.Replace(updMsgText, "instagram", "ddinstagram", 1)
				updMsgText = fmt.Sprintf("<a href=\"%s\">ㅤ</a>", updMsgText)
				msg = tgbotapi.NewMessage(updChatID, updMsgText)
				msg.ParseMode = "HTML"
			} else {
				msg = tgbotapi.NewMessage(updChatID, "Отправь мне ссылку из инсты)")
			}

			send_msg(msg)
		}
	}
}
