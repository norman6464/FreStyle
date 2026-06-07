package com.normanblog.frestyle.config;

import static org.springframework.security.test.web.servlet.setup.SecurityMockMvcConfigurers.springSecurity;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.get;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.options;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.header;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.status;

import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.http.HttpHeaders;
import org.springframework.test.web.servlet.MockMvc;
import org.springframework.test.web.servlet.setup.MockMvcBuilders;
import org.springframework.web.context.WebApplicationContext;

/**
 * フロント(normanblog.com)と API(api.normanblog.com)が別オリジンのため、許可オリジンには
 * CORS ヘッダ(credentials 付き)を返すことを検証する。これが無いとブラウザが全 API を遮断し
 * メンテナンス画面になる(本番障害の再発防止)。
 */
@SpringBootTest
class CorsConfigTest {

  private static final String PROD_ORIGIN = "https://normanblog.com";

  @Autowired private WebApplicationContext context;
  private MockMvc mvc;

  @BeforeEach
  void setUp() {
    mvc = MockMvcBuilders.webAppContextSetup(context).apply(springSecurity()).build();
  }

  @Test
  void preflight_fromAllowedOrigin_returnsCorsHeaders() throws Exception {
    mvc.perform(
            options("/api/v2/health")
                .header(HttpHeaders.ORIGIN, PROD_ORIGIN)
                .header("Access-Control-Request-Method", "GET"))
        .andExpect(status().isOk())
        .andExpect(header().string("Access-Control-Allow-Origin", PROD_ORIGIN))
        .andExpect(header().string("Access-Control-Allow-Credentials", "true"));
  }

  @Test
  void simpleGet_fromAllowedOrigin_echoesAllowOrigin() throws Exception {
    mvc.perform(get("/api/v2/health").header(HttpHeaders.ORIGIN, PROD_ORIGIN))
        .andExpect(status().isOk())
        .andExpect(header().string("Access-Control-Allow-Origin", PROD_ORIGIN))
        .andExpect(header().string("Access-Control-Allow-Credentials", "true"));
  }

  @Test
  void preflight_fromDisallowedOrigin_hasNoAllowOrigin() throws Exception {
    mvc.perform(
            options("/api/v2/health")
                .header(HttpHeaders.ORIGIN, "https://evil.example.com")
                .header("Access-Control-Request-Method", "GET"))
        .andExpect(header().doesNotExist("Access-Control-Allow-Origin"));
  }
}
