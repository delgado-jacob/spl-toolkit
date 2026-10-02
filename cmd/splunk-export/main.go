// spl-toolkit-export is a bounded live acquisition companion to the offline toolkit.
package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/delgado-jacob/spl-toolkit/internal/splunkexport"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	os.Exit(splunkexport.Run(ctx, os.Args[1:], os.Stdout, os.Stderr))
}
