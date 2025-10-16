package com.example.FreStyle.utils;

import org.springframework.context.annotation.Configuration;
import org.springframework.security.crypto.bcrypt.BCryptPasswordEncoder;

@Configuration
public class PasswordUtils {
  private static final BCryptPasswordEncoder encoder = new BCryptPasswordEncoder();
  
  public static String hash(String password) {
    return encoder.encode(password);
  }
}
