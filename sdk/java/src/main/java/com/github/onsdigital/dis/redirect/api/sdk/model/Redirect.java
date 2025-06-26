package com.github.onsdigital.dis.redirect.api.sdk.model;

import lombok.Getter;
import lombok.Setter;

/**
 * The model of a redirect as provided by the redirect API.
 */
public class Redirect {
    /** The original path to redirect from. */
    private @Setter @Getter String from;

    /** The destination path to redirect to. */
    private @Setter @Getter String to;

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
        this.from = fromPath;
        this.to = toPath;
    }
}
