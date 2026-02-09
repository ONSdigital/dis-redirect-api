package com.github.onsdigital.dis.redirect.api.sdk;

import java.io.IOException;
import java.net.URI;
import java.net.URISyntaxException;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.github.onsdigital.dis.redirect.api.sdk.model.Redirect;
import com.github.onsdigital.dis.redirect.api.sdk.model.Redirects;
import com.github.onsdigital.dis.redirect.api.sdk.exception.BadRequestException;
import com.github.onsdigital.dis.redirect.api.sdk.exception.RedirectAPIException;
import com.github.onsdigital.dis.redirect.api.sdk.exception.RedirectNotFoundException;

import org.apache.hc.core5.net.URIBuilder;
import org.apache.hc.client5.http.impl.classic.CloseableHttpClient;
import org.apache.hc.client5.http.impl.classic.HttpClients;
import org.apache.commons.lang3.StringUtils;
import org.apache.hc.core5.http.HttpEntity;
import org.apache.hc.core5.http.HttpStatus;
import org.apache.hc.core5.util.Args;
import org.apache.hc.core5.http.ParseException;

import org.apache.hc.core5.http.io.entity.StringEntity;
import org.apache.hc.core5.http.ContentType;

import org.apache.hc.core5.http.io.entity.EntityUtils;

import org.apache.hc.client5.http.impl.classic.CloseableHttpResponse;

import org.apache.hc.client5.http.classic.methods.HttpPut;
import org.apache.hc.client5.http.classic.methods.HttpGet;
import org.apache.hc.client5.http.classic.methods.HttpDelete;
import org.apache.hc.client5.http.classic.methods.HttpUriRequestBase;
import org.apache.hc.client5.http.classic.methods.HttpUriRequest;
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
            RedirectAPIException, ParseException {

        validateRedirectID(redirectID);

        String encodedID = Redirect.encodeRedirectID(redirectID);

        String path = "/v1/redirects/" + encodedID;
        URI uri = redirectAPIUri.resolve(path);

        HttpGet req = new HttpGet(uri);
        req.addHeader(SERVICE_TOKEN_HEADER_NAME, authToken);

        try (CloseableHttpResponse resp = executeRequest(req)) {
            validateResponseCode(req, resp);
            return parseResponseBody(resp, Redirect.class);
        }
    }

    /**
     * Upserts a redirect by sending a PUT request to /redirects/{id}.
     *
     * @param payload  the redirect payload with 'from' and 'to' fields
     * @throws IOException            if request fails
     * @throws RedirectAPIException   if non-2xx response returned
     */
    @Override
    public void putRedirect(final Redirect payload)
            throws IOException, RedirectAPIException {

        if (payload.getFrom() == null || payload.getFrom().isEmpty()) {
            throw new IllegalArgumentException(
                "'from' must not be null or empty"
            );
        }

        URI requestUri = redirectAPIUri.resolve("/v1/redirects/"
                + payload.getId());
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
                int statusCode = response.getCode();

            if (statusCode != HttpStatus.SC_CREATED
                    && statusCode != HttpStatus.SC_OK) {
                throw new RedirectAPIException(
                        formatErrResponse(put, response,
                        HttpStatus.SC_CREATED),
                        statusCode);
            }
        }
    }

    /**
     * Deletes a redirect by sending a DELETE request to /redirects/{id}.
     * The {@code fromPath} is base64 URL-encoded internally.
     *
     * @param fromPath the raw unencoded redirect source path
     * @throws IOException            if the request fails
     * @throws RedirectAPIException   if a non-204 response is returned
     */
    @Override
    public void deleteRedirect(final String fromPath)
            throws IOException, RedirectAPIException {

        if (fromPath == null || fromPath.isEmpty()) {
            throw new
            IllegalArgumentException("'fromPath' must not be null or empty");
        }

        String base64Id = Redirect.encodeRedirectID(fromPath);

        URI requestUri = redirectAPIUri.resolve("/v1/redirects/" + base64Id);
        HttpDelete delete = new HttpDelete(requestUri);

        delete.addHeader(SERVICE_TOKEN_HEADER_NAME, "Bearer " + authToken);

        try (CloseableHttpResponse response = executeRequest(delete)) {
            int statusCode = response.getCode();

            if (statusCode != HttpStatus.SC_NO_CONTENT) {
                throw new RedirectAPIException(
                        formatErrResponse(delete, response,
                        HttpStatus.SC_NO_CONTENT),
                        statusCode
                );
            }
        }
    }

    private void validateRedirectID(final String redirectID) {
        Args.check(StringUtils.isNotBlank(redirectID),
                "a redirect id must be provided.");
    }

    private void validateResponseCode(final HttpUriRequestBase httpRequest,
            final CloseableHttpResponse response)
            throws IOException, BadRequestException, RedirectNotFoundException,
            RedirectAPIException {
        int statusCode = response.getCode();

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

    private <T> T parseResponseBody(final CloseableHttpResponse response,
            final Class<T> type) throws IOException, ParseException  {
        HttpEntity entity = response.getEntity();
        String responseString = EntityUtils.toString(entity);
        return JSON.readValue(responseString, type);
    }

     private String formatErrResponse(final HttpUriRequestBase httpRequest,
            final CloseableHttpResponse response,
            final int expectedStatusCode) {
        int responseCode = response.getCode();

        try {
            String requestURI = httpRequest.getUri().toString();
            return String.format(
                "the redirect api returned a %s response for %s (expected %s)",
                            responseCode,
                            requestURI,
                            expectedStatusCode);
        } catch (URISyntaxException e) {
            return String.format(
                "the redirect api returned a %s response for %s (expected %s)",
                responseCode,
                httpRequest.getRequestUri(),
                expectedStatusCode);
        }
    }

    private CloseableHttpResponse executeRequest(final HttpUriRequest req)
            throws IOException {
        // TODO: remove reliance on CloseableHttpClient.execute
        // as it is deprecated in HttpClient 5.4.0 - instead
        // use HttpClient.execute with a ResponseHandler
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

    /**
     * Get a redirects object containing the requested number of
     * redirect objects in a list.
     *
     * @param count - the number of redirect objects requested
     * @param cursor - the location, in the store, to start counting from
     * @return throws an exception to indicate an error
     * @throws IOException
     * @throws BadRequestException
     * @throws RedirectAPIException
     * @throws RedirectNotFoundException
     * @throws URISyntaxException
     * @throws ParseException
     */
    @Override
    public Redirects getRedirects(final String count, final String cursor)
            throws IOException, BadRequestException, ParseException,
            RedirectAPIException, RedirectNotFoundException,
            URISyntaxException {
        String path = "/v1/redirects";
        URIBuilder builder = new URIBuilder(redirectAPIUri.resolve(path));
        if (StringUtils.isNotBlank(count)) {
            builder.setParameter("count", count);
        }

        if (StringUtils.isNotBlank(cursor)) {
             builder.setParameter("cursor", cursor);
        }
        URI uri = builder.build();

        HttpGet req = new HttpGet(uri);
        req.addHeader(SERVICE_TOKEN_HEADER_NAME, authToken);

        try (CloseableHttpResponse resp = executeRequest(req)) {
            validateResponseCode(req, resp);
            Redirects response = parseResponseBody(resp, Redirects.class);
            return response;
        }
    }
}
