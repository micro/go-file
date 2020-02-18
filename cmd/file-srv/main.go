package main

import (
	"context"
	"time"

	"github.com/micro/go-micro"
	mclient "github.com/micro/go-micro/client"
	"github.com/micro/go-micro/registry/memory"
	"github.com/micro/go-micro/web"
	"github.com/spf13/cobra"

	"github.com/partitio/go-file"
)

func main() {
	// service cancellation context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cmd := cobra.Command{
		Use: "file-srv [path]",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// wait chan
			wait := make(chan bool)
			r := memory.NewRegistry()
			// make service
			s := micro.NewService(
				micro.Name("go.micro.srv.file"),
				micro.Registry(r),
				micro.Context(ctx),
				micro.AfterStart(func() error {
					close(wait)
					return nil
				}),
			)

			// register file handler
			if err := file.RegisterFileHandler(s.Server(), args[0]); err != nil {
				return err
			}

			// start service
			go s.Run()

			// wait for start
			<-wait

			// new file client
			mc := mclient.NewClient(mclient.Registry(r), mclient.RequestTimeout(24 * time.Hour))
			wh := file.NewHttpHandler("go.micro.srv.file", mc)
			w := web.NewService(web.Address(":18888"), web.Context(ctx))
			w.Handle("/uploads", wh)
			w.Handle("/uploads/", wh)

			if err := w.Run(); err != nil {
				return err
			}
			return nil
		},
	}
	cmd.Execute()
}