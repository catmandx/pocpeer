package sources

import (
	"log"

	"github.com/catmandx/pocpeer/models"
)

func LoadSources() (sources []models.Source, err error) {
	slc := make([]models.Source, 0)
	twt := &Twitter{}
	initErr := twt.Init()
	if initErr != nil {
		log.Println(initErr)
	}else{
		slc = append(slc, twt)
	}
	return slc, nil
}