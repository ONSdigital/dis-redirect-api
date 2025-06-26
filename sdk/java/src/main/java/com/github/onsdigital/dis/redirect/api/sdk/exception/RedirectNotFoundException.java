package com.github.onsdigital.dis.redirect.api.sdk.exception;

import org.apache.http.HttpStatus;

import lombok.Getter;

public class RedirectNotFoundException extends Exception {

    /**
     * Status code of the error.
     */
    @Getter
    private final int code;

    /**
     *
     * @param message    A string detailing the reason for the exception
     * @param statusCode The http status code that caused the API exception
     */
    public RedirectNotFoundException(final String message,
            final int statusCode) {
        super(message);
        this.code = statusCode;
    }

    /**
     * New default constructor.
     */
    public RedirectNotFoundException() {
        this.code = HttpStatus.SC_NOT_FOUND;
    }
}
