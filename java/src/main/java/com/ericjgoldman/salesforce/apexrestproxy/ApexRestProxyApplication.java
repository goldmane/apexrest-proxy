package com.ericjgoldman.salesforce.apexrestproxy;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.context.annotation.Bean;

import org.springframework.cloud.gateway.route.RouteLocator;
import org.springframework.cloud.gateway.route.builder.RouteLocatorBuilder;

@SpringBootApplication
public class ApexRestProxyApplication {

	public static void main(String[] args) {
		SpringApplication.run(ApexRestProxyApplication.class, args);
	}

}
