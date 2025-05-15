# dis-redirect-api Java SDK

## Overview

The dis-redirect-api contains a Java client for interacting with the API. The client contains a methods for each API endpoint
so that any Java application wanting to interact with the API can do so. Please refer to the [swagger specification](../swagger.yaml)
as the source of truth of how each endpoint works.

## Example use of the API Java SDK

### Add to your pom.xml

```xml
    <dependency>
      <groupId>com.github.onsdigital</groupId>
      <artifactId>dis-redirect-api-sdk-java</artifactId>
      <version>${redirectSDK.version}</version>
    </dependency>
```

### Initialise a client

```java
package com.github.onsdigital.dis.my.application;

import com.github.onsdigital.dis.redirect.api.sdk.RedirectClient;
import com.github.onsdigital.dis.redirect.api.sdk.model.HelloWorld;
import com.github.onsdigital.dis.redirect.api.sdk.exception.RedirectAPIException;

public class MyApplicationClass() {

    private String REDIRECT_API_URL = "http://localhost:29900";
    private String SERVICE_AUTH_TOKEN = "xyz1234";

    public static void main(String[] args) {
        RedirectClient client = new RedirectAPIClient(
                REDIRECT_API_URL, SERVICE_AUTH_TOKEN);
    }
}
```

### Hello World

Use the GetHelloWorld method to send a request to say hello. This will be removed as additional functionality is added.

```java
package com.github.onsdigital.dis.my.application;

import com.github.onsdigital.dis.redirect.api.sdk.RedirectClient;
import com.github.onsdigital.dis.redirect.api.sdk.model.HelloWorld;
import com.github.onsdigital.dis.redirect.api.sdk.exception.RedirectAPIException;

...
    try {
        HelloWorld helloWorldResponse = redirectAPIClient.getHelloWorld();
    } catch (RedirectAPIException ex){
        System.exit(1);
    }
...
```
