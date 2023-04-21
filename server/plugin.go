package main

import (
	"fmt"
	"net/http"
	"strings"
	"sync"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/mattermost/mattermost-server/v6/plugin"
)

// Plugin implements the interface expected by the Mattermost server to communicate between the server and plugin processes.
type Plugin struct {
	plugin.MattermostPlugin
	client        *MattermostClient
	pluginClient  *pluginapi.Client
	siteURL       string
	UserID        string
	command       model.Command
	ephemeralPost *model.Post

	// configurationLock synchronizes access to the configuration.
	configurationLock sync.RWMutex

	// configuration is the active plugin configuration. Consult getConfiguration and
	// setConfiguration for usage.
	configuration *configuration
}

func (p *Plugin) OnActivate() error {
	p.API.LogInfo("Before New Client")
	if p.pluginClient == nil {
		p.pluginClient = pluginapi.NewClient(p.API, p.Driver)
	}
	p.API.LogInfo("After New Client")
	p.pluginClient.Log.Info("Activating...")

	p.pluginClient.Log.Info("Registering slash command...")
	p.command = model.Command{
		Trigger:          "partyparrotssync",
		AutoComplete:     true,
		AutoCompleteDesc: `Sync Party Parrots emojis`,
	}

	if err := p.pluginClient.SlashCommand.Register(&p.command); err != nil {
		p.pluginClient.Log.Error(err.Error())
		return err
	}

	if err := p.ensureConnected(); err != nil {
		p.pluginClient.Log.Error(err.Error())
	}
	p.pluginClient.Log.Info("Done.")
	return nil
}

func (p *Plugin) OnDeactivate() error {
	p.pluginClient.Log.Info("Deactivating...")

	p.pluginClient.Log.Info("Unregistering slash command...")
	if err := p.pluginClient.SlashCommand.Unregister("", p.command.Trigger); err != nil {
		p.pluginClient.Log.Error(err.Error())
		return err
	}
	p.pluginClient.Log.Info("Done.")
	return nil
}

// ExecuteCommand handle commands that are created by this plugin
func (p *Plugin) ExecuteCommand(_ *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	p.pluginClient.Log.Info("Slash command received.")
	p.SendEphemeralPost(args.ChannelId, args.UserId, args.RootId, "Starting Party Parrots sync...")
	if err := p.ensureConnected(); err != nil {
		p.pluginClient.Log.Error(err.Error())
	}
	for _, parrotType := range parrotTypes {
		p.pluginClient.Log.Info(fmt.Sprintf("Fetching gif list for %s.", parrotType))
		list, err := fetchParrotList(parrotType)
		if err != nil {
			// Try the next parrot type
			p.pluginClient.Log.Error(fmt.Sprintf("Could not fetch gif list for type %s", parrotType))
			continue
		}

		for i, parrot := range list {
			// Show progress to user
			p.UpdateEphemeralPost(fmt.Sprintf("Processing emoji %d of %d from type %s", i+1, len(list), parrotType))
			p.CreateEmoji(parrot, parrotType)
		}
	}
	p.UpdateEphemeralPost("Party Parrots sync done! Enjoy :partyparrot:")
	return &model.CommandResponse{}, nil
}

func (p *Plugin) CreateEmoji(parrot Parrot, parrotType string) {
	// Validate if we already have an emoji matching this gif
	if p.EmojiExists(parrot.name) {
		// We already have it, skip
		p.pluginClient.Log.Info(fmt.Sprintf("Emoji :%s: already exists. Skipping.", parrot.name))
		return
	}
	// Fetch the gif data from GitHub
	p.pluginClient.Log.Info(fmt.Sprintf("Fetching gif for %s.", parrot.name))
	if fetchParrotGif(&parrot, parrotType) != nil {
		p.pluginClient.Log.Error(fmt.Sprintf("Failed to fetch %s", parrot.name))
		return
	}
	appErr := p.client.RegisterNewEmoji(parrot.gif, parrot.name, p.UserID)
	if appErr != nil && strings.Contains(appErr.Error(), "Name conflicts with existing system emoji name") {
		parrot.name += "2"
		if !p.EmojiExists(parrot.name) {
			p.pluginClient.Log.Info(fmt.Sprintf("Emoji :%s: already exists. Skipping.", parrot.name))
			return
		}
		appErr := p.client.RegisterNewEmoji(parrot.gif, parrot.name, p.UserID)
		if appErr != nil {
			p.pluginClient.Log.Error(fmt.Sprintf("Could not create emoji %s: %s", parrot.name, appErr.Error()))
		}
	}
}

func (p *Plugin) EmojiExists(name string) bool {
	emoji, _ := p.pluginClient.Emoji.GetByName(name)
	return emoji != nil
}

// SendEphemeralPost sends an ephemeral post to a user as the bot account
func (p *Plugin) SendEphemeralPost(channelID, userID, rootID, message string) {
	p.ephemeralPost = &model.Post{
		ChannelId: channelID,
		UserId:    p.UserID,
		RootId:    rootID,
		Message:   message,
	}
	p.pluginClient.Post.SendEphemeralPost(userID, p.ephemeralPost)
}

func (p *Plugin) UpdateEphemeralPost(message string) {
	p.ephemeralPost.Message = message
	p.pluginClient.Post.UpdateEphemeralPost(p.UserID, p.ephemeralPost)
}

func (p *Plugin) ServeHTTP(_ *plugin.Context, w http.ResponseWriter, _ *http.Request) {
	fmt.Fprint(w, "Hello, world!")
}

// See https://developers.mattermost.com/extend/plugins/server/reference/
