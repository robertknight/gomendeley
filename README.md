Mendeley API Go Demo
====================

This is a very simple demo showing how to authenticate
with the Mendeley API and perform requests against
it using Go.

Visit [dev.mendeley.com](http://dev.mendeley.com/) for API documentation and details.

Authentication is handled using [goauth2](https://code.google.com/p/goauth2).
See the Mendeley [authentication reference](http://dev.mendeley.com/html/authentication.html) for more details on aspects of OAuth2 specific to Mendeley.

Using this demo:

1. Go to [dev.mendeley.com](http://dev.mendeley.com) and register a new app. Set the 'Redirect URI' to http://localhost:8080 and remember to record the client secret somewhere.

2. Create a JSON file client_config.json with the client ID and secret:

```
{
	"ClientId" : "<your app's client ID>",
	"ClientSecret" : "<your app's client secret>"
}
```

3. Run `go build` to build the app and `./gomendeley` to run it

4. Navigate to [http://localhost:8080](http://localhost:8080)
