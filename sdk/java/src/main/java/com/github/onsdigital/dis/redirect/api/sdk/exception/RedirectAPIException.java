package com.github.onsdigital.dis.redirect.api.sdk.exception;

import org.apache.http.HttpStatus;

import lombok.Getter;

public class RedirectAPIException extends Exception {
    /**
     * Status code of the error.
     */
    @Getter
    private final int code;

    /**
     * Create a new instance of an RedirectAPIException.
     *
     * @param message    A string detailing the reason for the exception
     * @param statusCode The http status code that caused the API exception
     */
    public RedirectAPIException(final String message, final int statusCode) {
        super(message);
        this.code = statusCode;
    }

    /**
     * New default constructor.
     */
    public RedirectAPIException() {
        this.code = HttpStatus.SC_INTERNAL_SERVER_ERROR;
    }
}
