package com.github.onsdigital.dis.redirect.api.sdk;

import java.io.Closeable;
import java.io.IOException;

import com.github.onsdigital.dis.redirect.api.sdk.exception.BadRequestException;
import com.github.onsdigital.dis.redirect.api.sdk.exception.RedirectAPIException;
import com.github.onsdigital.dis.redirect.api.sdk.exception.RedirectNotFound;
import com.github.onsdigital.dis.redirect.api.sdk.model.Redirect;

public interface RedirectClient extends Closeable {
    /**
     * @param redirectID
     * @return throws an exception to indicate an error
     * @throws IOException
     * @throws BadRequestException
     * @throws RedirectNotFound
     * @throws RedirectAPIException
     */
    Redirect getRedirect(String redirectID) throws IOException,
            BadRequestException, RedirectNotFound, RedirectAPIException;
}
