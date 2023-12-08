package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/mattermost/mattermost/server/public/model"
)

type MattermostClient struct {
	client *model.Client4
}

func Login(url, token string) (*MattermostClient, error) {
	c := model.NewAPIv4Client(url)
	c.AuthToken = token
	c.AuthType = model.HeaderBearer
	return &MattermostClient{
		client: c,
	}, nil
}

func (p *Plugin) ensureConnected() error {
	if p.client == nil || p.UserID == "" {
		return p.initializeMattermostClient()
	}
	return nil
}

func (p *Plugin) initializeMattermostClient() error {
	p.API.LogDebug("Initializing client")
	configuration := p.getConfiguration()
	if configuration == nil || configuration.AccessToken == "" {
		p.API.LogInfo("Not configured yet, skipping login.")
		return nil
	}
	config := p.API.GetConfig()
	p.siteURL = *config.ServiceSettings.SiteURL

	p.API.LogDebug(fmt.Sprintf("Login to %s using token %s", p.siteURL, configuration.AccessToken))
	c, err := Login(p.siteURL, configuration.AccessToken)
	if err != nil {
		return nil
	}
	id, err := c.getUserID()
	if err != nil {
		return err
	}
	p.API.LogDebug(fmt.Sprintf("Got user id %s", id))

	p.client = c
	p.UserID = id

	return nil
}

func (c *MattermostClient) getUserID() (string, error) {
	user, resp, err := c.client.GetMe(context.TODO(), "")
	if err != nil {
		return "", fmt.Errorf("failed to get user id: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get user id: %d", resp.StatusCode)
	}
	return user.Id, nil
}

// RegisterNewEmoji send a request for creating emoji
func (c *MattermostClient) RegisterNewEmoji(b []byte, name, userID string) error {
	_, resp, err := c.client.CreateEmoji(context.TODO(), &model.Emoji{
		CreatorId: userID,
		Name:      name,
	}, b, name)

	if err != nil || resp.StatusCode != http.StatusOK {
		return fmt.Errorf(err.Error())
	}
	return nil
}
