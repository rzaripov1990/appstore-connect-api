package console

import "fmt"

type newLine struct{}

func (nl newLine) NL() *newLine {
	fmt.Println()
	return &newLine{}
}

func PrintGreen(format string, args ...any) *newLine {
	fmt.Printf("%s", GREEN)
	fmt.Printf(format, args...)
	fmt.Printf("%s", NC)
	return &newLine{}
}

func PrintGray(format string, args ...any) *newLine {
	fmt.Printf("%s", GRAY)
	fmt.Printf(format, args...)
	fmt.Printf("%s", NC)
	return &newLine{}
}

func PrintRed(format string, args ...any) *newLine {
	fmt.Printf("%s", RED)
	fmt.Printf(format, args...)
	fmt.Printf("%s", NC)
	return &newLine{}
}

func Print(format string, args ...any) *newLine {
	fmt.Printf(format, args...)
	return &newLine{}
}
