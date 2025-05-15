package com.github.onsdigital.dis.redirect.api.client.model;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;
import com.fasterxml.jackson.annotation.JsonInclude;
import com.fasterxml.jackson.databind.PropertyNamingStrategies;
import com.fasterxml.jackson.databind.annotation.JsonNaming;

import lombok.Getter;
import lombok.Setter;
import lombok.ToString;

/**
 * The model of an helloworld as provided by the redirect API.
 * TODO: Remove this.
 */
@JsonNaming(PropertyNamingStrategies.SnakeCaseStrategy.class)
@JsonIgnoreProperties(ignoreUnknown = true)
@ToString
public class HelloWorld {

    /**
     * Hello world message.
     */
    @JsonInclude(JsonInclude.Include.NON_EMPTY)
    @Getter
    @Setter
    private String message;

}
