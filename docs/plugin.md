# Plugins

Proctor using GO plugin, for official documentation about it please read [here](https://golang.org/pkg/plugin/).

Proctor decide to use plugin for some feature to make it easily integrate into different component.

### Notification Plugin

Plugin can be made for publishing notification to external channel such as slack.
In order to use notification plugin you should compile the plugin and fill these environment variables:
  1. PROCTOR_NOTIFICATION_PLUGIN_BINARY

     Fill this variable with path to compiled plugin separated by comma since it's possible to use multiple notification channel.

  2. PROCTOR_NOTIFICATION_PLUGIN_EXPORTED
  
     Fill this variable with name of variable exported from respective plugin binary separated by comma

### Authentication and Authorization Plugin

Authentication and authorization process is delegated to plugin so you can use any existing user management system with proctor.
In order to use auth plugin you should compile the plugin and fill these environment variables:
  1. PROCTOR_AUTH_PLUGIN_BINARY

     Fill this variable with path to compiled plugin, only single plugin is allowed

  2. PROCTOR_AUTH_PLUGIN_EXPORTED
  
     Fill this variable with name of variable exported from plugin binary
  
  3. PROCTOR_AUTH_ENABLED
  
     Fill this with `true` to activate auth using plugin

### Provided plugin

#### Gate Auth Plugin

In order to use gate auth plugin, you need a running [Gate](https://github.com/gate-sso/gate) server.
Authenticated user mean an user that registered to Gate server.
Authorized user mean an user need to be member of at least one group from groups list specified on authorized_groups metadata for procs.

Compile gate auth plugin by running `make plugin.auth` and fill `PROCTOR_AUTH_PLUGIN_BINARY` with generated `auth.so` in `./_output/bin/plugin/auth.so`.



#### Slack Notification Plugin

Proctor will send notification to slack when some event happen, see below for a list of events and it's content.
Create a [slack app](https://api.slack.com/incoming-webhooks) then fill `SLACK_PLUGIN_URL` environment variable with incoming webhook url, it should look like `https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX`.

Compile slack notification plugin by running `make plugin.slack` and fill `PROCTOR_NOTIFICATION_PLUGIN_BINARY` with generated `slack.so` in `./_output/bin/plugin/slack.so`.
