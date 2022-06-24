package integrations

import 
(
	"fmt"
	"net/url"
	"io/ioutil"
	"log"
	"net/http"
	"github.com/catmandx/pocpeer/models"
)
type Telegram struct {
	ChannelName string
	ApiKey      string
}

func (o *Telegram) SendMessage(news models.News) {
	apiurl := "https://api.telegram.org/bot%s/sendMessage?chat_id=%s&text=%s"
	text := fmt.Sprintf("https://twitter.com/%s/status/%s\n%s", news.Author, news.PlatformId, news.Text)
	apiurl = fmt.Sprintf(apiurl, o.ApiKey, o.ChannelName, url.QueryEscape(text))
	
	resp, err := http.Get(apiurl)
	if err != nil {
	   log.Println(err)
	}
	
	if resp.StatusCode == 200{
		log.Println("Sent to Telegram channel!")
		return 
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
	   log.Println(err)
	}

	sb := string(body)
	log.Println(sb)
}