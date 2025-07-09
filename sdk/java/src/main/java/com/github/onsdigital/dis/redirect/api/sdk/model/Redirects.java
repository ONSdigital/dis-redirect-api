package com.github.onsdigital.dis.redirect.api.sdk.model;

import lombok.Getter;
import lombok.Setter;
import java.util.ArrayList;

/**
 * The model of a redirects object, which contains a list of redirect objects, as provided by the redirect API.
 */
public class Redirects {
    /** The approximate number of redirects requested, defaulted to 10 and limited to 1000. */
    @Getter
    @Setter
    private int count;

    /** The list of Redirect objects. */
    @Getter
    @Setter
    private ArrayList<Redirect> redirectList;

    /** The cursor value returned from a previous response. 0 should be used for the first request. */
    @Getter
    @Setter
    private String cursor;

    /** The cursor value to use in the next request. */
    @Getter
    @Setter
    private String next_cursor;

    /** The total number of redirects in the store. */
    @Getter
    @Setter
    private int total_count;

    /**
     * Default no-argument constructor.
     */
    public Redirects() {
        // Default constructor
    }

    /**
     * Constructs a new Redirects object with the specified values.
     *
     * @param count - the number of redirects requested
     * @param redirectList - the list of Redirect objects
     * @param cursor - the position to start counting from
     * @param next_cursor - the cursor value to use in the next request
     * @param total_count - the total number of redirects in the store
     */
    public Redirects(final int count, final ArrayList<Redirect> redirectList, String cursor, String next_cursor, int total_count) {
        this.count = count;
        this.redirectList = redirectList;
        this.cursor = cursor;
        this.next_cursor = next_cursor;
        this.total_count = total_count;
    }
}
