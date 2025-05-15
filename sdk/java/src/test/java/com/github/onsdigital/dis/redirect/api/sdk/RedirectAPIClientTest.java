package com.github.onsdigital.dis.redirect.api.client;

import org.junit.jupiter.api.Test;
import org.mockito.ArgumentCaptor;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.github.onsdigital.dis.redirect.api.client.model.HelloWorld;
import com.github.onsdigital.dis.redirect.api.client.exception.RedirectAPIException;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertNotNull;
import static org.junit.jupiter.api.Assertions.assertNull;
import static org.junit.jupiter.api.Assertions.assertThrows;

import java.io.IOException;
import java.io.UnsupportedEncodingException;
import java.net.URISyntaxException;

import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

import org.apache.http.HttpStatus;
import org.apache.http.client.methods.CloseableHttpResponse;
import org.apache.http.client.methods.HttpRequestBase;
import org.apache.http.impl.client.CloseableHttpClient;

class RedirectAPIClientTest {
    /**
     * Auth token for testing.
     */
    private static final String SERVICE_AUTH_TOKEN = "67856";

    /**
     * Redirect API URL for testing.
     */
    private static final String REDIRECT_API_URL = "http://redirect-api:1234";

    /**
     * Auth header for testing.
     */
    private static final String SERVICE_TOKEN_HEADER_NAME = "Authorization";

    @Test
    void testRedirectAPIInvalidURI() {

        CloseableHttpClient mockHttpClient = mock(CloseableHttpClient.class);

        // Given an invalid URI
        String invalidURI = "{{}}";

        // When a new RedirectAPIClient is created
        // Then the expected exception is thrown
        assertThrows(URISyntaxException.class,
                () -> new RedirectAPIClient(
                        invalidURI, SERVICE_AUTH_TOKEN, mockHttpClient));
    }

    @Test
    void testRedirectAPIGetHelloWorld() throws Exception {
        CloseableHttpClient mockHttpClient = mock(CloseableHttpClient.class);
        RedirectClient redirectAPIClient = getRedirectClient(mockHttpClient);

        // Given a mock helloworld response from the redirect API
        CloseableHttpResponse mockHttpResponse = MockHttp.response(
                HttpStatus.SC_OK);

        when(mockHttpClient.execute(
                any(HttpRequestBase.class))).thenReturn(mockHttpResponse);

        HelloWorld mockHelloWorldResponse = mockHelloWorldResponse(
                mockHttpResponse);

        // When getHelloWorld is called
        HelloWorld actualHelloWorld = redirectAPIClient.getHelloWorld();

        assertNotNull(actualHelloWorld);

        HttpRequestBase httpRequest = captureHttpRequest(mockHttpClient);

        // Then no query params are in the URI
        assertNull(httpRequest.getURI().getQuery());

        // Then the request should contain the service token header
        String actualServiceToken = httpRequest.getFirstHeader(
                SERVICE_TOKEN_HEADER_NAME).getValue();
        assertEquals(SERVICE_AUTH_TOKEN, actualServiceToken);

        // Then the response should be whats returned from the redirect API
        assertEquals(mockHelloWorldResponse.getMessage(),
                actualHelloWorld.getMessage());
    }

    @Test
    void testRedirectAPIGetHellowWorldInternalError() throws Exception {
        CloseableHttpClient mockHttpClient = mock(CloseableHttpClient.class);
        RedirectClient client = getRedirectClient(mockHttpClient);

        // Given a request to the redirect API that returns a 500
        CloseableHttpResponse mockHttpResponse = MockHttp.response(
                HttpStatus.SC_INTERNAL_SERVER_ERROR);
        when(mockHttpClient.execute(
                any(HttpRequestBase.class))).thenReturn(mockHttpResponse);

        // When getHelloWorld is called
        // Then the expected exception is thrown
        assertThrows(RedirectAPIException.class,
                () -> client.getHelloWorld());
    }

    private RedirectClient getRedirectClient(
            final CloseableHttpClient mockHttpClient)
            throws URISyntaxException {
        return new RedirectAPIClient(
                REDIRECT_API_URL, SERVICE_AUTH_TOKEN, mockHttpClient);
    }

    private HttpRequestBase captureHttpRequest(
            final CloseableHttpClient mockHttpClient)
            throws IOException {
        ArgumentCaptor<HttpRequestBase> requestCaptor = ArgumentCaptor.forClass(
                HttpRequestBase.class);
        verify(mockHttpClient).execute(requestCaptor.capture());
        return requestCaptor.getValue();
    }

    private HelloWorld mockHelloWorldResponse(
            final CloseableHttpResponse mockHttpResponse)
            throws JsonProcessingException, UnsupportedEncodingException {
        HelloWorld responseBody = new HelloWorld();
        responseBody.setMessage("hello");
        MockHttp.responseBody(mockHttpResponse, responseBody);

        return responseBody;
    }

}
