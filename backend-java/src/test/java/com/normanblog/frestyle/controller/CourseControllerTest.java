package com.normanblog.frestyle.controller;

import static org.hamcrest.Matchers.hasItem;
import static org.hamcrest.Matchers.hasSize;
import static org.hamcrest.Matchers.not;
import static org.springframework.security.test.web.servlet.request.SecurityMockMvcRequestPostProcessors.jwt;
import static org.springframework.security.test.web.servlet.setup.SecurityMockMvcConfigurers.springSecurity;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.get;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.jsonPath;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.status;

import com.normanblog.frestyle.entity.Course;
import com.normanblog.frestyle.entity.Role;
import com.normanblog.frestyle.entity.User;
import com.normanblog.frestyle.repository.CourseRepository;
import com.normanblog.frestyle.repository.UserRepository;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.test.web.servlet.MockMvc;
import org.springframework.test.web.servlet.setup.MockMvcBuilders;
import org.springframework.web.context.WebApplicationContext;

/** コース閲覧 API の認可(company / published / role)を検証する。 */
@SpringBootTest
class CourseControllerTest {

  @Autowired private WebApplicationContext context;
  @Autowired private UserRepository users;
  @Autowired private CourseRepository courses;

  private MockMvc mvc;

  @BeforeEach
  void setUp() {
    users.deleteAll();
    courses.deleteAll();
    mvc = MockMvcBuilders.webAppContextSetup(context).apply(springSecurity()).build();
  }

  private void seedUser(String sub, Long companyId, String role) {
    users.save(User.builder().cognitoSub(sub).companyId(companyId).role(role).build());
  }

  private Long seedCourse(Long companyId, boolean published, String title) {
    return courses
        .save(
            Course.builder()
                .companyId(companyId)
                .createdByUserId(1L)
                .title(title)
                .isPublished(published)
                .build())
        .getId();
  }

  @Test
  void list_withoutAuth_returns401() throws Exception {
    mvc.perform(get("/api/v2/courses")).andExpect(status().isUnauthorized());
  }

  @Test
  void list_asTrainee_returnsOnlyPublishedOfOwnCompany() throws Exception {
    seedUser("t1", 1L, Role.TRAINEE);
    seedCourse(1L, true, "公開A");
    seedCourse(1L, false, "下書きB");
    seedCourse(2L, true, "他社C");

    mvc.perform(get("/api/v2/courses").with(jwt().jwt(j -> j.subject("t1"))))
        .andExpect(status().isOk())
        .andExpect(jsonPath("$", hasSize(1)))
        .andExpect(jsonPath("$[*].title", hasItem("公開A")))
        .andExpect(jsonPath("$[*].title", not(hasItem("下書きB"))))
        .andExpect(jsonPath("$[*].title", not(hasItem("他社C"))));
  }

  @Test
  void list_asCompanyAdmin_includesDrafts() throws Exception {
    seedUser("a1", 1L, Role.COMPANY_ADMIN);
    seedCourse(1L, true, "公開A");
    seedCourse(1L, false, "下書きB");

    mvc.perform(get("/api/v2/courses").with(jwt().jwt(j -> j.subject("a1"))))
        .andExpect(status().isOk())
        .andExpect(jsonPath("$", hasSize(2)));
  }

  @Test
  void list_withoutCompany_returnsEmpty() throws Exception {
    seedUser("n1", null, Role.TRAINEE);
    seedCourse(1L, true, "公開A");

    mvc.perform(get("/api/v2/courses").with(jwt().jwt(j -> j.subject("n1"))))
        .andExpect(status().isOk())
        .andExpect(jsonPath("$", hasSize(0)));
  }

  @Test
  void get_unpublishedAsTrainee_returns403() throws Exception {
    seedUser("t1", 1L, Role.TRAINEE);
    Long draftId = seedCourse(1L, false, "下書き");

    mvc.perform(get("/api/v2/courses/" + draftId).with(jwt().jwt(j -> j.subject("t1"))))
        .andExpect(status().isForbidden());
  }

  @Test
  void get_otherCompanyCourse_returns403() throws Exception {
    seedUser("t1", 1L, Role.TRAINEE);
    Long otherId = seedCourse(2L, true, "他社公開");

    mvc.perform(get("/api/v2/courses/" + otherId).with(jwt().jwt(j -> j.subject("t1"))))
        .andExpect(status().isForbidden());
  }

  @Test
  void get_missingCourse_returns404() throws Exception {
    seedUser("t1", 1L, Role.TRAINEE);

    mvc.perform(get("/api/v2/courses/999999").with(jwt().jwt(j -> j.subject("t1"))))
        .andExpect(status().isNotFound());
  }
}
