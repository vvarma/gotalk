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
	"github.com/vvarma/gotalk/pkg/paraU/chat"
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
					c.Register(context.Context, &simpleCallback{bufio.NewWriter(os.Stdout)})
					// todo unregister
					if context.Args().Len() != 1 {
						return errors.New("username needed")
					}
					userName := context.Args().Get(0)
					d, err := c.DostByUserName(context.Context, userName)
					if err != nil {
						return err
					}
					err = c.Start(context.Context, d)
					if err != nil {
						return err
					}
					r := bufio.NewReader(os.Stdin)

					for {
						line, err := r.ReadString('\n')
						if err != nil {
							return err
						}
						err = c.Send(context.Context, d, line)
						if err != nil {
							return err
						}
					}

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

type simpleCallback struct {
	w *bufio.Writer
}

func (s *simpleCallback) OnIncoming(ctx context.Context, msg *chat.ChatMessage) {
	var text string
	switch textMsg := msg.Msg.(type) {
	case *chat.ChatMessage_Text_:
		text = textMsg.Text.Body
	}
	_, err := s.w.WriteString(fmt.Sprintf("[%s]:%s", msg.Meta.FromPeer, text))
	if err != nil {
		logger.Error("error writing callback", err)
	}
}

func (s *simpleCallback) OnOutgoin(ctx context.Context, msg *chat.ChatMessage) {
}
