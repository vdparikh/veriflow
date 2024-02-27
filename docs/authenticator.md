## Enable Authenticator
Well in event your enterprise does not support 2FA, Veriflow allows enabling authenticator which allows users to use Google or MS authenticator as 2FA
Currently Veriflow maintains a DB table with users to store the TOTP secret and verification flow. The generated image is also stored on the disk. 

### Configure
Users will need to call `/veriflow configure` and the system will generate a QR code which they can scan from the authenticator app

### Verification Process
After successful authentication, Veriflow will prompt the user to provide the OTP code before verification request is approved/completed. 

