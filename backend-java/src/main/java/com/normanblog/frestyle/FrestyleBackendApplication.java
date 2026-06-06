package com.normanblog.frestyle;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.boot.context.properties.ConfigurationPropertiesScan;

@SpringBootApplication
@ConfigurationPropertiesScan
public class FrestyleBackendApplication {

	public static void main(String[] args) {
		SpringApplication.run(FrestyleBackendApplication.class, args);
	}

}
