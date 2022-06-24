package utils

import (
	"regexp"
)

func GetEnv(env string){

}

func ExtractCveNum(text string) []string {
	r, _ := regexp.Compile(`((?i)CVE-\d{3,4}-\d{1,10})`)
	return r.FindAllString(text, -1)
}