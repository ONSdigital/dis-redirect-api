package com.github.onsdigital.dis.redirect.api.sdk;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import org.apache.http.StatusLine;
import org.apache.http.client.methods.CloseableHttpResponse;
import org.apache.http.entity.StringEntity;

import java.io.UnsupportedEncodingException;

import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.when;

class MockHttp {

    protected MockHttp() {
        // prevents calls from subclass
        throw new UnsupportedOperationException();
      }

    /**
     * JSON mapper.
     */
    private static final ObjectMapper JSON = new ObjectMapper();

    public static CloseableHttpResponse response(final int httpStatus) {

        CloseableHttpResponse mockHttpResponse = mock(
                CloseableHttpResponse.class);

        StatusLine mockResponseStatus = mock(StatusLine.class);
        when(mockResponseStatus.getStatusCode()).thenReturn(httpStatus);
        when(mockHttpResponse.getStatusLine()).thenReturn(mockResponseStatus);

        return mockHttpResponse;
    }

    public static void responseBody(
            final CloseableHttpResponse mockHttpResponse,
            final Object responseBody)
            throws JsonProcessingException, UnsupportedEncodingException {
        String responseJSON = JSON.writeValueAsString(responseBody);
        when(mockHttpResponse.getEntity()).thenReturn(
                new StringEntity(responseJSON));
    }
}
