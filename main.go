package main

import (
	// "fmt"
	"fmt"
	"log"
	"sync"

	"github.com/catmandx/pocpeer/integrations"
	"github.com/catmandx/pocpeer/sources"
	"github.com/catmandx/pocpeer/utils"

	"github.com/catmandx/pocpeer/models"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	app := models.Application{}
	app.Db, err = utils.ConnectDb()
	if err != nil {
		log.Fatalln("Error connecting to Database!", err)
	}
	
	app.Sources, err = sources.LoadSources()
	if err != nil {
		log.Fatalln("Error loading sources!", err)
	}
	

	app.Sinks, err = integrations.LoadSinks()
	if err != nil {
		log.Fatalln("Error loading sinks!", err)
	}
	
	var wg sync.WaitGroup
	for _, source := range app.Sources {
		wg.Add(1)
		go func(){
			fmt.Println("abcd")
			defer wg.Done()
			source.Run(app)
		}()
	}
	wg.Wait()
}