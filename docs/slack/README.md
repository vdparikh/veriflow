# Setup Slack App for Veriflow

### Getting your manifest file ready
Edit the mainfest.yaml file to update the `features.slash_commands.url` and `settings.request_url`

### Head on to https://api.slack.com/apps?new_app=1 to create a new app
![Alt text](/docs/slack/images/2.png)


### Choose to create an app from an app manifest.
![Alt text](/docs/slack/images/3.png)

### Pick a development workspace and click Next.
![Alt text](/docs/slack/images/4.png)

### Paste manifest configuration in the input field provided and click Next.
Edit the mainfest.yaml file to update the `features.slash_commands.url` and `settings.request_url`
![Alt text](/docs/slack/images/5.png)

### Install to the workspace
![Alt text](/docs/slack/images/6.png)


### Generate App Token
On the Basic Information page, scroll to the section which says **App Level Token** and click on generate Tokens and Scope. Generate new token with `connection:write` scope and copy the generate token
![Alt text](/docs/slack/images/8.png)
![Alt text](/docs/slack/images/9.png)

### Get your BOT Token
Navigate to oAuth and Permissions to copy the bot token
![Alt text](/docs/slack/images/7.png)

### Update the config file