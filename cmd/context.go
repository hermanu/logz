package cmd

import "context"

type globalsCtxKey struct{}

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
