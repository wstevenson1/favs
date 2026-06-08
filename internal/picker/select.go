package picker

import (
	"errors"
	"strconv"
	"strings"
)

var ErrQuit = errors.New("quit")
var ErrInvalid = errors.New("invalid selection")

func ParseSelection(input string, count int) (int, error) {
	input = strings.TrimSpace(input)
	if input == "q" || input == "Q" {
		return 0, ErrQuit
	}
	n, err := strconv.Atoi(input)
	if err != nil || n < 1 || n > count {
		return 0, ErrInvalid
	}
	return n, nil
}
