package com.normanblog.frestyle.controller;

import static org.assertj.core.api.Assertions.assertThat;
import static org.springframework.security.test.web.servlet.request.SecurityMockMvcRequestPostProcessors.jwt;
import static org.springframework.security.test.web.servlet.setup.SecurityMockMvcConfigurers.springSecurity;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.get;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.post;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.jsonPath;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.status;

import com.normanblog.frestyle.entity.LearningReport;
import com.normanblog.frestyle.entity.LearningReportStatus;
import com.normanblog.frestyle.entity.Role;
import com.normanblog.frestyle.entity.User;
import com.normanblog.frestyle.repository.LearningReportRepository;
import com.normanblog.frestyle.repository.UserRepository;
import java.time.Instant;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.http.MediaType;
import org.springframework.test.web.servlet.MockMvc;
import org.springframework.test.web.servlet.setup.MockMvcBuilders;
import org.springframework.web.context.WebApplicationContext;

/** 学習レポート API の一覧(自分のみ・期間降順)・生成要求(202/pending)・バリデーションを検証する。 */
@SpringBootTest
class LearningReportControllerTest {

  private static final String SUB = "report-sub";

  @Autowired private WebApplicationContext context;
  @Autowired private UserRepository users;
  @Autowired private LearningReportRepository reports;

  private MockMvc mvc;
  private Long myId;
  private Long otherId;

  @BeforeEach
  void setUp() {
    reports.deleteAll();
    users.deleteAll();
    myId =
        users
            .save(User.builder().cognitoSub(SUB).email("me@example.com").role(Role.TRAINEE).build())
            .getId();
    otherId =
        users
            .save(
                User.builder()
                    .cognitoSub("other")
                    .email("o@example.com")
                    .role(Role.TRAINEE)
                    .build())
            .getId();
    mvc = MockMvcBuilders.webAppContextSetup(context).apply(springSecurity()).build();
  }

  private void save(Long userId, Instant periodTo) {
    reports.save(
        LearningReport.builder()
            .userId(userId)
            .periodFrom(periodTo.minusSeconds(86400))
            .periodTo(periodTo)
            .status(LearningReportStatus.READY)
            .createdAt(Instant.now())
            .build());
  }

  @Test
  void list_withoutAuth_returns401() throws Exception {
    mvc.perform(get("/api/v2/learning-reports")).andExpect(status().isUnauthorized());
  }

  @Test
  void list_returnsOnlyOwn_orderedByPeriodToDesc() throws Exception {
    save(myId, Instant.parse("2026-01-01T00:00:00Z"));
    save(myId, Instant.parse("2026-03-01T00:00:00Z"));
    save(otherId, Instant.parse("2026-12-01T00:00:00Z"));

    mvc.perform(get("/api/v2/learning-reports").with(jwt().jwt(j -> j.subject(SUB))))
        .andExpect(status().isOk())
        .andExpect(jsonPath("$.length()").value(2))
        .andExpect(jsonPath("$[0].periodTo").value("2026-03-01T00:00:00Z"))
        .andExpect(jsonPath("$[1].periodTo").value("2026-01-01T00:00:00Z"));
  }

  @Test
  void generate_createsPendingReportForMonth_returns202() throws Exception {
    mvc.perform(
            post("/api/v2/learning-reports/generate")
                .with(jwt().jwt(j -> j.subject(SUB)))
                .contentType(MediaType.APPLICATION_JSON)
                .content("{\"year\":2026,\"month\":3}"))
        .andExpect(status().isAccepted())
        .andExpect(jsonPath("$.status").value(LearningReportStatus.PENDING))
        .andExpect(jsonPath("$.userId").value(myId))
        // 3 月 → [2026-03-01, 2026-04-01) の半開区間。
        .andExpect(jsonPath("$.periodFrom").value("2026-03-01T00:00:00Z"))
        .andExpect(jsonPath("$.periodTo").value("2026-04-01T00:00:00Z"));

    assertThat(reports.findByUserIdOrderByPeriodToDesc(myId)).hasSize(1);
  }

  @Test
  void generate_invalidMonth_returns400() throws Exception {
    mvc.perform(
            post("/api/v2/learning-reports/generate")
                .with(jwt().jwt(j -> j.subject(SUB)))
                .contentType(MediaType.APPLICATION_JSON)
                .content("{\"year\":2026,\"month\":13}"))
        .andExpect(status().isBadRequest());
  }
}
