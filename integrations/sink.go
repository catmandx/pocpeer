package integrations

import (
	"os"

	"github.com/catmandx/pocpeer/models"
)

func LoadSinks() (sinks []models.Sink, err error) {
	slc := make([]models.Sink, 0)
	slc = append(slc, &Telegram{
		ApiKey: os.Getenv("TELEGRAM_APIKEY"), 
		ChannelName: os.Getenv("TELEGRAM_CHANNELNAME"),
	})
	return slc, nil
}

func SendMessageToAllSinks(application models.Application, news models.News){
	for _, sink := range application.Sinks{
		sink.SendMessage(news)
	}
}