package com.example.FreStyle.controller;

import java.util.HashMap;
import java.util.Map;

import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

@RestController
@RequestMapping("/api")
public class HelloController {
  
  @GetMapping("/hello")
  public ResponseEntity<Map> hello() {
    
    Map<String,String> response = new HashMap<>();
    response.put("name", "takuma");
    
    return new ResponseEntity<>(response, HttpStatus.OK);
  }
  
}
