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
import org.apache.commons.lang3.StringUtils;
import org.apache.http.HttpEntity;
import org.apache.http.HttpStatus;
import org.apache.http.util.Args;

import org.apache.http.entity.StringEntity;
import org.apache.http.entity.ContentType;

import org.apache.http.util.EntityUtils;

import org.apache.http.client.methods.CloseableHttpResponse;

import org.apache.http.client.methods.HttpPut;
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

        String encodedID = encodeToBase64(redirectID);

        String path = "/v1/redirects/" + encodedID;
        URI uri = redirectAPIUri.resolve(path);

        HttpGet req = new HttpGet(uri);
        req.addHeader(SERVICE_TOKEN_HEADER_NAME, authToken);

        try (CloseableHttpResponse resp = executeRequest(req)) {
            validateResponseCode(req, resp);
            Redirect response = parseResponseBody(resp, Redirect.class);
            response.setFrom(redirectID);
            return response;
        }
    }

    /**
         * Upserts a redirect by sending a PUT request to /redirects/{id}.
         *
         * @param base64Id the base64 URL-encoded redirect key
         * @param payload  the redirect payload with 'from' and 'to' fields
         * @throws IOException            if request fails
         * @throws RedirectAPIException   if non-2xx response returned
         */
        @Override
        public void putRedirect(final String base64Id,
        final Redirect payload)
                throws IOException, RedirectAPIException {

            URI requestUri = redirectAPIUri.resolve("/v1/redirects/" + base64Id);
            HttpPut put = new HttpPut(requestUri);

            // Add Authorization header
            put.addHeader(SERVICE_TOKEN_HEADER_NAME, "Bearer " + authToken);
            put.addHeader("Content-Type", "application/json");

            // Serialize payload
            String jsonPayload = JSON.writeValueAsString(payload);
            put.setEntity(new StringEntity(
                    jsonPayload,
                    ContentType.APPLICATION_JSON));

            try (CloseableHttpResponse response = executeRequest(put)) {
                int statusCode = response.getStatusLine().getStatusCode();

                if (statusCode != HttpStatus.SC_CREATED
                        && statusCode != HttpStatus.SC_OK) {
                    throw new RedirectAPIException(
                            formatErrResponse(put, response, HttpStatus.SC_CREATED),
                            statusCode);
                }
            }
        }

    private void validateRedirectID(final String redirectID) {
        Args.check(StringUtils.isNotBlank(redirectID),
                "a redirect id must be provided.");
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

    private static String encodeToBase64(final String str) {
        return new String(Base64.getEncoder().encodeToString(str.getBytes()));
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
