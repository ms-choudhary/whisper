# whisper
A temporary scratch pad for sharing throwaway secrets

### Why?
How do you share secrets (like new passwords) with users securely? There's a lot
of tooling available for storing secrets for services (apps), but sharing
secrets with users is still painful.

Following are requirements of a similar service:
- Users might not be technical, you shouldn't expect users to install X before
  sharing secrets with them.
- Mail are not a secure way of sending secrets. (Although predominantly used,
  and this tool aims as being replacement to this)
- We can setup a third party system (like Vault) and give users access to that
  once, and then we can use vault for all the secrets. Although this is a pretty
  good solution. We don't want to maintain a third party service.


### Idea
This project runs a webserver that anyone can use to generate new secrets. It
stores the secret temporarily on S3 (encrypted) and generates a presigned URL. 
This presigned URL can be shared with users only (on email or slack). This URL
is configured to expire within 10 minutes. A cron deletes the secrets regularly.

No tooling is required to either generate the secrets or use them. 


### Usage

```
Usage examples:
# from stdin
- echo 'somesecret' | curl --data-binary @- https://<secrets-server-url>
# from file: aws_user_secrets
- curl --data-binary @aws_user_secrets https://<secrets-server-url>
```

### Future Improvements
- S3 is not really a secure store
- secrets should be encrypted so that even the manager of this service is not
  able to see them
- secret should expire the moment it's used (although there can be time limit
  expiry as well)


