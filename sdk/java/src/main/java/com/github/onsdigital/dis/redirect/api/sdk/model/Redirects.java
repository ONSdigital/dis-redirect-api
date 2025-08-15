package com.github.onsdigital.dis.redirect.api.sdk.model;

import lombok.Getter;
import lombok.Setter;
import java.util.ArrayList;

/**
 * The model of a redirects object, which contains a list of redirect objects,
 * as provided by the redirect API.
 */
public class Redirects {
    /** The number of redirects requested. */
    @Getter
    @Setter
    private int count;

    /** The list of Redirect objects. */
    @Getter
    @Setter
    private ArrayList<Redirect> redirectList;

    /** The cursor value returned from a previous response.
     * 0 should be used for the first request. */
    @Getter
    @Setter
    private String cursor;

    /** The cursor value to use in the next request. */
    @Getter
    @Setter
    private String nextCursor;

    /** The total number of redirects in the store. */
    @Getter
    @Setter
    private int totalCount;

    /**
     * Default no-argument constructor.
     */
    public Redirects() {
        // Default constructor
    }

    /**
     * Constructs a new Redirects object with the specified values.
     *
     * @param initCount - initialises the number of redirects requested
     * @param initRedirectList - the list of Redirect objects
     * @param initCursor - the position to start counting from
     * @param initNextCursor - the cursor value to use in the next request
     * @param initTotalCount - the total number of redirects in the store
     */
    public Redirects(final int initCount,
                     final ArrayList<Redirect> initRedirectList,
                     final String initCursor, final String initNextCursor,
                     final int initTotalCount) {
        this.count = initCount;
        this.redirectList = initRedirectList;
        this.cursor = initCursor;
        this.nextCursor = initNextCursor;
        this.totalCount = initTotalCount;
    }
}
