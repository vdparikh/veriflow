## Understanding config.yml

### Base URL
`base_url: "https://b860-73-158-140-75.ngrok-free.app"`

Base URL is the primary URL for where the application is hosted. 

### provider
User provider to fetch details about user like Name, Email etc.

### communication
All different communication methods where you want the verification message to be sent out

### email
Enaul setup for where you want user to also get verification messages

### auth
OIDC setup for where you want user to get redirected when they click "authenticate"
The callback URL for OIDC

### authenticator
If you like to setup an additional 2FA 

### report
Where to redirect users where they click "report"

### Messages
Customize message sent to the users


