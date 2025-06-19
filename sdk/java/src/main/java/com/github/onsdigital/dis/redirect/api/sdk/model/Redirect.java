package com.github.onsdigital.dis.redirect.api.sdk.model;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;
import com.fasterxml.jackson.annotation.JsonInclude;

/**
 * The model of a redirect as provided by the redirect API.
 */
@JsonIgnoreProperties(ignoreUnknown = true)
public class Redirect {
    /**
     * The redirect id.
     */
    @JsonInclude(JsonInclude.Include.NON_EMPTY)
    private String id;

    /**
     * Get redirect id.
     * @return id
     */
    public String getId() {
        return id;
    }
}
