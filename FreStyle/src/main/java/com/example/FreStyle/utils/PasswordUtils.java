package com.example.FreStyle.utils;

import org.springframework.security.crypto.bcrypt.BCryptPasswordEncoder;

public class PasswordUtils {
  private static final BCryptPasswordEncoder encoder = new BCryptPasswordEncoder();
  
  public static String hash(String password) {
    return encoder.encode(password);
  }
}
