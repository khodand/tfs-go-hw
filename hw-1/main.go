package main

import (
	"fmt"
)

// settings numbers for symbol
const (
	CHAR          = iota
	SIZE          = iota
	COLOR_START   = iota
	COLOR_END     = iota
	SETTINGS_SIZE = iota
)

type (
	Symbol []string
	Mod    func(s *Symbol)
)

func printSymbol(s Symbol) {
	fmt.Print(s[COLOR_START] + s[CHAR] + s[COLOR_END])
}

func printLine(size int, s Symbol) {
	for j := 0; j < size; j++ {
		printSymbol(s)
	}
	fmt.Println()
}

func setColor(color int) Mod {
	return func(s *Symbol) {
		(*s)[COLOR_START] = fmt.Sprintf("\u001B[%vm", color)
		(*s)[COLOR_END] = "\u001B[0m"
	}
}

func setChar(char string) Mod {
	return func(s *Symbol) {
		(*s)[0] = char
	}
}

func setSize(size int) Mod {
	return func(s *Symbol) {
		(*s)[SIZE] = fmt.Sprintf("%v", size)
	}
}

func sandglass(mods ...Mod) {
	symbol := make(Symbol, SETTINGS_SIZE)
	symbol[CHAR] = "X"
	symbol[SIZE] = "1"

	for _, mod := range mods {
		mod(&symbol)
	}

	width := 0
	_, err := fmt.Sscan(symbol[SIZE], &width)
	if err != nil {
		return
	}

	w := width - 4
	halfHeight := int(float32(width)/2 + 0.5)

	printLine(width, symbol)
	for i := 1; i < halfHeight; i++ {
		for j := i; j > 0; j-- {
			fmt.Print(" ")
		}
		if w != -1 {
			printSymbol(symbol)
		}
		for j := 0; j < w; j++ {
			fmt.Print(" ")
		}
		printSymbol(symbol)
		w -= 2
		fmt.Println()
	}
	w += 4
	for i := 2; i < halfHeight; i++ {
		for j := i; j < halfHeight; j++ {
			fmt.Print(" ")
		}
		printSymbol(symbol)
		for j := 0; j < w; j++ {
			fmt.Print(" ")
		}
		printSymbol(symbol)
		w += 2
		fmt.Println()
	}
	printLine(width, symbol)
}

func main() {
	sandglass()
	sandglass(setSize(7))
	sandglass(setChar("O"), setSize(9))
	sandglass(setSize(10), setColor(31), setChar("S"))
}
