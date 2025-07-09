package com.github.onsdigital.dis.redirect.api.sdk;

import java.io.Closeable;
import java.io.IOException;

import com.github.onsdigital.dis.redirect.api.sdk.exception.BadRequestException;
import com.github.onsdigital.dis.redirect.api.sdk.exception.RedirectAPIException;
import com.github.onsdigital.dis.redirect.api.sdk.exception.RedirectNotFoundException;
import com.github.onsdigital.dis.redirect.api.sdk.model.Redirect;
import com.github.onsdigital.dis.redirect.api.sdk.model.Redirects;

public interface RedirectClient extends Closeable {
    /**
     * @param redirectID
     * @return throws an exception to indicate an error
     * @throws IOException
     * @throws BadRequestException
     * @throws RedirectNotFoundException
     * @throws RedirectAPIException
     */
    Redirect getRedirect(String redirectID)
            throws IOException, BadRequestException, RedirectNotFoundException,
            RedirectAPIException;

    /**
     * @param count
     * @param cursor
     * @return throws an exception to indicate an error
     * @throws IOException
     * @throws BadRequestException
     * @throws RedirectNotFoundException
     * @throws RedirectAPIException
     */
    Redirects getRedirects(String count, String cursor)
            throws IOException, BadRequestException, RedirectNotFoundException,
            RedirectAPIException;
}
