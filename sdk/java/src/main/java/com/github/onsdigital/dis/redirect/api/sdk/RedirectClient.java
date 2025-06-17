package com.github.onsdigital.dis.redirect.api.sdk;

import java.io.Closeable;
import java.io.IOException;

import com.github.onsdigital.dis.redirect.api.sdk.exception.RedirectAPIException;
import com.github.onsdigital.dis.redirect.api.sdk.model.HelloWorld;

public interface RedirectClient extends Closeable {
    /**
     * Hello world implementation for repo and test setup.
     * TODO: Remove this.
     *
     * @return HelloWorld
     */
    HelloWorld getHelloWorld() throws IOException, RedirectAPIException;
}
