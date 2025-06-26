package com.github.onsdigital.dis.redirect.api.sdk;

import java.io.IOException;
import java.net.URI;
import java.net.URISyntaxException;
import java.util.Base64;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.github.onsdigital.dis.redirect.api.sdk.model.Redirect;
import com.github.onsdigital.dis.redirect.api.sdk.exception.BadRequestException;
import com.github.onsdigital.dis.redirect.api.sdk.exception.RedirectAPIException;
import com.github.onsdigital.dis.redirect.api.sdk.exception.RedirectNotFoundException;

import org.apache.http.impl.client.CloseableHttpClient;
import org.apache.http.impl.client.HttpClients;

import org.apache.http.HttpEntity;
import org.apache.http.HttpStatus;
import org.apache.http.util.Args;
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
            final String serviceAuthToken, final CloseableHttpClient httpClient)
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
    public RedirectAPIClient(final String redirectAPIURL,
            final String serviceAuthToken) throws URISyntaxException {
        this(redirectAPIURL, serviceAuthToken, createDefaultHttpClient());
    }

    private static CloseableHttpClient createDefaultHttpClient() {
        return HttpClients.createDefault();
    }

    /**
     * Get the redirect for the given redirect ID.
     *
     * @param redirectID
     * @return throws an exception to indicate an error
     * @throws IOException
     * @throws BadRequestException
     * @throws RedirectNotFoundException
     * @throws RedirectAPIException
     */
    @Override
    public Redirect getRedirect(final String redirectID)
            throws IOException, BadRequestException, RedirectNotFoundException,
            RedirectAPIException {

        validateRedirectID(redirectID);

        String path = "/redirects/" + redirectID;
        URI uri = redirectAPIUri.resolve(path);

        HttpGet req = new HttpGet(uri);
        req.addHeader(SERVICE_TOKEN_HEADER_NAME, authToken);

        try (CloseableHttpResponse resp = executeRequest(req)) {
            validateResponseCode(req, resp);
            Redirect response = parseResponseBody(resp, Redirect.class);
            response.setFrom(decodeBase64(redirectID));
            return response;
        }
    }

    private void validateRedirectID(final String redirectID) {
        Args.check(isNotEmpty(redirectID), "a redirect id must be provided.");
        Args.check(isBase64(redirectID), "redirect id must be base 64");
    }

    private void validateResponseCode(final HttpRequestBase httpRequest,
            final CloseableHttpResponse response)
            throws IOException, BadRequestException, RedirectNotFoundException,
            RedirectAPIException {
        int statusCode = response.getStatusLine().getStatusCode();

        switch (statusCode) {
        case HttpStatus.SC_OK:
            return;
        case HttpStatus.SC_BAD_REQUEST:
            throw new BadRequestException(formatErrResponse(httpRequest,
                    response, HttpStatus.SC_BAD_REQUEST), statusCode);
        case HttpStatus.SC_NOT_FOUND:
            throw new RedirectNotFoundException(formatErrResponse(httpRequest,
                    response, HttpStatus.SC_NOT_FOUND), statusCode);
        default:
            throw new RedirectAPIException(formatErrResponse(httpRequest,
                    response, HttpStatus.SC_INTERNAL_SERVER_ERROR), statusCode);
        }
    }

    private static boolean isNotEmpty(final String str) {
        return str != null && str.length() > 0;
    }

    private static boolean isBase64(final String str) {
        return Base64.getDecoder().decode(str) != null;
    }

    private static String decodeBase64(final String str) {
        return new String(Base64.getDecoder().decode(str));
    }

    private <T> T parseResponseBody(final CloseableHttpResponse response,
            final Class<T> type) throws IOException {
        HttpEntity entity = response.getEntity();
        String responseString = EntityUtils.toString(entity);
        return JSON.readValue(responseString, type);
    }

    private String formatErrResponse(final HttpRequestBase httpRequest,
            final CloseableHttpResponse response, final int expectedStatus) {
        return String.format(
                "the redirect api returned a %s response for %s (expected %s)",
                response.getStatusLine().getStatusCode(), httpRequest.getURI(),
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
