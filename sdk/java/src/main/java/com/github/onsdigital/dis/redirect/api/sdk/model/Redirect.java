package com.github.onsdigital.dis.redirect.api.sdk.model;

/**
 * The model of a redirect as provided by the redirect API.
 */
public class Redirect {
    /** The original path to redirect from. */
    private String from;

    /** The destination path to redirect to. */
    private String to;

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

    /**
     * Gets the source path of the redirect.
     *
     * @return the from path
     */
    public String getFrom() {
        return this.from;
    }

    /**
     * Sets the source path of the redirect.
     *
     * @param fromPath the new from path
     */
    public void setFrom(final String fromPath) {
        this.from = fromPath;
    }

    /**
     * Gets the target path of the redirect.
     *
     * @return the to path
     */
    public String getTo() {
        return this.to;
    }

    /**
     * Sets the target path of the redirect.
     *
     * @param toPath the new to path
     */
    public void setTo(final String toPath) {
        this.to = toPath;
    }
}
