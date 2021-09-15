package main

import "fmt"

type (
	Symbol []string
	Mod    func(s *Symbol)
)

func printSymbol(s Symbol) {
	if len(s) > 1 {
		fmt.Print(s[1] + s[0] + s[2])
	} else {
		fmt.Print(s[0])
	}
}

func printLine(size int, s Symbol) {
	for j := 0; j < size; j++ {
		printSymbol(s)
	}
	fmt.Println()
}

func setColor(color int) Mod {
	return func(s *Symbol) {
		*s = append(*s, fmt.Sprintf("\u001B[%vm", color), "\u001B[0m")
	}
}

func setChar(char string) Mod {
	return func(s *Symbol) {
		(*s)[0] = char
	}
}

func sandglass(size int, mods ...Mod) {
	symbol := Symbol{"X"}
	for _, mod := range mods {
		mod(&symbol)
	}

	printLine(size, symbol)
	draw(size-2, 0, symbol)
	printLine(size, symbol)
}

func draw(width int, spaces int, s Symbol) {
	fmt.Print(" ")
	stars(width, spaces, width, s)
	if width > 2 {
		draw(width-2, spaces+1, s)
		fmt.Print(" ")
		stars(width, spaces, width, s)
	}
}

func stars(n int, spaces int, width int, s Symbol) {
	switch {
	case spaces > 0:
		fmt.Print(" ")
		stars(n, spaces-1, width, s)
	case n > 0:
		if n == 1 || n == width {
			printSymbol(s)
		} else {
			fmt.Print(" ")
		}
		stars(n-1, spaces, width, s)
	default:
		fmt.Println()
	}
}

func main() {
	sandglass(11)
	sandglass(6, setChar("O"))
	sandglass(13, setColor(31), setChar("S"))
}
