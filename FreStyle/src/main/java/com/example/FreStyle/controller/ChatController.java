package com.example.FreStyle.controller;

import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

@RestController
@RequestMapping("/api/chat/")
public class ChatController {
  
  
  @GetMapping("/history")
  public ResponseEntity<?> history()
}
