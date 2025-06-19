package com.github.onsdigital.dis.redirect.api.sdk.model;

import com.fasterxml.jackson.annotation.JsonInclude;
import com.fasterxml.jackson.annotation.JsonIgnoreProperties;

@JsonIgnoreProperties(ignoreUnknown = true)
public class RedirectResponse {
    /**
     * The redirect id.
     */
    @JsonInclude(JsonInclude.Include.NON_NULL)
    private String id;

    /**
     * The next redirect.
     */
    @JsonInclude(JsonInclude.Include.NON_NULL)
    private Redirect next;

    /**
     * Get redirect id.
     * @return id
     */
    public String getId() {
        return id;
    }

    /**
     * Get next redirect.
     * @return next
     */
    public Redirect getNext() {
        return next;
    }

    /**
     * Set the redirect id.
     * @param redirectID
     */
    public void setId(final String redirectID) {
        this.id = redirectID;
    }

    /**
     * Set the next redirect.
     * @param nextRedirect
     */
    public void setNext(final Redirect nextRedirect) {
        this.next = nextRedirect;
    }
}
