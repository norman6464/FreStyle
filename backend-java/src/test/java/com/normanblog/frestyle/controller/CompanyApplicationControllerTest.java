package com.normanblog.frestyle.controller;

import static org.springframework.security.test.web.servlet.request.SecurityMockMvcRequestPostProcessors.jwt;
import static org.springframework.security.test.web.servlet.setup.SecurityMockMvcConfigurers.springSecurity;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.get;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.patch;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.post;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.jsonPath;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.status;

import com.normanblog.frestyle.entity.CompanyApplicationStatus;
import com.normanblog.frestyle.entity.Role;
import com.normanblog.frestyle.entity.User;
import com.normanblog.frestyle.repository.CompanyApplicationRepository;
import com.normanblog.frestyle.repository.UserRepository;
import com.normanblog.frestyle.service.CompanyApplicationService;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.http.MediaType;
import org.springframework.test.web.servlet.MockMvc;
import org.springframework.test.web.servlet.setup.MockMvcBuilders;
import org.springframework.web.context.WebApplicationContext;

/** 企業利用申請 API(公開作成 + super_admin の一覧/status)の HTTP 振る舞いを検証する。 */
@SpringBootTest
class CompanyApplicationControllerTest {

  @Autowired private WebApplicationContext context;
  @Autowired private UserRepository users;
  @Autowired private CompanyApplicationRepository applications;
  @Autowired private CompanyApplicationService service;

  private MockMvc mvc;

  @BeforeEach
  void setUp() {
    users.deleteAll();
    applications.deleteAll();
    mvc = MockMvcBuilders.webAppContextSetup(context).apply(springSecurity()).build();
  }

  private void seedUser(String sub, String role) {
    users.save(User.builder().cognitoSub(sub).email(sub + "@x.com").role(role).build());
  }

  @Test
  void create_isPublic_returns201() throws Exception {
    mvc.perform(
            post("/api/v2/company-applications")
                .contentType(MediaType.APPLICATION_JSON)
                .content(
                    "{\"companyName\":\"Acme\",\"applicantName\":\"山田\",\"email\":\"y@acme.com\"}"))
        .andExpect(status().isCreated())
        .andExpect(jsonPath("$.companyName").value("Acme"))
        .andExpect(jsonPath("$.status").value(CompanyApplicationStatus.PENDING));
  }

  @Test
  void create_missingRequiredField_returns400() throws Exception {
    mvc.perform(
            post("/api/v2/company-applications")
                .contentType(MediaType.APPLICATION_JSON)
                .content("{\"companyName\":\"Acme\"}"))
        .andExpect(status().isBadRequest());
  }

  @Test
  void adminList_asTrainee_returns403() throws Exception {
    seedUser("trainee-sub", Role.TRAINEE);

    mvc.perform(get("/api/v2/admin/company-applications").with(jwt().jwt(j -> j.subject("trainee-sub"))))
        .andExpect(status().isForbidden());
  }

  @Test
  void adminList_asSuperAdmin_returns200() throws Exception {
    seedUser("admin-sub", Role.SUPER_ADMIN);
    service.create("Acme", "山田", "y@acme.com", null);

    mvc.perform(get("/api/v2/admin/company-applications").with(jwt().jwt(j -> j.subject("admin-sub"))))
        .andExpect(status().isOk())
        .andExpect(jsonPath("$[0].companyName").value("Acme"));
  }

  @Test
  void adminUpdateStatus_asSuperAdmin_returns204() throws Exception {
    seedUser("admin-sub", Role.SUPER_ADMIN);
    var app = service.create("Acme", "山田", "y@acme.com", null);

    mvc.perform(
            patch("/api/v2/admin/company-applications/{id}/status", app.getId())
                .with(jwt().jwt(j -> j.subject("admin-sub")))
                .contentType(MediaType.APPLICATION_JSON)
                .content("{\"status\":\"approved\"}"))
        .andExpect(status().isNoContent());
  }
}
