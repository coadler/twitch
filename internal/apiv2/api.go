package api

import (
	"context"

	"github.com/coadler/twitch/pb"
)

var _ pb.TwitchServer = &twitch{}

type twitch struct{}

func (t *twitch) GetChannels(context.Context, *pb.GetChannelsRequest) (*pb.GetChannelsResponse, error) {
	panic("not implemented")
}

func (t *twitch) NewWebhook(context.Context, *pb.NewWebhookRequest) (*pb.NewWebhookResponse, error) {
	panic("not implemented")
}

func (t *twitch) DeleteWebook(context.Context, *pb.DeleteWebhookRequest) (*pb.DeleteWebhookResponse, error) {
	panic("not implemented")
}
