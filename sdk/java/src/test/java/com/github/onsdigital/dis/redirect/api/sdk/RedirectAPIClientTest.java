package com.github.onsdigital.dis.redirect.api.sdk;

import org.junit.jupiter.api.Test;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.github.onsdigital.dis.redirect.api.sdk.exception.BadRequestException;
import com.github.onsdigital.dis.redirect.api.sdk.exception.RedirectAPIException;
import com.github.onsdigital.dis.redirect.api.sdk.exception.RedirectNotFoundException;
import com.github.onsdigital.dis.redirect.api.sdk.model.Redirect;
import com.github.onsdigital.dis.redirect.api.sdk.model.Redirects;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertNotNull;
import static org.junit.jupiter.api.Assertions.assertThrows;

import java.io.UnsupportedEncodingException;
import java.net.URISyntaxException;
import java.util.ArrayList;

import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.mock;
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
     * Plain redirect ID for testing
     */
    private static final String redirectID = "/economy/old-path";

    /**
     * Count for testing
     */
    private static final String count = "3";

    /**
     * Cursor for testing
     */
    private static final String cursor = "2";

    @Test
    void testRedirectAPIInvalidURI() {

        CloseableHttpClient mockHttpClient = mock(CloseableHttpClient.class);

        // Given an invalid URI
        String invalidURI = "{{}}";

        // When a new RedirectAPIClient is created
        // Then the expected exception is thrown
        assertThrows(URISyntaxException.class,
                () -> new RedirectAPIClient(invalidURI, SERVICE_AUTH_TOKEN, mockHttpClient));
    }

    @Test
    public void testRedirectAPI_getRedirect() throws Exception {
        CloseableHttpClient mockHttpClient = mock(CloseableHttpClient.class);
        RedirectClient redirectAPIClient = getRedirectClient(mockHttpClient);

        // Given a mock redirect response from the redirect API
        CloseableHttpResponse mockHttpResponse = MockHttp.response(HttpStatus.SC_OK);
        when(mockHttpClient.execute(any(HttpRequestBase.class))).thenReturn(mockHttpResponse);

        Redirect mockRedirect = mockRedirect(mockHttpResponse);
        Redirect expectedRedirect = mockRedirect;

        // When getRedirect is called
        Redirect actualRedirect = redirectAPIClient.getRedirect(redirectID);

        assertNotNull(actualRedirect);

        // Then the response should be whats returned from the redirect API
        assertEquals(expectedRedirect.getTo(), actualRedirect.getTo());
        assertEquals(expectedRedirect.getFrom(), actualRedirect.getFrom());
    }

    @Test
    void testRedirectAPI_getRedirect_badRequest() throws Exception {
        CloseableHttpClient mockHttpClient = mock(CloseableHttpClient.class);
        RedirectClient redirectAPIClient = getRedirectClient(mockHttpClient);

        // Given a request to the redirect API that returns a 400
        CloseableHttpResponse mockHttpResponse = MockHttp.response(HttpStatus.SC_BAD_REQUEST);
        when(mockHttpClient.execute(any(HttpRequestBase.class))).thenReturn(mockHttpResponse);

        // When getRedirect is called
        // Then the expected exception is thrown
        assertThrows(BadRequestException.class,
                () -> redirectAPIClient.getRedirect(redirectID));
    }

    @Test
    void testRedirectAPI_getRedirect_redirectNotFound() throws Exception {
        CloseableHttpClient mockHttpClient = mock(CloseableHttpClient.class);
        RedirectClient redirectAPIClient = getRedirectClient(mockHttpClient);

        // Given a request to the redirect API that returns a 404
        CloseableHttpResponse mockHttpResponse = MockHttp.response(HttpStatus.SC_NOT_FOUND);
        when(mockHttpClient.execute(any(HttpRequestBase.class))).thenReturn(mockHttpResponse);

        // When getRedirect is called
        // Then the expected exception is thrown
        assertThrows(RedirectNotFoundException.class,
                () -> redirectAPIClient.getRedirect(redirectID));
    }

    @Test
    void testRedirectAPI_getRedirect_internalError() throws Exception {
        CloseableHttpClient mockHttpClient = mock(CloseableHttpClient.class);
        RedirectClient redirectAPIClient = getRedirectClient(mockHttpClient);

        // Given a request to the redirect API that returns a 500
        CloseableHttpResponse mockHttpResponse = MockHttp.response(HttpStatus.SC_INTERNAL_SERVER_ERROR);
        when(mockHttpClient.execute(any(HttpRequestBase.class))).thenReturn(mockHttpResponse);

        // When getRedirect is called
        // Then the expected exception is thrown
        assertThrows(RedirectAPIException.class,
                () -> redirectAPIClient.getRedirect(redirectID));
    }

    private Redirect mockRedirect(CloseableHttpResponse mockHttpResponse)
            throws JsonProcessingException, UnsupportedEncodingException {
        Redirect responseBody = new Redirect("/economy/old-path", "/economy/new-path");

        MockHttp.responseBody(mockHttpResponse, responseBody);

        return responseBody;
    }

    private Redirects mockRedirects(CloseableHttpResponse mockHttpResponse)
            throws JsonProcessingException, UnsupportedEncodingException {
        ArrayList<Redirect> redirectList = new ArrayList<>();
        Redirects responseBody = new Redirects(3, redirectList, "2", "0", 3);

        MockHttp.responseBody(mockHttpResponse, responseBody);

        return responseBody;
    }

    private RedirectClient getRedirectClient(
            final CloseableHttpClient mockHttpClient)
            throws URISyntaxException {
        return new RedirectAPIClient(
                REDIRECT_API_URL, SERVICE_AUTH_TOKEN, mockHttpClient);
    }

    @Test
    public void testRedirectAPI_getRedirects() throws Exception {
        CloseableHttpClient mockHttpClient = mock(CloseableHttpClient.class);
        RedirectClient redirectAPIClient = getRedirectClient(mockHttpClient);

        // Given a mock redirects response from the redirect API
        CloseableHttpResponse mockHttpResponse = MockHttp.response(HttpStatus.SC_OK);
        when(mockHttpClient.execute(any(HttpRequestBase.class))).thenReturn(mockHttpResponse);

        Redirects mockRedirects = mockRedirects(mockHttpResponse);
        Redirects expectedRedirects = mockRedirects;

        // When getRedirects is called
        Redirects observedRedirects = redirectAPIClient.getRedirects(count, cursor);

        assertNotNull(observedRedirects);

        // Then the response should be what's returned from the redirect API
        assertEquals(expectedRedirects.getCount(), observedRedirects.getCount());
        assertEquals(expectedRedirects.getRedirectList(), observedRedirects.getRedirectList());
        assertEquals(expectedRedirects.getCursor(), observedRedirects.getCursor());
        assertEquals(expectedRedirects.getNextCursor(), observedRedirects.getNextCursor());
        assertEquals(expectedRedirects.getTotalCount(), observedRedirects.getTotalCount());
    }

    @Test
    void testRedirectAPI_getRedirects_badRequest() throws Exception {
        CloseableHttpClient mockHttpClient = mock(CloseableHttpClient.class);
        RedirectClient redirectAPIClient = getRedirectClient(mockHttpClient);

        // Given a request to the redirect API that returns a 400
        CloseableHttpResponse mockHttpResponse = MockHttp.response(HttpStatus.SC_BAD_REQUEST);
        when(mockHttpClient.execute(any(HttpRequestBase.class))).thenReturn(mockHttpResponse);

        // When getRedirects is called
        // Then the expected exception is thrown
        assertThrows(BadRequestException.class,
                () -> redirectAPIClient.getRedirects(count, cursor));
    }

    @Test
    void testRedirectAPI_getRedirects_internalError() throws Exception {
        CloseableHttpClient mockHttpClient = mock(CloseableHttpClient.class);
        RedirectClient redirectAPIClient = getRedirectClient(mockHttpClient);

        // Given a request to the redirect API that returns a 500
        CloseableHttpResponse mockHttpResponse = MockHttp.response(HttpStatus.SC_INTERNAL_SERVER_ERROR);
        when(mockHttpClient.execute(any(HttpRequestBase.class))).thenReturn(mockHttpResponse);

        // When getRedirects is called
        // Then the expected exception is thrown
        assertThrows(RedirectAPIException.class,
                () -> redirectAPIClient.getRedirects(count, cursor));
    }
}
