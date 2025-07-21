package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/go-telegram/bot"
)

const (
	token    = "7209712528:AAF7o20ysTcpgQb8JlVH4_CLmqH_iz5GiL8"
	chatID   = -4709535234
	timezone = "Asia/Tashkent"
)

type Swords struct {
	Sword5nm           int `json:"sword5nm"`
	Sword7nm           int `json:"sword7nm"`
	Sword5             int `json:"sword5"`
	Sword6             int `json:"sword6"`
	Sword7             int `json:"sword7"`
	Megasword          int `json:"megasword"`
	NetheriteLeggings  int `json:"netherite_leggings"`
	NetheriteChestplate int `json:"netherite_chestplate"`
	NetheriteHelmet    int `json:"netherite_helmet"`
	NetheriteBoots     int `json:"netherite_boots"`
}

type DailyData struct {
	Date      string `json:"date"`
	MessageID int    `json:"message_id"`
	Swords    Swords `json:"swords"`
}

var (
	tgBot     *bot.Bot
	data      DailyData
	dataMutex sync.Mutex
	loc       *time.Location
)

func init() {
	var err error
	loc, err = time.LoadLocation(timezone)
	if err != nil {
		log.Fatalf("Error loading location: %v", err)
	}
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	loadData()

	b, err := bot.New(token)
	if err != nil {
		log.Fatalf("Error creating bot: %v", err)
	}
	tgBot = b

	initTelegramMessage(ctx)

	http.HandleFunc("/update", updateHandler)
	go func() {
		log.Println("Server started on :8080")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	go dailyResetChecker(ctx)

	<-ctx.Done()
}

func loadData() {
	today := time.Now().In(loc).Format("2006-01-02")
	filename := fmt.Sprintf("data_%s.json", today)

	file, err := os.ReadFile(filename)
	if err != nil {
		data = DailyData{
			Date:   today,
			Swords: Swords{},
		}
		return
	}

	if err := json.Unmarshal(file, &data); err != nil {
		log.Printf("Error decoding data file: %v", err)
		data = DailyData{
			Date:   today,
			Swords: Swords{},
		}
	}
}

func saveData() {
	today := time.Now().In(loc).Format("2006-01-02")
	filename := fmt.Sprintf("data_%s.json", today)

	file, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		log.Printf("Error encoding data: %v", err)
		return
	}

	if err := os.WriteFile(filename, file, 0644); err != nil {
		log.Printf("Error saving data: %v", err)
	}
}

func initTelegramMessage(ctx context.Context) {
	if data.MessageID == 0 {
		msg, err := tgBot.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   generateMessageText(),
		})
		if err != nil {
			log.Printf("Error sending message: %v", err)
			return
		}
		data.MessageID = msg.ID
		saveData()
	} else {
		_, err := tgBot.EditMessageText(ctx, &bot.EditMessageTextParams{
			ChatID:    chatID,
			MessageID: data.MessageID,
			Text:      generateMessageText(),
		})
		if err != nil {
			log.Printf("Error updating message: %v", err)
		}
	}
}

func updateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		Type string `json:"type"` // Используем поле type вместо id
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	dataMutex.Lock()
	defer dataMutex.Unlock()

	switch request.Type { // Обрабатываем поле type
	case "5nomend":
		data.Swords.Sword5nm++
	case "7nomend":
		data.Swords.Sword7nm++
	case "sword5":
		data.Swords.Sword5++
	case "sword6":
		data.Swords.Sword6++
	case "sword7":
		data.Swords.Sword7++
	case "megasword":
		data.Swords.Megasword++
	case "бошмаки":
		data.Swords.NetheriteBoots++
	case "шлем":
		data.Swords.NetheriteHelmet++
	case "нагрудник":
		data.Swords.NetheriteChestplate++
	case "штаны":
		data.Swords.NetheriteLeggings++
	default:
		http.Error(w, "Invalid sword type", http.StatusBadRequest)
		return
	}

	saveData()
	updateTelegramMessage(context.Background())

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(data.Swords)
}

func updateTelegramMessage(ctx context.Context) {
	_, err := tgBot.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:    chatID,
		MessageID: data.MessageID,
		Text:      generateMessageText(),
	})
	if err != nil {
		log.Printf("Error updating Telegram message: %v", err)
	}
}

func generateMessageText() string {
	today := time.Now().In(loc).Format("2006-01-02")
	return fmt.Sprintf("🗡 Статистика за %s:\n\n"+
		"5nm: %d\n7nm: %d\n"+
		"5: %d\n6: %d\n"+
		"7: %d\nMEGA: %d\n\n"+
		"Ботинки: %d\nШлем: %d\n"+
		"Нагрудник: %d\nШтаны: %d",
		today,
		data.Swords.Sword5nm, data.Swords.Sword7nm,
		data.Swords.Sword5, data.Swords.Sword6,
		data.Swords.Sword7, data.Swords.Megasword,
		data.Swords.NetheriteBoots, data.Swords.NetheriteHelmet,
		data.Swords.NetheriteChestplate, data.Swords.NetheriteLeggings)
}

func dailyResetChecker(ctx context.Context) {
	for {
		now := time.Now().In(loc)
		nextMidnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, loc)
		duration := nextMidnight.Sub(now)

		select {
		case <-time.After(duration):
			dataMutex.Lock()
			data = DailyData{
				Date:   time.Now().In(loc).Format("2006-01-02"),
				Swords: Swords{},
			}
			dataMutex.Unlock()

			initTelegramMessage(ctx)

		case <-ctx.Done():
			return
		}
	}
}