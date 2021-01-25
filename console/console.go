package console

import (
	"context"
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/rivo/tview"
	"github.com/vvarma/gotalk/pkg/paraU"
	"github.com/vvarma/gotalk/pkg/paraU/chat"
	"github.com/vvarma/gotalk/pkg/paraU/client"
	"github.com/vvarma/gotalk/pkg/paraU/dost"
	"io"
)

var logger = log.Logger("console")

func init() {
}

type paraUGrid struct {
	Pages        *tview.Pages
	paraUOptions client.Options
	paraU        paraU.ParaU
	ParentFlex   *tview.Flex
	loginWait    chan struct{}
	MainFlex     *tview.Flex
}

func (pg *paraUGrid) setUserName(userName string) error {
	pg.paraUOptions.Username = userName
	pu, err := paraU.New(context.Background(), pg.paraUOptions)
	if err != nil {
		return err
	}
	pg.paraU = pu
	pg.Pages.SwitchToPage("Main")
	close(pg.loginWait)
	return nil
}
func inputModal(title string, inputS string, fn func(string) error, escFn func()) tview.Primitive {
	var userName string
	errorView := tview.NewTextView().SetScrollable(false)
	loginModal := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(tview.NewTextView().
			SetScrollable(false).
			SetText(title), 0, 1, false).
		AddItem(tview.NewInputField().
			SetLabel(inputS).
			SetChangedFunc(func(text string) {
				userName = text
			}).SetDoneFunc(func(key tcell.Key) {
			switch key {
			case tcell.KeyEnter:
				err := fn(userName)
				if err != nil {
					errorView.SetText(fmt.Sprintf("Error: (%s)", err))
				}
			case tcell.KeyEscape:
				escFn()
			}
		}), 0, 1, true).
		AddItem(errorView, 0, 1, false)
	return tview.NewGrid().
		SetRows(0, 0, 0).
		SetColumns(0, 0, 0).
		AddItem(loginModal, 1, 1, 1, 1, 0, 0, true)

}

func (pg *paraUGrid) loginPage() tview.Primitive {
	logger.Info("Booting up the login page")
	return inputModal("Enter details to login", "Enter username:", func(userName string) error {
		return pg.setUserName(userName)
	}, func() {

	})
}

func (pg *paraUGrid) addFriendPage() tview.Primitive {
	return inputModal("Add a friend", "Enter friends peerId (base56):", func(peerId string) error {
		if err := pg.paraU.AddFriend(context.Background(), peerId); err != nil {
			return err
		}
		pg.Pages.SwitchToPage("Main")
		return nil
	}, func() {
		pg.Pages.SwitchToPage("Main")
	})
}
func (pg *paraUGrid) incomingFriendsPage() tview.Primitive {
	reviewList := tview.NewFlex().SetDirection(tview.FlexRow)
	reviewList.SetTitle("Incoming Friend Requests")

	for _, d := range pg.paraU.ListIncoming(context.Background()) {
		d := d
		reviewList.AddItem(
			tview.NewFlex().
				SetDirection(tview.FlexRow).
				AddItem(
					tview.NewTextView().
						SetText(fmt.Sprintf("UserName: %s", d.UserName)),
					0, 1, false).
				AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
					AddItem(tview.NewButton("Accept").SetSelectedFunc(func() {
						logger.Debug("action", "accepting request")
						logger.Error(pg.paraU.AcceptIncoming(context.Background(), d.PeerId))
					}), 0, 1, true).
					AddItem(tview.NewButton("Reject"), 0, 1, false),
					0, 1, true),
			0, 1, true)
	}
	reviewList.AddItem(tview.NewButton("Done").SetSelectedFunc(func() {
		pg.Pages.SwitchToPage("Main")
	}), 0, 1, true)
	return reviewList
}

func (pg *paraUGrid) logView() tview.Primitive {
	logView := tview.NewTextView()
	logView.SetBorder(true)
	logReader := log.NewPipeReader()
	go func() {
		_, _ = io.Copy(logView, logReader)
	}()
	return logView
}

func (pg *paraUGrid) dostView(onSelected func(peerId peer.ID)) tview.Primitive {
	friendList := tview.NewList()
	friendList.SetBorder(true)
	go func() {
		for {
			<-pg.loginWait
			if pg.paraU != nil {
				dosts := pg.paraU.List(context.Background())
				for _, d := range dosts {
					d := d
					friendList.AddItem(d.UserName, d.PeerId.String(), 0, func() {
						onSelected(d.PeerId)
					})
				}
				pg.paraU.RegisterCallback(func(ctx context.Context, event dost.Event) {
					switch event.EventType {
					case dost.Approved:
						friendList.AddItem(event.Dost.UserName, event.Dost.PeerId.String(), 0, nil)
					}
				})
				break
			}
		}
	}()
	return friendList
}

func (pg *paraUGrid) cmdView() tview.Primitive {
	cmdList := tview.NewList()
	cmdList.SetBorder(true)
	cmdList.AddItem("Add Friend", "", 0, func() {
		pg.Pages.SwitchToPage("AddFriend")
	})
	cmdList.AddItem("Review Incoming", "", 0, func() {
		pg.Pages.AddAndSwitchToPage("ReviewFriends", pg.incomingFriendsPage(), true)
	})
	return cmdList
}

func (pg *paraUGrid) infoView() tview.Primitive {
	infoBox := tview.NewTextView()
	infoBox.SetBorder(true)
	go func() {
		<-pg.loginWait
		for status := range pg.paraU.Updates() {
			logger.Info("Got a status update %s", status)
			infoBox.SetText(status.String())
		}
	}()
	return infoBox
}

type currentChat struct {
	currentDost *dost.Dost
	chatView    *tview.TextView
	dostCache   map[peer.ID]*dost.Dost
	pu          paraU.ParaU
}

func (cc *currentChat) print(ctx context.Context, msg *chat.ChatMessage) error {
	var userName string
	peerId, err := peer.Decode(msg.Meta.FromPeer)
	if err != nil {
		return err
	}
	if d, ok := cc.dostCache[peerId]; ok {
		userName = d.UserName
	} else {
		d := cc.pu.DostByPeerId(ctx, peerId)
		if d == nil {
			// todo no reference to self peer Id
			userName = "You"
		} else {
			cc.dostCache[peerId] = d
			userName = d.UserName
		}
	}
	var body string
	switch chatBody := msg.Msg.(type) {
	case *chat.ChatMessage_Text_:
		body = chatBody.Text.Body
	default:
		body = "non text message"
	}
	text := fmt.Sprintf("[%s]: %s\n", userName, body)
	_, err = cc.chatView.Write([]byte(text))
	if err != nil {
		return err
	}
	return nil
}

func (c *currentChat) OnIncoming(ctx context.Context, msg *chat.ChatMessage) {
	if c.currentDost != nil && msg.Meta.FromPeer != peer.Encode(c.currentDost.PeerId) {
		logger.Info("Got message from un focused dost", msg.Meta.FromPeer)
	}
	err := c.print(ctx, msg)
	if err != nil {
		logger.Error("error in printing incoming ", err)
	}
}

func (c *currentChat) OnOutgoin(ctx context.Context, msg *chat.ChatMessage) {
	err := c.print(ctx, msg)
	if err != nil {
		logger.Error("error in printing outgoing ", err)
	}
}

func (pg *paraUGrid) chatView() (tview.Primitive, func(id peer.ID)) {
	chatView := tview.NewTextView()
	chatView.SetBorder(true)
	chatInp := tview.NewInputField()
	chatInp.SetBorder(true)
	chatFlex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(chatView, 0, 10, false).
		AddItem(chatInp, 0, 1, true)
	cc := &currentChat{dostCache: make(map[peer.ID]*dost.Dost), chatView: chatView}
	var message string
	chatInp.SetChangedFunc(func(text string) {
		message = text
	})
	chatInp.SetDoneFunc(func(key tcell.Key) {
		if message != "" && cc.currentDost != nil {
			err := pg.paraU.Send(context.Background(), cc.currentDost, message)
			if err != nil {
				logger.Error("error in sending chat ", err)
			}
		}
		chatInp.SetText("")
	})

	onSelected := func(id peer.ID) {
		logger.Info("Selected dost to chat with id", id)
		ctx := context.Background()
		// todo when parent fn is called it goes to login page and paraU is nil ( change at root )
		if cc.pu == nil {
			cc.pu = pg.paraU
			pg.paraU.Register(context.Background(), cc)
		}
		d := pg.paraU.DostByPeerId(ctx, id)
		cc.currentDost = d
		chatMessages, err := pg.paraU.Read(ctx, d)
		if err != nil {
			logger.Error("error reading chat messages ", err)
		}
		chatView.Clear()
		for _, msg := range chatMessages {
			err := cc.print(ctx, msg)
			if err != nil {
				logger.Error("Error in printing older chats", err)
			}
		}
	}
	return chatFlex, onSelected
}

func New() *paraUGrid {
	loginWait := make(chan struct{})
	pg := &paraUGrid{loginWait: loginWait}
	logView := pg.logView()
	chatView, onDostSelected := pg.chatView()
	sideFlex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(pg.dostView(onDostSelected), 0, 2, true).
		AddItem(pg.cmdView(), 0, 1, true).
		AddItem(pg.infoView(), 0, 1, false)

	mainFlex := tview.NewFlex()
	mainFlex.AddItem(sideFlex, 0, 1, true).
		AddItem(chatView, 0, 4, false)
	pages := tview.NewPages().
		AddPage("Login", pg.loginPage(), true, true).
		AddPage("Main", mainFlex, true, false).
		AddPage("AddFriend", pg.addFriendPage(), true, false)
	pg.MainFlex = mainFlex
	pg.Pages = pages
	pg.ParentFlex = tview.NewFlex().
		AddItem(pg.Pages, 0, 10, true).
		AddItem(logView, 0, 1, false).
		SetDirection(tview.FlexRow)

	return pg
}
