package main

import (
	"fmt"

	color "github.com/logrusorgru/aurora"

	"github.com/jllopis/gmu/cmd/gmucli/action"
)

func main() {
	fmt.Printf("%s  %s\n\n", color.Bold("gmu version"), color.Cyan("0.0.1").Bold())
	action.Execute()
}
