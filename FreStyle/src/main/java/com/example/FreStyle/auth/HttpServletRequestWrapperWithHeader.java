package com.example.FreStyle.auth;

import java.util.Collections;
import java.util.Enumeration;
import java.util.HashMap;
import java.util.HashSet;
import java.util.Map;
import java.util.Set;


import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletRequestWrapper;

// Cookieが入っているトークンをSpring Securityが標準的に扱えるAuthorizationヘッダーとして渡せるようになる
public class HttpServletRequestWrapperWithHeader extends HttpServletRequestWrapper {
  
  private final Map<String, String> customHeaders = new HashMap<>();
  
  public HttpServletRequestWrapperWithHeader(HttpServletRequest request, String headerName, String headerValue) {
    super(request);
    this.customHeaders.put(headerName, headerValue);
  }
  
  @Override
  public String getHeader(String name) {
    String headerValue = customHeaders.get(name);
    if (headerValue != null) {
      return headerValue;
    }
    return super.getHeader(name);
  }
  
  
  @Override
  public Enumeration<String> getHeaderNames() {
    Set<String> names = new HashSet<>(customHeaders.keySet());
    Enumeration<String> parent = super.getHeaderNames();
    while (parent.hasMoreElements()) {
      names.add(parent.nextElement());
    }
    
    return Collections.enumeration(names);
  }
  
}
