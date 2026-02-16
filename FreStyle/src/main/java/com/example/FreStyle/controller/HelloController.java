package com.example.FreStyle.controller;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import lombok.extern.slf4j.Slf4j;

@RestController
@RequestMapping("/api")
@Slf4j
public class HelloController {

  @GetMapping("/hello")
  public String hello() {
    log.debug("Hello health check endpoint called");
    return "hello";
  }

}
