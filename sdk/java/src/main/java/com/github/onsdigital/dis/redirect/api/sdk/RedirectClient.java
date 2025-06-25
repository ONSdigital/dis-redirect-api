package com.github.onsdigital.dis.redirect.api.sdk;

import java.io.Closeable;
import java.io.IOException;

import com.github.onsdigital.dis.redirect.api.sdk.exception.BadRequestException;
import com.github.onsdigital.dis.redirect.api.sdk.exception.RedirectAPIException;
import com.github.onsdigital.dis.redirect.api.sdk.exception.RedirectNotFoundException;
import com.github.onsdigital.dis.redirect.api.sdk.model.Redirect;

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
       * Upserts a redirect by sending a PUT request to the /redirects/{id}
       * endpoint.
       *
       * The {@code id} must be a base64 URL-encoded version of the
       * {@code from} path in the {@link Redirect} payload.
       * This method will create or update the redirect mapping in the
       * remote API.
       *
       * @param base64Id the base64 URL-encoded identifier representing
       *                 the {@code from} path
       * @param payload  the {@link Redirect} object containing the source
       *                 and target paths
       * @throws IOException          if an I/O error occurs during the request
       * @throws RedirectAPIException if the API returns an unexpected status code
       */
      void putRedirect(String base64Id, Redirect payload) throws IOException,
           RedirectAPIException;

}
