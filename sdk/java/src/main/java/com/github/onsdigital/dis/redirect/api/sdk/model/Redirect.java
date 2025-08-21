package com.github.onsdigital.dis.redirect.api.sdk.model;

import java.util.Base64;

import org.apache.commons.lang3.StringUtils;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;

import lombok.Getter;
import lombok.Setter;

/**
 * The model of a redirect as provided by the redirect API.
 */
@JsonIgnoreProperties(ignoreUnknown = true)
public class Redirect {
    /** The original path to redirect from. */
    @Getter
    @Setter
    private String from;

    /** The destination path to redirect to. */
    @Getter
    @Setter
    private String to;

    /** The id - base64 encoded version of from. */
    @Getter
    @Setter
    private String id;

    /**
     * Default no-argument constructor.
     */
    public Redirect() {
        // Default constructor
    }

    /**
     * Constructs a new Redirect with the specified from and to paths.
     *
     * @param fromPath the source path
     * @param toPath   the target path
     */
    public Redirect(final String fromPath, final String toPath) {
        this(fromPath, toPath, Redirect.encodeRedirectID(fromPath));
    }

    /**
     * Constructs a new Redirect with the specified from and to paths.
     *
     * @param fromPath  the source path
     * @param toPath    the target path
     * @param encodedId the id of the redirect
     */
    public Redirect(final String fromPath, final String toPath,
            final String encodedId) {
        this.from = fromPath;
        this.to = toPath;
        this.id = encodedId;
    }

    /**
     * Encodes a string as a redirect ID - consolidated functionality
     * to keep encoding in one place for consistency.
     *
     * @param str
     * @return base64 encoded string
     */
    public static String encodeRedirectID(final String str) {
        if (StringUtils.isNotEmpty(str)) {
            return Base64.getEncoder().encodeToString(str.getBytes());
        }
        return "";
    }
}
