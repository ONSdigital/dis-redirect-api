package com.github.onsdigital.dis.redirect.api.client;

import java.io.Closeable;
import java.io.IOException;

import com.github.onsdigital.dis.redirect.api.client.exception.RedirectAPIException;
import com.github.onsdigital.dis.redirect.api.client.model.HelloWorld;

public interface RedirectClient extends Closeable {
    /**
     * Hello world implementation for repo and test setup.
     * TODO: Remove this.
     *
     * @return HelloWorld
     */
    HelloWorld getHelloWorld() throws IOException, RedirectAPIException;
}
