package main

import (
	"context"
	"flag"

	"github.com/klintcheng/kim/logger"
	"github.com/klintcheng/kim/services/gateway"
	"github.com/klintcheng/kim/services/router"
	"github.com/klintcheng/kim/services/server"
	"github.com/klintcheng/kim/services/service"
	"github.com/spf13/cobra"
)

const version = "v1"

/**
业务层：链路层+控制层
代码在 /services/
*/
func main() {
	flag.Parse()

	root := &cobra.Command{
		Use:     "kim",
		Version: version,
		Short:   "King IM Cloud",
	}
	ctx := context.Background()

	// go run main.go gateway
	// go run main.go server
	// go run main.go royal
	root.AddCommand(gateway.NewServerStartCmd(ctx, version))
	root.AddCommand(server.NewServerStartCmd(ctx, version))
	root.AddCommand(service.NewServerStartCmd(ctx, version))
	root.AddCommand(router.NewServerStartCmd(ctx, version))

	if err := root.Execute(); err != nil {
		logger.WithError(err).Fatal("Could not run command")
	}
}
