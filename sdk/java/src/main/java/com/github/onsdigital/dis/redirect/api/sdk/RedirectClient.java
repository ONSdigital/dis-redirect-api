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

     /**
      * Upserts a redirect by sending a PUT request to the /redirects/{id}
      * endpoint.
      *
      * The {@code id} must be a base64 URL-encoded version of the
      * {@code from} path in the {@link Redirect} payload.
      * This method will create or update the redirect mapping in the
      * remote API.
      *
      * @param payload  the {@link Redirect} object containing the source
      *                 and target paths
      * @throws IOException          if an I/O error occurs during the request
      * @throws RedirectAPIException if the API returns an unexpected status
      *                 code
      */
    void putRedirect(Redirect payload) throws IOException,
       RedirectAPIException;

    /**
       * Deletes a redirect by sending a DELETE request to the /redirects/{id}
       * endpoint.
       *
       * The {@code id} must be a base64 URL-encoded version of the path to
       * delete. This method attempts to delete the redirect mapping in the
       * remote API.
       *
       * A {@code 204 No Content} status indicates successful deletion.
       * A {@code 404 Not Found} status indicates the redirect does not exist.
       *
       * @param base64Id the base64 URL-encoded identifier representing the path
       *                 to delete
       * @throws IOException          if an I/O error occurs during the request
       * @throws RedirectAPIException if the API returns an unexpected status
       *                              code
       */
    void deleteRedirect(String base64Id) throws IOException,
      RedirectAPIException;

}
