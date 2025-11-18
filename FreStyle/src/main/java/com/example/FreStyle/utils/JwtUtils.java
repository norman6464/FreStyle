package com.example.FreStyle.utils;

import java.text.ParseException;
import java.util.Optional;

import com.nimbusds.jwt.JWTClaimsSet;
import com.nimbusds.jwt.SignedJWT;

public class JwtUtils {
  
  public static Optional<JWTClaimsSet> decode(String token) {
    try {
      
      SignedJWT signedJWT = SignedJWT.parse(token);
      
      return Optional.of(signedJWT.getJWTClaimsSet());
      
    } catch(ParseException e) {
      return Optional.empty();
    }
  }
  
}
