package cmd

import (
	"context"
	"os"

	"github.com/mattn/go-isatty"
)

type globalsCtxKey struct{}

type terminalCtx struct {
	isInteractive bool
}

func withGlobals(ctx context.Context, g *globalFlags) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, globalsCtxKey{}, g)
}

func getGlobals(ctx context.Context) *globalFlags {
	if ctx == nil {
		return &globalFlags{}
	}
	g, ok := ctx.Value(globalsCtxKey{}).(*globalFlags)
	if !ok || g == nil {
		return &globalFlags{}
	}
	return g
}

var isTerminal = isTerminalCheck()

func isTerminalCheck() bool {
	return os.Getenv("LOGZ_NO_INTERACTIVE") == "" &&
		(os.Getenv("TERM") != "dumb" || os.Getenv("FORCE_INTERACTIVE") != "") &&
		(isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd()))
}
