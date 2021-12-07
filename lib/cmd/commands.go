package cmd

import (
	"regexp"
	"strings"
)

const (
	CreateCmd       = "create"
	ConnectCmd      = "connect"
	LeaveCmd        = "leave"
	DeleteCmd       = "delete"
	SuccessResponse = "success"
)

func GetWords(str string) []string {
	r := regexp.MustCompile("\\s+")
	replace := r.ReplaceAllString(str, " ")
	return strings.Split(replace, " ")
}
