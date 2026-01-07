package com.example.FreStyle.controller;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

@RestController
@RequestMapping("/api")
public class HelloController {
  
  @GetMapping("/hello")
  public String hello() {
    System.out.println("========== Hello Endpoint Called ==========");
    System.out.println("Hello health check endpoint called.");
    System.out.println("This is a test endpoint to verify the backend is running.");


    return "hello";
  }
  
}
