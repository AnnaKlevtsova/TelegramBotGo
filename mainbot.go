package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/Syfaro/telegram-bot-api"
)

var (
	// глобальные переменные
	telegramBotToken = "5851601199:AAGTcLdoxTQUnSpi7ycS8kZ7DRr_aJstcuo"
	bot              *tgbotapi.BotAPI
	reply            string
	farenTemp        float64
)

const (
	dumpRaw = false
	zip     = "101000,ru"
	api     = "08f2a575dda978b9c539199e54df03b0"
)

var (
	weatherKeys = map[string]bool{"main": false, "wind": false, "coord": false, "weather": true, "sys": false, "clouds": false}
)

func weather(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	urlString := fmt.Sprintf("http://api.openweathermap.org/data/2.5/weather?zip=%s&APPID=%s", zip, api)
	u, err := url.Parse(urlString)
	res, err := http.Get(u.String())
	if err != nil {
		log.Fatal(err)
	}

	jsonBlob, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}

	var data map[string]interface{}

	if dumpRaw {
		fmt.Printf("blob = %s\n\n", jsonBlob)
	}
	err = json.Unmarshal(jsonBlob, &data)
	if err != nil {
		fmt.Println("error:", err)
	}

	if dumpRaw {
		fmt.Printf("%+v", data)
	}
	for k, v := range data {
		val, isAnArray := isKey(k)
		if val {
			dumpMap(k, v, isAnArray, bot, update)
		} else {
		}
	}
}

func dumpMap(k string, v interface{}, isArray bool, bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	if isArray {
		for i := 0; i < len(v.([]interface{})); i++ {
			nmap := v.([]interface{})[i].(map[string]interface{})
			for kk, vv := range nmap {
				if kk == "description" {
					reply = "Состояние погоды: "
					if vv == "overcast clouds" {
						reply = reply + "Пасмурно"
					}
					if vv == "clear" {
						reply = reply + "Ясно"
					}
					if vv == "cloudy" {
						reply = reply + "Облачно"
					}
					if vv == "rain" {
						reply = reply + "Дождь"
					}
					if vv == "thunderstorm" {
						reply = reply + "Гроза"
					}
					if vv == "snow" {
						reply = reply + "Снег"
					}
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, reply)
					bot.Send(msg)
				}
			}

		}
	} else {
		nmap := v.(map[string]interface{})
		for kk, vv := range nmap {
			if isTempVal(kk) {
				farenTemp := faren(vv.(float64))
				if kk == "temp_max" {
					reply = "Максимальная температура: " + strconv.FormatFloat(farenTemp, 'f', 2, 64) + "°C"
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, reply)
					bot.Send(msg)
				}
				if kk == "temp_min" {
					reply = "Минимальная температура: " + strconv.FormatFloat(farenTemp, 'f', 2, 64) + "°C"
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, reply)
					bot.Send(msg)
				}
				if kk == "temp" {
					reply = "Средняя температура: " + strconv.FormatFloat(farenTemp, 'f', 2, 64) + "°C"
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, reply)
					bot.Send(msg)
				}
			} else if isSunVal(kk) {
				sunTime := time.Unix((int64(vv.(float64))), 0)

				if kk == "sunrise" {
					reply = "Восход: " + sunTime.Format("2006-01-02 15:04:05")
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, reply)
					bot.Send(msg)
				}
				if kk == "sunset" {
					reply = "Закат: " + sunTime.Format("2006-01-02 15:04:05")
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, reply)
					bot.Send(msg)
				}
			}
		}
	}
}

func isKey(k string) (ok bool, isArray bool) {
	isArray, ok = weatherKeys[k]
	return ok, isArray
}

func faren(kelvin float64) float64 {
	return (kelvin - 273.0)
}

func isTempVal(k string) bool {
	return strings.Contains(k, "temp")
}

func isSunVal(k string) bool {
	return strings.Contains(k, "sun")
}

func init() {
	flag.Parse()
	if telegramBotToken == "" {
		log.Print("-telegrambottoken is required")
		os.Exit(1)
	}
}

func main() {
	bot, err := tgbotapi.NewBotAPI(telegramBotToken)
	if err != nil {
		log.Panic(err)
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)
	// u - структура с конфигом для получения апдейтов
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	// используя конфиг u создаем канал в котором будут новые сообщения
	updates := bot.GetUpdatesChan(u)
	// обрабатываем
	for update := range updates {
		reply := "Такой команды нет"
		if update.Message == nil {
			continue
		}

		// от кого какое сообщение пришло
		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		// обработка комманд
		switch update.Message.Command() {
		case "Инструкция":
			reply = "Привет! Я создан для того, чтобы узнавать текущую погоду в Москве.\nУ меня есть несколько команд для того, чтобы тебе было удобнее со мной работать.\n/help - Инструкция\n/weather - Узнать погоду"
		case "weather":
			reply = ""
			t := time.Now()
			reply = "Город Москва \nТекущее время: " + t.Format("2006-01-02 15:04:05")
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, reply)
			bot.Send(msg)
			reply = ""
			weather(bot, update)
		case "start":
			reply = "Привет! Я создан для того, чтобы узнавать текущую погоду в Москве.\nУ меня есть несколько команд для того, чтобы тебе было удобнее со мной работать.\n/help - Инструкция\n/weather - Узнать погоду"
		case "help":
			reply = "Привет! Я создан для того, чтобы узнавать текущую погоду в Москве.\nУ меня есть несколько команд для того, чтобы тебе было удобнее со мной работать.\n/help - Инструкция\n/weather - Узнать погоду"
		}
		// ответ
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, reply)
		// отправка
		bot.Send(msg)
	}
}
