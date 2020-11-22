package subcmd

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/urfave/cli/v2"
	"github.com/vvarma/gotalk/pkg/paraU"
	"github.com/vvarma/gotalk/pkg/paraU/dost"
	"os"
	"strings"
)

var (
	logger = log.Logger("subcmd")
	app    = cli.App{
		Name: "paraU",
		Commands: []*cli.Command{
			{
				Name: "add",
				Action: func(context *cli.Context) error {
					c := paraU.GetFromContext(context.Context)
					peerId := context.Args().Get(0)
					return c.AddFriend(context.Context, peerId)
				},
			},
			{
				Name: "list",
				Action: func(context *cli.Context) error {
					c := paraU.GetFromContext(context.Context)
					for _, d := range c.List(context.Context) {
						fmt.Printf("Peer %s Username %s\n", peer.Encode(d.PeerId), d.UserName)
					}
					return nil
				},
			},
			{
				Name: "chat",
				Action: func(context *cli.Context) error {
					c := paraU.GetFromContext(context.Context)
					if context.Args().Len() != 1 {
						return errors.New("username needed")
					}
					userName := context.Args().Get(0)
					d, err := c.DostByUserName(context.Context, userName)
					if err != nil {
						return err
					}
					rw := bufio.NewReadWriter(bufio.NewReader(os.Stdin), bufio.NewWriter(os.Stdout))
					return c.Start(context.Context, d, rw)
				},
			},
			{
				Name: "review",
				Action: func(context *cli.Context) error {
					c := paraU.GetFromContext(context.Context)
					reviewFn := func(d *dost.Dost) bool {
						fmt.Printf("Approve Peer: %s (y/n)", d.PeerId.Pretty())
						r := bufio.NewReader(os.Stdin)
						for {
							line, err := r.ReadString('\n')
							if err != nil {
								logger.Error("error reading input")
								continue
							}
							if line == "y\n" {
								return true
							} else if line == "n\n" {
								break
							}
						}
						return false
					}
					c.Review(context.Context, reviewFn)
					return nil
				},
			},
		},
	}
)

func Execute(ctx context.Context, cmd string) error {
	args := []string{"hack"}
	args = append(args, strings.Split(cmd, " ")...)
	return app.RunContext(ctx, args)
}
