package config

import (
	"flag"
	"fmt"
	"testing"
)

func TestFlag(t *testing.T) {
	args := []string{"-known1", "-known2", "-known3", "-known4", "-unknown1", "-unknown2", "-unknown3"}
	flagset := flag.NewFlagSet("", flag.ContinueOnError)
	flagset.Bool("known1", false, "")
	flagset.Bool("known2", false, "")
	flagset.Bool("known3", false, "")
	known4 := flagset.String("known4", "", "")

	err := flagset.Parse(args)
	if err != nil {
		fmt.Println(err)
	}
	remArgs := flagset.Args()

	fmt.Println(remArgs)
	fmt.Println(*known4)
}
