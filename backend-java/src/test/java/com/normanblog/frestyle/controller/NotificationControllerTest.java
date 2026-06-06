package com.normanblog.frestyle.controller;

import static org.assertj.core.api.Assertions.assertThat;
import static org.springframework.security.test.web.servlet.request.SecurityMockMvcRequestPostProcessors.jwt;
import static org.springframework.security.test.web.servlet.setup.SecurityMockMvcConfigurers.springSecurity;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.get;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.patch;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.jsonPath;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.status;

import com.normanblog.frestyle.entity.Notification;
import com.normanblog.frestyle.entity.Role;
import com.normanblog.frestyle.entity.User;
import com.normanblog.frestyle.repository.NotificationRepository;
import com.normanblog.frestyle.repository.UserRepository;
import java.time.Instant;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.test.web.servlet.MockMvc;
import org.springframework.test.web.servlet.setup.MockMvcBuilders;
import org.springframework.web.context.WebApplicationContext;

/** 通知 API の一覧・既読化・未読数・所有者検証を検証する。 */
@SpringBootTest
class NotificationControllerTest {

  private static final String SUB = "notif-sub";

  @Autowired private WebApplicationContext context;
  @Autowired private UserRepository users;
  @Autowired private NotificationRepository notifications;

  private MockMvc mvc;
  private Long myId;
  private Long otherId;

  @BeforeEach
  void setUp() {
    notifications.deleteAll();
    users.deleteAll();
    myId =
        users
            .save(User.builder().cognitoSub(SUB).email("me@example.com").role(Role.TRAINEE).build())
            .getId();
    otherId =
        users
            .save(
                User.builder()
                    .cognitoSub("other-sub")
                    .email("other@example.com")
                    .role(Role.TRAINEE)
                    .build())
            .getId();
    mvc = MockMvcBuilders.webAppContextSetup(context).apply(springSecurity()).build();
  }

  private Notification save(Long userId, String title, boolean read, Instant createdAt) {
    return notifications.save(
        Notification.builder()
            .userId(userId)
            .type("system")
            .title(title)
            .body("body")
            .isRead(read)
            .createdAt(createdAt)
            .build());
  }

  @Test
  void list_withoutAuth_returns401() throws Exception {
    mvc.perform(get("/api/v2/notifications")).andExpect(status().isUnauthorized());
  }

  @Test
  void list_returnsOnlyOwnNotifications_orderedByCreatedAtDesc() throws Exception {
    save(myId, "古い", false, Instant.parse("2026-01-01T00:00:00Z"));
    save(myId, "新しい", false, Instant.parse("2026-02-01T00:00:00Z"));
    save(otherId, "他人宛", false, Instant.now());

    mvc.perform(get("/api/v2/notifications").with(jwt().jwt(j -> j.subject(SUB))))
        .andExpect(status().isOk())
        .andExpect(jsonPath("$.length()").value(2))
        .andExpect(jsonPath("$[0].title").value("新しい"))
        .andExpect(jsonPath("$[1].title").value("古い"));
  }

  @Test
  void unreadCount_countsOnlyOwnUnread() throws Exception {
    save(myId, "未読1", false, Instant.now());
    save(myId, "未読2", false, Instant.now());
    save(myId, "既読", true, Instant.now());
    save(otherId, "他人未読", false, Instant.now());

    mvc.perform(get("/api/v2/notifications/unread-count").with(jwt().jwt(j -> j.subject(SUB))))
        .andExpect(status().isOk())
        .andExpect(jsonPath("$").value(2));
  }

  @Test
  void markRead_marksOwnNotification() throws Exception {
    Long id = save(myId, "未読", false, Instant.now()).getId();

    mvc.perform(patch("/api/v2/notifications/" + id + "/read").with(jwt().jwt(j -> j.subject(SUB))))
        .andExpect(status().isNoContent());

    assertThat(notifications.findById(id).orElseThrow().isRead()).isTrue();
  }

  @Test
  void markRead_otherUsersNotification_isNotAffected() throws Exception {
    Long id = save(otherId, "他人宛", false, Instant.now()).getId();

    mvc.perform(patch("/api/v2/notifications/" + id + "/read").with(jwt().jwt(j -> j.subject(SUB))))
        .andExpect(status().isNoContent());

    // 所有者検証により他人の通知は既読化されない。
    assertThat(notifications.findById(id).orElseThrow().isRead()).isFalse();
  }

  @Test
  void markAllRead_marksOnlyOwnNotifications() throws Exception {
    save(myId, "未読1", false, Instant.now());
    save(myId, "未読2", false, Instant.now());
    Long otherUnread = save(otherId, "他人未読", false, Instant.now()).getId();

    mvc.perform(patch("/api/v2/notifications/read-all").with(jwt().jwt(j -> j.subject(SUB))))
        .andExpect(status().isNoContent());

    assertThat(notifications.countByUserIdAndIsReadFalse(myId)).isZero();
    assertThat(notifications.findById(otherUnread).orElseThrow().isRead()).isFalse();
  }
}
