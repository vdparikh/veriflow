![Alt text](/docs/images/logo.png)

# Veriflow

[![Veriflow Overview](https://img.youtube.com/vi/SOp-C50vlns/0.jpg)](https://www.youtube.com/watch?v=SOp-C50vlns)


## What is Veriflow?
Veriflow is an automated and ad-hoc verification system designed to address the challenges of safeguarding from vishing and smishing attacks. By implementing Veriflow, you can establish a zero-trust verification process for out-of-band communications.

Simply put - Veriflow helps securing Human-2-Human auth for out of band communications


At a high level, here is the flow of Veriflow

![Alt text](/docs/images/flow.png)

### How does Veriflow work?
1. **Caller initiates the Communication with the Callee**  
This could be a voice call, video call, or any other form of direct communication. With the advancement of deep fakes, you cannot trust the number or the voice of the caller.

2. **Callee triggers   `/veriflow verify <caller> <optional message>` in a communication channel (e.g., Slack)**  

![Alt text](/docs/images/recipient.png)

3. **The Veriflow app understands the ask, generates a unique authentication link on your auth provider (Google, OKTA, PING etc.), and sends a message to the Caller for verification**  

![Alt text](/docs/images/requestor.png)

4. **The caller clicks on the Authenticate link and follows standard 2FA login to verify**  

5. **On successful Authentication, Veriflow sends a message to both Caller and Callee with a message**  

![Alt text](/docs/images/recipient_2.png)


### Benefits
- Provides frictionless flow for identity verification.
- Harnesses the power of existing infrastructure and authentication, to enhance identity verification and security.

### Problems to Solve / Future Use Cases
- Users initiate verification when they contact the help desk or require access to sensitive services. In this scenario, the user does not have access to company communication.


### Ready to run Veriflow?
Head on the docs folder to get started on installation Veriflow.
[Getting Started](/docs/setup.md)
