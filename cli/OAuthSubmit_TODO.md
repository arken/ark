## A list of all the things I'll have to do to get this working

### 1. Make the JSON structs for step 1 in `types`

### 2. Use bare http requests to set the user up with our app.
- If we have the PAT, just fetch it from the config. Otherwise...
- Send the initial POST:
```http
POST https://github.com/login/device/code?client_id=<client-id>&scope=repo
Accept: application/json
```
- Get the device code out of ^
- Display the device code and tell the user to enter it at 
  https://github.com/login/device
- Every `interval` seconds (from the last request), poll this request until 
  either a timeout is exceeded, or the response's error field no longer says 
  `authorization_pending`:
```http
POST https://github.com/login/oauth/access_token?client_id=<client-id>&device_code=<dev-code>&grant-type=urn:ietf:params:oauth:grant-type:device_code
Accept: application/json
```
- If the poll eventually yields a token, create a new Github client:
```go
tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: patFromPoll})
client := github.NewClient(oauth2.NewClient(ctx, tokenSource))
```
- Ask the user if they want to stay logged in. If so, save the PAT to the config.

### 3. Fork the repository. Should still work the same, I think.
- Get all user info from the API, not the user

### 4. Use `client.Repositories.GetContents()` to get the contents of ***the parent directory of*** `remotePath`
- Fetch the desired file's SHA by iterating over the files in the directory
- Possible alternative is the Trees API? Not sure.
- If the file does exist, ask the user if they want to update it or overwrite it
- Delete the file if they said overwrite using `client.Repositories.DeleteFile()`

### 5. With the desired file's SHA, use `client.Repositories.UpdateFile()` or `client.Repositories.CreateFile()`
- Create/update the file within the fork

### 6. Make the pull request. Should also work the same.

