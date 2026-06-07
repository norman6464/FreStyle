package com.normanblog.frestyle.controller;

import static org.springframework.security.test.web.servlet.request.SecurityMockMvcRequestPostProcessors.jwt;
import static org.springframework.security.test.web.servlet.setup.SecurityMockMvcConfigurers.springSecurity;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.get;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.put;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.jsonPath;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.status;

import com.normanblog.frestyle.entity.Company;
import com.normanblog.frestyle.entity.Role;
import com.normanblog.frestyle.entity.User;
import com.normanblog.frestyle.repository.CompanyRepository;
import com.normanblog.frestyle.repository.UserRepository;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.http.MediaType;
import org.springframework.test.web.servlet.MockMvc;
import org.springframework.test.web.servlet.setup.MockMvcBuilders;
import org.springframework.web.context.WebApplicationContext;

/** 会社設定 API の認可(company_admin のみ)と更新の永続化を検証する。 */
@SpringBootTest
class CompanySettingsControllerTest {

  private static final String ADMIN_SUB = "company-admin-sub";
  private static final String TRAINEE_SUB = "trainee-sub";
  private static final String URL = "/api/v2/company/settings";

  @Autowired private WebApplicationContext context;
  @Autowired private UserRepository users;
  @Autowired private CompanyRepository companies;

  private MockMvc mvc;
  private Long companyId;

  @BeforeEach
  void setUp() {
    users.deleteAll();
    companies.deleteAll();
    Company company =
        companies.save(Company.builder().name("Acme").aiChatEnabledForTrainees(true).build());
    companyId = company.getId();
    users.save(
        User.builder()
            .cognitoSub(ADMIN_SUB)
            .email("admin@acme.com")
            .role(Role.COMPANY_ADMIN)
            .companyId(companyId)
            .build());
    users.save(
        User.builder()
            .cognitoSub(TRAINEE_SUB)
            .email("trainee@acme.com")
            .role(Role.TRAINEE)
            .companyId(companyId)
            .build());
    mvc = MockMvcBuilders.webAppContextSetup(context).apply(springSecurity()).build();
  }

  @Test
  void get_asCompanyAdmin_returnsCurrentFlag() throws Exception {
    mvc.perform(get(URL).with(jwt().jwt(j -> j.subject(ADMIN_SUB))))
        .andExpect(status().isOk())
        .andExpect(jsonPath("$.aiChatEnabledForTrainees").value(true));
  }

  @Test
  void get_asTrainee_returns403() throws Exception {
    mvc.perform(get(URL).with(jwt().jwt(j -> j.subject(TRAINEE_SUB))))
        .andExpect(status().isForbidden());
  }

  @Test
  void put_asCompanyAdmin_disablesAndPersists() throws Exception {
    mvc.perform(
            put(URL)
                .with(jwt().jwt(j -> j.subject(ADMIN_SUB)))
                .contentType(MediaType.APPLICATION_JSON)
                .content("{\"aiChatEnabledForTrainees\":false}"))
        .andExpect(status().isOk())
        .andExpect(jsonPath("$.aiChatEnabledForTrainees").value(false));

    // 永続化を確認(再取得で false)。
    mvc.perform(get(URL).with(jwt().jwt(j -> j.subject(ADMIN_SUB))))
        .andExpect(jsonPath("$.aiChatEnabledForTrainees").value(false));
  }

  @Test
  void put_asTrainee_returns403() throws Exception {
    mvc.perform(
            put(URL)
                .with(jwt().jwt(j -> j.subject(TRAINEE_SUB)))
                .contentType(MediaType.APPLICATION_JSON)
                .content("{\"aiChatEnabledForTrainees\":false}"))
        .andExpect(status().isForbidden());
  }

  @Test
  void put_missingField_returns400() throws Exception {
    mvc.perform(
            put(URL)
                .with(jwt().jwt(j -> j.subject(ADMIN_SUB)))
                .contentType(MediaType.APPLICATION_JSON)
                .content("{}"))
        .andExpect(status().isBadRequest());
  }
}
