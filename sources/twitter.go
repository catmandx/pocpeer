package sources

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/catmandx/pocpeer/integrations"
	"github.com/catmandx/pocpeer/models"
	"github.com/catmandx/pocpeer/utils"
	twitterstream "github.com/fallenstedt/twitter-stream"
	"github.com/fallenstedt/twitter-stream/rules"
	"github.com/fallenstedt/twitter-stream/stream"
)
type Twitter struct{
	ConsumerKey 	string
	ConsumerSecret 	string
	// AccessToken		string
	// AccessSecret	string
}

type StreamData struct {
	Data struct {
		Text      string    `json:"text"`
		ID        string    `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		AuthorID  string    `json:"author_id"`
		Entities  struct {
			Urls  []struct {
				ExpandedUrl string	`json:"expanded_url"`
			} `json:"urls"`
		} `json:"entities"`
	} `json:"data"`
	Includes struct {
		Users []struct {
			ID       string `json:"id"`
			Name     string `json:"name"`
			Username string `json:"username"`
		} `json:"users"`
	} `json:"includes"`
	MatchingRules []struct {
		ID  string `json:"id"`
		Tag string `json:"tag"`
	} `json:"matching_rules"`
}

func (o *Twitter) Init() (err error) {
	o.ConsumerKey = os.Getenv("TWITTER_CONSUMERAPIKEY")
	o.ConsumerSecret = os.Getenv("TWITTER_CONSUMERAPISECRET")
	if o.ConsumerKey == "" && o.ConsumerSecret == "" {
		return errors.New("unable to initialize Twitter")
	}
	o.deleteRules()
	o.addRules()
	return
}

// This will run forever
func (o *Twitter) Run(application models.Application) {
	fmt.Println("Starting Stream")

	// Start the stream
	// And return the library's api
	api := o.fetchTweets()

	// When the loop below ends, restart the stream
	defer o.Run(application)

	// Start processing data from twitter
	for tweet := range api.GetMessages() {

		// Handle disconnections from twitter
		// https://developer.twitter.com/en/docs/twitter-api/tweets/volume-streams/integrate/handling-disconnections
		if tweet.Err != nil {
			fmt.Printf("got error from twitter: %v", tweet.Err)

			// Notice we "StopStream" and then "continue" the loop instead of breaking.
			// StopStream will close the long running GET request to Twitter's v2 Streaming endpoint by
			// closing the `GetMessages` channel. Once it's closed, it's safe to perform a new network request
			// with `StartStream`
			api.StopStream()
			continue
		}
		result := tweet.Data.(StreamData)

		news := models.News{}
		news.Platform = "Twitter"
		news.PlatformId = result.Data.ID
		news.Author = result.Includes.Users[0].Username
		news.Text = result.Data.Text
		news.CveNum = strings.Join(utils.ExtractCveNum(news.Text), " ")
		if len(result.Data.Entities.Urls) > 0{
			slc := make([]string, 0)
			for _, url := range result.Data.Entities.Urls {
				slc = append(slc, url.ExpandedUrl)
			}
			news.Links = strings.Join(slc, " ")
		}
		integrations.SendMessageToAllSinks(application, news)
	}

	fmt.Println("Stopped Stream")
}

func (o *Twitter) fetchTweets() stream.IStream {
	// Get Bearer Token using API keys
	tok, err := o.getTwitterToken()
	if err != nil {
		panic(err)
	}

	// Instantiate an instance of twitter stream using the bearer token
	api := o.getTwitterStreamApi(tok)

	// On Each tweet, decode the bytes into a StreamDataExample struct
	api.SetUnmarshalHook(func(bytes []byte) (interface{}, error) {
		fmt.Println("Byte arr to string: ", string(bytes[:]))
		data := StreamData{}
		if err := json.Unmarshal(bytes, &data); err != nil {
			fmt.Printf("failed to unmarshal bytes: %v", err)
		}
		return data, err
	})

	// Request additional data from each tweet
	streamExpansions := twitterstream.NewStreamQueryParamsBuilder().
		AddExpansion("author_id").
		AddTweetField("created_at").
		AddTweetField("entities").
		Build()

	// Start the Stream
	err = api.StartStream(streamExpansions)
	if err != nil {
		panic(err)
	}

	// Return the twitter stream api instance
	return api
}

func (o *Twitter) getTwitterToken() (string, error) {
	tok, err := twitterstream.NewTokenGenerator().SetApiKeyAndSecret(o.ConsumerKey, o.ConsumerSecret).RequestBearerToken()
	return tok.AccessToken, err
}

func (o *Twitter) getTwitterStreamApi(tok string) stream.IStream {
	return twitterstream.NewTwitterStream(tok).Stream
}

func (o *Twitter) addRules() {

	tok, err := twitterstream.NewTokenGenerator().SetApiKeyAndSecret(o.ConsumerKey, o.ConsumerSecret).RequestBearerToken()
	if err != nil {
		panic(err)
	}
	api := twitterstream.NewTwitterStream(tok.AccessToken)
	rules := twitterstream.NewRuleBuilder().
		AddRule("cve poc -is:retweet", "CVE POC").
		// AddRule("vulnerability poc -is:retweet", "VULN POC").
		// AddRule("lang:en  -is:quote (#golangjobs OR #gojobs)", "golang jobs").
		Build()

	res, err := api.Rules.Create(rules, false) // dryRun is set to false.

	if err != nil {
		panic(err)
	}

	if res.Errors != nil && len(res.Errors) > 0 {
		//https://developer.twitter.com/en/support/twitter-api/error-troubleshooting
		panic(fmt.Sprintf("Received an error from twitter: %v", res.Errors))
	}

	fmt.Println("I have created rules.")
	o.printRules(res.Data)
}

func (o *Twitter) getRules() []rules.DataRule {
	tok, err := twitterstream.NewTokenGenerator().SetApiKeyAndSecret(o.ConsumerKey, o.ConsumerSecret).RequestBearerToken()
	if err != nil {
		panic(err)
	}
	api := twitterstream.NewTwitterStream(tok.AccessToken)
	res, err := api.Rules.Get()

	if err != nil {
		panic(err)
	}

	if res.Errors != nil && len(res.Errors) > 0 {
		//https://developer.twitter.com/en/support/twitter-api/error-troubleshooting
		panic(fmt.Sprintf("Received an error from twitter: %v", res.Errors))
	}

	if len(res.Data) > 0 {
		return res.Data
	} else {
		return make([]rules.DataRule, 0)
	}
}

func (o *Twitter) deleteRules() {
	tok, err := twitterstream.NewTokenGenerator().SetApiKeyAndSecret(o.ConsumerKey, o.ConsumerSecret).RequestBearerToken()
	if err != nil {
		panic(err)
	}
	api := twitterstream.NewTwitterStream(tok.AccessToken)

	listOfRules := o.getRules()
	idArr := make([]int, len(listOfRules))
	for i, rule := range listOfRules {
		intVal,err := strconv.ParseInt(rule.Id, 10, 64)
		if err != nil {
			fmt.Println("Error parsing number", rule.Id)
		}
		idArr[i] = int(intVal)
	}
	// use api.Rules.Get to find the ID number for an existing rule
	res, err := api.Rules.Delete(rules.NewDeleteRulesRequest(idArr...), false)

	if err != nil {
		panic(err)
	}

	if res.Errors != nil && len(res.Errors) > 0 {
		//https://developer.twitter.com/en/support/twitter-api/error-troubleshooting
		panic(fmt.Sprintf("Received an error from twitter: %v", res.Errors))
	}

	fmt.Println("I have deleted rules ")
}


func (o *Twitter) printRules(data []rules.DataRule) {
	for _, datum := range data {
		fmt.Printf("Id: %v\n", datum.Id)
		fmt.Printf("Tag: %v\n",datum.Tag)
		fmt.Printf("Value: %v\n\n", datum.Value)
	}
}
