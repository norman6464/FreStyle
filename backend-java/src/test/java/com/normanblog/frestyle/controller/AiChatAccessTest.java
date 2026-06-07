package com.normanblog.frestyle.controller;

import static org.springframework.security.test.web.servlet.request.SecurityMockMvcRequestPostProcessors.jwt;
import static org.springframework.security.test.web.servlet.setup.SecurityMockMvcConfigurers.springSecurity;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.get;
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
import org.springframework.test.web.servlet.MockMvc;
import org.springframework.test.web.servlet.setup.MockMvcBuilders;
import org.springframework.web.context.WebApplicationContext;

/**
 * 会社設定による AI チャットの遮断(サーバ側 403)と、/auth/me の aiChatEnabledForTrainees を検証する。
 * UI 非表示だけでなく API 直叩きも止まることを担保する。
 */
@SpringBootTest
class AiChatAccessTest {

  private static final String AI = "/api/v2/ai-chat/sessions";
  private static final String ME = "/api/v2/auth/me";

  @Autowired private WebApplicationContext context;
  @Autowired private UserRepository users;
  @Autowired private CompanyRepository companies;

  private MockMvc mvc;

  @BeforeEach
  void setUp() {
    users.deleteAll();
    companies.deleteAll();
    mvc = MockMvcBuilders.webAppContextSetup(context).apply(springSecurity()).build();
  }

  private void seedUser(String sub, String role, boolean companyAiEnabled) {
    Company company =
        companies.save(
            Company.builder().name("Acme").aiChatEnabledForTrainees(companyAiEnabled).build());
    users.save(
        User.builder()
            .cognitoSub(sub)
            .email(sub + "@acme.com")
            .role(role)
            .companyId(company.getId())
            .build());
  }

  @Test
  void trainee_companyDisabled_aiChatReturns403() throws Exception {
    seedUser("t-off", Role.TRAINEE, false);
    mvc.perform(get(AI).with(jwt().jwt(j -> j.subject("t-off"))))
        .andExpect(status().isForbidden());
  }

  @Test
  void trainee_companyEnabled_aiChatAllowed() throws Exception {
    seedUser("t-on", Role.TRAINEE, true);
    mvc.perform(get(AI).with(jwt().jwt(j -> j.subject("t-on")))).andExpect(status().isOk());
  }

  @Test
  void companyAdmin_companyDisabled_aiChatStillAllowed() throws Exception {
    seedUser("admin-off", Role.COMPANY_ADMIN, false);
    mvc.perform(get(AI).with(jwt().jwt(j -> j.subject("admin-off")))).andExpect(status().isOk());
  }

  @Test
  void me_reflectsCompanyAiFlag() throws Exception {
    seedUser("t-off2", Role.TRAINEE, false);
    mvc.perform(get(ME).with(jwt().jwt(j -> j.subject("t-off2"))))
        .andExpect(status().isOk())
        .andExpect(jsonPath("$.aiChatEnabledForTrainees").value(false));

    seedUser("t-on2", Role.TRAINEE, true);
    mvc.perform(get(ME).with(jwt().jwt(j -> j.subject("t-on2"))))
        .andExpect(status().isOk())
        .andExpect(jsonPath("$.aiChatEnabledForTrainees").value(true));
  }
}
