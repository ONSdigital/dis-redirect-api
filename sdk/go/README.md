# dis-redirect-api SDK

## Overview

The dis-redirect-api contains a Go client for interacting with the API. The client contains a methods for each API endpoint
so that any Go application wanting to interact with the API can do so. Please refer to the [swagger specification](../swagger.yaml)
as the source of truth of how each endpoint works.

## Example use of the API SDK

Initialise new redirect API client

```go
package main

import (
    "context"
    redirectAPI "github.com/ONSdigital/dis-redirect-api/sdk/go"
)

func main() {
    ...
    redirectAPIClient := redirectAPI.NewClient("http://localhost:29900")
    ...
}
```

### Hello World

Use the GetHelloWorld method to send a request to say hello. This will be removed as additional functionality is added.

```go
...
    // Set authorisation header
    headers := make(map[header][]string)
    headers[Authorisation] = []string{"Bearer authorised-user"}

    resp, err := redirectAPIClient.GetHelloWorld(ctx, sdk.Options{sdk.Headers: headers})
    if err != nil {
        // handle error
    }

    /* If successful the resp value will be *HelloWorldReponse struct found in github.com/ONSdigital/dis-redirect-api/api package

    JSON equivalent:
    {
        "message": <value>,
    }
    */
...
```
