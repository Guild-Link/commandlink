package webhook

import (
	"context"
	"errors"
	"fmt"
	"slices"

	d "github.com/bwmarrin/discordgo"
)

func New(session *d.Session, name string) *Client {
	return &Client{session: session, name: name}
}

func (c *Client) Execute(ctx context.Context, channelID string, params *d.WebhookParams) (*d.Message, error) {
	w, err := c.Get(ctx, channelID)
	if err != nil {
		return nil, err
	}

	message, err := c.session.WebhookExecute(w.ID, w.Token, true, params, d.WithContext(ctx))
	if !isDeleted(err) {
		return message, err
	}

	c.webhooks.Delete(channelID)
	if w, err = c.Get(ctx, channelID); err != nil {
		return nil, err
	}

	return c.session.WebhookExecute(w.ID, w.Token, true, params, d.WithContext(ctx))
}

func (c *Client) Edit(ctx context.Context, channelID, messageID string, edit *d.WebhookEdit) (*d.Message, error) {
	w, err := c.Get(ctx, channelID)
	if err != nil {
		return nil, err
	}

	return c.session.WebhookMessageEdit(w.ID, w.Token, messageID, edit, d.WithContext(ctx))
}

func (c *Client) Get(ctx context.Context, channelID string) (*d.Webhook, error) {
	if w, ok := c.webhooks.Load(channelID); ok {
		return w.(*d.Webhook), nil
	}

	webhooks, err := c.session.ChannelWebhooks(channelID, d.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("get Discord channel webhooks: %w", err)
	}

	if i := slices.IndexFunc(webhooks, func(w *d.Webhook) bool { return w.Name == c.name && w.Token != "" }); i >= 0 {
		c.webhooks.Store(channelID, webhooks[i])
		return webhooks[i], nil
	}

	w, err := c.session.WebhookCreate(channelID, c.name, "", d.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("create Discord channel webhook: %w", err)
	}

	c.webhooks.Store(channelID, w)
	return w, nil
}

func (c *Client) Delete(ctx context.Context, channelID string) error {
	c.webhooks.Delete(channelID)
	webhooks, err := c.session.ChannelWebhooks(channelID, d.WithContext(ctx))
	if err != nil {
		return err
	}

	var errs []error
	for _, w := range webhooks {
		if w.Name == c.name {
			errs = append(errs, c.session.WebhookDelete(w.ID, d.WithContext(ctx)))
		}
	}

	return errors.Join(errs...)
}

func isDeleted(err error) bool {
	rest, ok := err.(*d.RESTError)
	return ok && rest.Message != nil && rest.Message.Code == d.ErrCodeUnknownWebhook
}
