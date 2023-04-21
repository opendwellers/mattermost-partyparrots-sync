package main

import (
	"fmt"
	"net/http"

	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/pkg/errors"
)

type MattermostClient struct {
	client *model.Client4
	token  *model.UserAccessToken
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
	p.pluginClient.Log.Debug("Initializing client")
	configuration := p.getConfiguration()
	if configuration == nil || configuration.AccessToken == "" {
		p.pluginClient.Log.Info("Not configured yet, skipping login.")
		return nil
	}
	config := p.pluginClient.Configuration.GetConfig()
	p.siteURL = *config.ServiceSettings.SiteURL
	bot := &model.Bot{
		Username:    "partyparrotssync",
		DisplayName: "Party Parrots Bot",
		Description: "Created by the Party Parrots Sync plugin.",
	}
	userID, err := p.pluginClient.Bot.EnsureBot(bot)
	if err != nil {
		return errors.Wrap(err, "Failed to ensure bot")
	}
	p.UserID = userID
	p.pluginClient.Log.Debug(fmt.Sprintf("Created bot with id %s", p.UserID))
	p.client.token, err = p.pluginClient.User.CreateAccessToken(p.UserID, "")
	if err != nil {
		return errors.Wrap(err, "Failed to create bot access token")
	}
	p.pluginClient.Log.Debug(fmt.Sprintf("Login to %s", p.siteURL))
	c, err := Login(p.siteURL, p.client.token.Token)
	if err != nil {
		return nil
	}
	id, err := c.getUserID()
	if err != nil {
		return err
	}
	p.pluginClient.Log.Debug(fmt.Sprintf("Got user id %s", id))

	p.client = c
	p.UserID = id

	return nil
}

func (c *MattermostClient) getUserID() (string, error) {
	user, resp, err := c.client.GetMe("")
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
	_, resp, err := c.client.CreateEmoji(&model.Emoji{
		CreatorId: userID,
		Name:      name,
	}, b, name)

	if err != nil || resp.StatusCode != http.StatusOK {
		return fmt.Errorf(err.Error())
	}
	return nil
}
