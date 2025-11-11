package com.example.FreStyle;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.boot.web.servlet.FilterRegistrationBean;
import org.springframework.context.annotation.Bean;
import org.springframework.web.filter.ForwardedHeaderFilter;

@SpringBootApplication
public class FreStyleApplication {

	public static void main(String[] args) {
		SpringApplication.run(FreStyleApplication.class, args);
	}
	
	@Bean
public FilterRegistrationBean<ForwardedHeaderFilter> forwardedHeaderFilter() {
    FilterRegistrationBean<ForwardedHeaderFilter> filterRegBean = new FilterRegistrationBean<>();
    filterRegBean.setFilter(new ForwardedHeaderFilter());
		filterRegBean.setOrder(0);
		filterRegBean.addUrlPatterns("/*");
    return filterRegBean;
}


}
