package console

import (
	"bufio"
	"os"
	"strings"
)

type Console struct {
	reader *bufio.Reader
}

func New() *Console {
	return &Console{
		reader: bufio.NewReader(os.Stdin),
	}
}

func (c *Console) ReadInput() (input string) {
	for {
		input, _ = c.reader.ReadString('\n')
		input = strings.Replace(input, "\n", "", -1)
		return
	}
}
