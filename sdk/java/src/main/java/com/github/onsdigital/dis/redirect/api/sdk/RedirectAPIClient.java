package com.github.onsdigital.dis.redirect.api.client;

import java.io.IOException;
import java.net.URI;
import java.net.URISyntaxException;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.github.onsdigital.dis.redirect.api.client.model.HelloWorld;
import com.github.onsdigital.dis.redirect.api.client.exception.RedirectAPIException;

import org.apache.http.impl.client.CloseableHttpClient;
import org.apache.http.impl.client.HttpClients;

import org.apache.http.HttpEntity;
import org.apache.http.HttpStatus;

import org.apache.http.util.EntityUtils;

import org.apache.http.client.methods.CloseableHttpResponse;

import org.apache.http.client.methods.HttpGet;
import org.apache.http.client.methods.HttpRequestBase;
import org.apache.http.client.methods.HttpUriRequest;

public class RedirectAPIClient implements RedirectClient {

    /**
     * uri for the redirectAPI.
     */
    private final URI redirectAPIUri;

    /**
     * Auth token to be used on all requests.
     */
    private final String authToken;

    /**
     * HTTP client to be used on all requests.
     */
    private final CloseableHttpClient client;

    /**
     * Header name to apply authToken to.
     */
    private static final String SERVICE_TOKEN_HEADER_NAME = "Authorization";

    /**
     * For mapping json to objects.
     */
    private static final ObjectMapper JSON = new ObjectMapper();

    /**
     * Create a new instance of RedirectAPIClient.
     *
     * @param redirectAPIURL   The URL of the redirect API
     * @param serviceAuthToken The authentication token for the redirect API
     * @param httpClient       The HTTP client to use internally
     */
    public RedirectAPIClient(final String redirectAPIURL,
            final String serviceAuthToken,
            final CloseableHttpClient httpClient)
            throws URISyntaxException {

        this.redirectAPIUri = new URI(redirectAPIURL);
        this.client = httpClient;
        this.authToken = serviceAuthToken;
    }

    /**
     * Create a new instance of RedirectAPIClient with a default Http client.
     *
     * @param redirectAPIURL   The URL of the redirect API
     * @param serviceAuthToken The authentication token for the redirect API
     * @throws URISyntaxException
     */
    public RedirectAPIClient(
            final String redirectAPIURL,
            final String serviceAuthToken)
            throws URISyntaxException {
        this(redirectAPIURL, serviceAuthToken, createDefaultHttpClient());
    }

    private static CloseableHttpClient createDefaultHttpClient() {
        return HttpClients.createDefault();
    }

    /**
     * Get Hello World.
     * TODO: Remove this.
     *
     * @return An {@link HelloWorld} object containing a message
     * @throws IOException
     * @throws RedirectAPIException
     */
    @Override
    public HelloWorld getHelloWorld() throws IOException, RedirectAPIException {

        StringBuilder pathBuilder = new StringBuilder("/hello");

        URI uri = redirectAPIUri.resolve(pathBuilder.toString());

        HttpGet req = new HttpGet(uri);
        req.addHeader(SERVICE_TOKEN_HEADER_NAME, authToken);

        try (CloseableHttpResponse resp = executeRequest(req)) {
            int statusCode = resp.getStatusLine().getStatusCode();

            switch (statusCode) {
                case HttpStatus.SC_OK:
                    return parseResponseBody(resp, HelloWorld.class);
                default:
                    throw new RedirectAPIException(formatErrResponse(
                            req, resp, HttpStatus.SC_OK), statusCode);
            }
        }
    }

    private <T> T parseResponseBody(
            final CloseableHttpResponse response, final Class<T> type)
            throws IOException {
        HttpEntity entity = response.getEntity();
        String responseString = EntityUtils.toString(entity);
        return JSON.readValue(responseString, type);
    }

    private String formatErrResponse(
            final HttpRequestBase httpRequest,
            final CloseableHttpResponse response,
            final int expectedStatus) {
        return String.format(
                "the redirect api returned a %s response for %s (expected %s)",
                response.getStatusLine().getStatusCode(),
                httpRequest.getURI(),
                expectedStatus);
    }

    private CloseableHttpResponse executeRequest(final HttpUriRequest req)
            throws IOException {
        return client.execute(req);
    }

    /**
     * Close the http client used by the APIClient.
     *
     * @throws IOException
     */
    @Override
    public void close() throws IOException {
        client.close();
    }
}
