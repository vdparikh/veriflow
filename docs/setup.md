# Setup Veriflow
Running Veriflow in your environment is straightword.


For now, Veriflow only supports Slack and AWS. More opitions coming soon

At a high level, below are the steps to get Veriflow running
1. Deploy App to AWS get your cloudfront URL
2. Setup your Slack App
3. Setup Auth Provider
4. Update configuration
5. Deploy Config

 
## Deploy App to AWS get your cloudfront URL
Before you deploy, please create an s3 bucket and export that to environment like `export VERIFLOW_BUCKET=veriflow`

- Run `make build`  to build your code and create a zip package for deployment
- Run `make package` to created a serverless deployment package
- Run `make deploy` to deploy this to your AWS environment.

Once the deploy completes successfully, you will see an output variable with a cloudfront URL. Copy that as you will need it for setting up your slack app

> **Want to run Veriflow app locally for testing?**<br/>
> Follow instructions in this [guide](/docs/local.md) to get it working locally and come back here for the remaining setup.

## Setup your Slack App
Following instructions in this [guide](/docs/slack/README.md) to get your `botToken` and `appToken` 

## Setup Auth Provider
Following instructions to setup your auth provider and get your `client_id` and `client_secret`. Update config for those 2 attribute along with `issuer` url

- For Google follow instruction at https://support.google.com/cloud/answer/6158849 

## Update configuration
Update the config.yaml file with all the setup details.

- `base_url`  - This is your cloudfront URL
- `communication.services.slack.app_token` - Slack Bot Token
- `communication.services.slack.bot_token` - Slack App Token
- `auth.client_id` - This is your OIDC client ID
- `auth.client_secret` - This is your OIDC client secret
- `auth.issuer` - If you are using a different provider than Google, than update the issuer URL. Make sure that is a valid OIDC issuer (ex: https://accounts.google.com/.well-known/openid-configuration)


## Deploy Config
> You can ignore this step if you are running it locally
> 
`make config` - This will upload the config file to the VERIFLOW_BUCKET S3 bucket. 











