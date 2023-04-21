# Mattermost Party Parrots Sync Plugin

This plugin allows you to sync all Party Parrots from [Cult of the Party Parrot](https://cultofthepartyparrot.com/) as emojis in your Mattermost instance.

## Getting Started

### Prebuilt

Get the latest `.tar.gz` archive from the releases page of this repo and upload it to your Mattermost instance through `System Console -> Plugin Management`.

### Build it yourself

1. Ensure you have `make`, `go` and `golangci`.
1. Clone the repo.
1. Build the plugin:

    ```sh
    make
    ```

This will produce a single plugin file (with support for multiple architectures) for upload to your Mattermost server:

`dist/org.polycancer.mattermost-partyparrots-sync-0.1.0.tar.gz`

### Configuration

For now, this plugin requires a Personal Access Token to allow access to the CreateEmoji API.
Hopefully, this can eventually be replaced by self-managed bot account.

### Usage

Once the plugin is installed, enabled and configured, you can run the following slash command:
`/partyparrotssync`

This will sync all parrots, flags and guests.
Subsequent runs will skip emojis that are already imported.

## Development

To avoid having to manually install your plugin, build and deploy your plugin using one of the following options. In order for the below options to work, you must first enable plugin uploads via your config.json or API and restart Mattermost.

```json
    "PluginSettings" : {
        ...
        "EnableUploads" : true
    }
```

### Deploying with Local Mode

If your Mattermost server is running locally, you can enable [local mode](https://docs.mattermost.com/administration/mmctl-cli-tool.html#local-mode) to streamline deploying your plugin. Edit your server configuration as follows:

```json
{
    "ServiceSettings": {
        ...
        "EnableLocalMode": true,
        "LocalModeSocketLocation": "/var/tmp/mattermost_local.socket"
    },
}
```

and then deploy your plugin:
```
make deploy
```

You may also customize the Unix socket path:
```
export MM_LOCALSOCKETPATH=/var/tmp/alternate_local.socket
make deploy
```

If developing a plugin with a webapp, watch for changes and deploy those automatically:
```
export MM_SERVICESETTINGS_SITEURL=http://localhost:8065
export MM_ADMIN_TOKEN=j44acwd8obn78cdcx7koid4jkr
make watch
```

### Deploying with credentials

Alternatively, you can authenticate with the server's API with credentials:
```
export MM_SERVICESETTINGS_SITEURL=http://localhost:8065
export MM_ADMIN_USERNAME=admin
export MM_ADMIN_PASSWORD=password
make deploy
```

or with a [personal access token](https://docs.mattermost.com/developer/personal-access-tokens.html):
```
export MM_SERVICESETTINGS_SITEURL=http://localhost:8065
export MM_ADMIN_TOKEN=j44acwd8obn78cdcx7koid4jkr
make deploy
```
