package com.github.onsdigital.dis.redirect.api.sdk;

import org.junit.jupiter.api.Test;
import org.mockito.ArgumentCaptor;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.github.onsdigital.dis.redirect.api.sdk.exception.BadRequestException;
import com.github.onsdigital.dis.redirect.api.sdk.exception.RedirectAPIException;
import com.github.onsdigital.dis.redirect.api.sdk.exception.RedirectNotFoundException;
import com.github.onsdigital.dis.redirect.api.sdk.model.Redirect;
import com.github.onsdigital.dis.redirect.api.sdk.model.Redirects;

import static org.junit.jupiter.api.Assertions.assertTrue;
import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertNotNull;
import static org.junit.jupiter.api.Assertions.assertThrows;

import java.io.UnsupportedEncodingException;
import java.io.IOException;
import java.net.URISyntaxException;
import java.util.ArrayList;
import java.nio.charset.StandardCharsets;
import java.util.Base64;

import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

import org.apache.http.HttpStatus;
import org.apache.http.StatusLine;
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

    @Test
    void putRedirectSuccessfullySendsPutRequest() throws Exception {
        CloseableHttpClient mockHttpClient = mock(CloseableHttpClient.class);
        CloseableHttpResponse mockResponse = mock(CloseableHttpResponse.class);
        StatusLine mockStatus = mock(StatusLine.class);
        when(mockStatus.getStatusCode()).thenReturn(HttpStatus.SC_CREATED);
        when(mockResponse.getStatusLine()).thenReturn(mockStatus);
        when(mockHttpClient.execute(any())).thenReturn(mockResponse);

        RedirectClient client = getRedirectClient(mockHttpClient);
        Redirect redirect = new Redirect("/from-path", "/to-path");
        client.putRedirect(redirect);

        HttpRequestBase request = captureHttpRequest(mockHttpClient);
        assertEquals("PUT", request.getMethod());

        String expectedId = Base64.getUrlEncoder()
                .withoutPadding()
                .encodeToString("/from-path".getBytes(StandardCharsets.UTF_8));

        assertTrue(request.getURI().toString().endsWith("/redirects/" + expectedId));
        assertEquals("Bearer " + SERVICE_AUTH_TOKEN, request.getFirstHeader("Authorization").getValue());
    }

    @Test
    void putRedirectReturnsErrorOnUnexpectedStatusCode() throws Exception {
        CloseableHttpClient mockHttpClient = mock(CloseableHttpClient.class);
        CloseableHttpResponse mockResponse = mock(CloseableHttpResponse.class);
        StatusLine mockStatus = mock(StatusLine.class);
        when(mockStatus.getStatusCode()).thenReturn(HttpStatus.SC_BAD_GATEWAY);
        when(mockResponse.getStatusLine()).thenReturn(mockStatus);
        when(mockHttpClient.execute(any())).thenReturn(mockResponse);

        RedirectClient client = getRedirectClient(mockHttpClient);
        Redirect redirect = new Redirect("/bad", "/to");

        RedirectAPIException ex = assertThrows(RedirectAPIException.class, () -> client.putRedirect(redirect));
        assertEquals(HttpStatus.SC_BAD_GATEWAY, ex.getCode());
    }

    @Test
    void putRedirectThrowsIOExceptionOnFailure() throws Exception {
        CloseableHttpClient mockHttpClient = mock(CloseableHttpClient.class);
        when(mockHttpClient.execute(any())).thenThrow(new IOException("connection fail"));

        RedirectClient client = getRedirectClient(mockHttpClient);
        Redirect redirect = new Redirect("/fail", "/to");

        assertThrows(IOException.class, () -> client.putRedirect(redirect));
    }

    @Test
    void putRedirectThrowsOnNullFromField() throws Exception {
        CloseableHttpClient mockHttpClient = mock(CloseableHttpClient.class);
        RedirectClient client = getRedirectClient(mockHttpClient);

        Redirect invalid = new Redirect(null, "/to");
        assertThrows(IllegalArgumentException.class, () -> client.putRedirect(invalid));
    }

    @Test
    void testDeleteRedirectSuccess() throws Exception {
        CloseableHttpClient mockHttpClient = mock(CloseableHttpClient.class);
        CloseableHttpResponse mockResponse = mock(CloseableHttpResponse.class);
        StatusLine mockStatusLine = mock(StatusLine.class);

        when(mockStatusLine.getStatusCode()).thenReturn(HttpStatus.SC_NO_CONTENT);
        when(mockResponse.getStatusLine()).thenReturn(mockStatusLine);
        when(mockHttpClient.execute(any(HttpRequestBase.class))).thenReturn(mockResponse);

        RedirectClient client = getRedirectClient(mockHttpClient);

        client.deleteRedirect("/from"); // raw path, not base64

        HttpRequestBase request = captureHttpRequest(mockHttpClient);
        assertEquals("DELETE", request.getMethod());

        // Ensure URL ends with base64("/from")
        String expectedId = Base64.getUrlEncoder()
                .withoutPadding()
                .encodeToString("/from".getBytes(StandardCharsets.UTF_8));

        assertTrue(request.getURI().toString().endsWith("/redirects/" + expectedId));
        assertNotNull(request.getFirstHeader("Authorization"));
    }

    @Test
    void testDeleteRedirectReturns404() throws Exception {
        CloseableHttpClient mockHttpClient = mock(CloseableHttpClient.class);
        CloseableHttpResponse mockResponse = mock(CloseableHttpResponse.class);
        StatusLine mockStatusLine = mock(StatusLine.class);

        when(mockStatusLine.getStatusCode()).thenReturn(HttpStatus.SC_NOT_FOUND);
        when(mockResponse.getStatusLine()).thenReturn(mockStatusLine);
        when(mockHttpClient.execute(any(HttpRequestBase.class))).thenReturn(mockResponse);

        RedirectClient client = getRedirectClient(mockHttpClient);

        RedirectAPIException exception = assertThrows(
                        RedirectAPIException.class,
                        () -> client.deleteRedirect("/from")
                );

        assertTrue(exception.getMessage().contains("404"));
    }

    @Test
    void testDeleteRedirectServerError() throws Exception {
        CloseableHttpClient mockHttpClient = mock(CloseableHttpClient.class);
        CloseableHttpResponse mockResponse = mock(CloseableHttpResponse.class);
        StatusLine mockStatusLine = mock(StatusLine.class);

        when(mockStatusLine.getStatusCode()).thenReturn(HttpStatus.SC_INTERNAL_SERVER_ERROR);
        when(mockResponse.getStatusLine()).thenReturn(mockStatusLine);
        when(mockHttpClient.execute(any(HttpRequestBase.class))).thenReturn(mockResponse);

        RedirectClient client = getRedirectClient(mockHttpClient);

        RedirectAPIException exception = assertThrows(
                RedirectAPIException.class,
                () -> client.deleteRedirect("/from")
        );

        assertEquals(HttpStatus.SC_INTERNAL_SERVER_ERROR, exception.getCode());
    }

    @Test
    void testDeleteRedirectIOException() throws Exception {
        CloseableHttpClient mockHttpClient = mock(CloseableHttpClient.class);
        when(mockHttpClient.execute(any(HttpRequestBase.class)))
                .thenThrow(new IOException("Simulated network failure"));

        RedirectClient client = getRedirectClient(mockHttpClient);

        assertThrows(IOException.class, () ->
                client.deleteRedirect("/from"));
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

    private HttpRequestBase captureHttpRequest(
            final CloseableHttpClient mockHttpClient)
            throws IOException {
        ArgumentCaptor<HttpRequestBase> requestCaptor = ArgumentCaptor.forClass(
                HttpRequestBase.class);
        verify(mockHttpClient).execute(requestCaptor.capture());
        return requestCaptor.getValue();
    }

}
