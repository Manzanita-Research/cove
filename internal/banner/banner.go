package banner

import "fmt"

func Warm(msg string) {
	fmt.Printf("\033[38;5;215m%s\033[0m\n", msg)
}

func Dim(msg string) {
	fmt.Printf("\033[38;5;245m%s\033[0m\n", msg)
}
