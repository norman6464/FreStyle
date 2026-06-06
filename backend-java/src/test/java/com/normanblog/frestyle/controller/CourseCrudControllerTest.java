package com.normanblog.frestyle.controller;

import static org.assertj.core.api.Assertions.assertThat;
import static org.springframework.security.test.web.servlet.request.SecurityMockMvcRequestPostProcessors.jwt;
import static org.springframework.security.test.web.servlet.setup.SecurityMockMvcConfigurers.springSecurity;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.delete;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.post;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.put;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.jsonPath;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.status;

import com.normanblog.frestyle.entity.Course;
import com.normanblog.frestyle.entity.Role;
import com.normanblog.frestyle.entity.TeachingMaterial;
import com.normanblog.frestyle.entity.User;
import com.normanblog.frestyle.repository.CourseRepository;
import com.normanblog.frestyle.repository.TeachingMaterialRepository;
import com.normanblog.frestyle.repository.UserRepository;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.http.MediaType;
import org.springframework.test.web.servlet.MockMvc;
import org.springframework.test.web.servlet.setup.MockMvcBuilders;
import org.springframework.web.context.WebApplicationContext;

/** コース・教材 CRUD の認可とロジックを検証する。 */
@SpringBootTest
class CourseCrudControllerTest {

  @Autowired private WebApplicationContext context;
  @Autowired private UserRepository users;
  @Autowired private CourseRepository courses;
  @Autowired private TeachingMaterialRepository materials;

  private MockMvc mvc;

  @BeforeEach
  void setUp() {
    users.deleteAll();
    courses.deleteAll();
    materials.deleteAll();
    mvc = MockMvcBuilders.webAppContextSetup(context).apply(springSecurity()).build();
  }

  private void seedUser(String sub, Long companyId, String role) {
    users.save(User.builder().cognitoSub(sub).companyId(companyId).role(role).build());
  }

  private Long seedCourse(Long companyId) {
    return courses
        .save(Course.builder().companyId(companyId).createdByUserId(1L).title("c").build())
        .getId();
  }

  @Test
  void create_asTrainee_returns403() throws Exception {
    seedUser("t1", 1L, Role.TRAINEE);

    mvc.perform(
            post("/api/v2/courses")
                .with(jwt().jwt(j -> j.subject("t1")))
                .contentType(MediaType.APPLICATION_JSON)
                .content("{\"title\":\"新コース\"}"))
        .andExpect(status().isForbidden());
  }

  @Test
  void create_asCompanyAdmin_succeedsAndSetsCompany() throws Exception {
    seedUser("a1", 5L, Role.COMPANY_ADMIN);

    mvc.perform(
            post("/api/v2/courses")
                .with(jwt().jwt(j -> j.subject("a1")))
                .contentType(MediaType.APPLICATION_JSON)
                .content("{\"title\":\"新コース\",\"isPublished\":true}"))
        .andExpect(status().isOk())
        .andExpect(jsonPath("$.title").value("新コース"))
        .andExpect(jsonPath("$.companyId").value(5))
        .andExpect(jsonPath("$.isPublished").value(true))
        .andExpect(jsonPath("$.sortOrder").value(100));
  }

  @Test
  void update_otherCompanyCourse_asCompanyAdmin_returns403() throws Exception {
    seedUser("a1", 1L, Role.COMPANY_ADMIN);
    Long otherCourse = seedCourse(2L);

    mvc.perform(
            put("/api/v2/courses/" + otherCourse)
                .with(jwt().jwt(j -> j.subject("a1")))
                .contentType(MediaType.APPLICATION_JSON)
                .content("{\"title\":\"乗っ取り\"}"))
        .andExpect(status().isForbidden());
  }

  @Test
  void delete_cascadesMaterials() throws Exception {
    seedUser("a1", 1L, Role.COMPANY_ADMIN);
    Long courseId = seedCourse(1L);
    materials.save(
        TeachingMaterial.builder()
            .companyId(1L)
            .courseId(courseId)
            .createdByUserId(1L)
            .title("m")
            .build());

    mvc.perform(delete("/api/v2/courses/" + courseId).with(jwt().jwt(j -> j.subject("a1"))))
        .andExpect(status().isNoContent());

    assertThat(courses.findById(courseId)).isEmpty();
    assertThat(materials.findByCourseIdOrderByOrderInCourseAscIdAsc(courseId)).isEmpty();
  }

  @Test
  void createMaterial_withoutCourseId_returns400() throws Exception {
    seedUser("a1", 1L, Role.COMPANY_ADMIN);

    mvc.perform(
            post("/api/v2/teaching-materials")
                .with(jwt().jwt(j -> j.subject("a1")))
                .contentType(MediaType.APPLICATION_JSON)
                .content("{\"title\":\"教材\"}"))
        .andExpect(status().isBadRequest());
  }

  @Test
  void createMaterial_inOwnCourse_succeeds() throws Exception {
    seedUser("a1", 1L, Role.COMPANY_ADMIN);
    Long courseId = seedCourse(1L);

    mvc.perform(
            post("/api/v2/teaching-materials")
                .with(jwt().jwt(j -> j.subject("a1")))
                .contentType(MediaType.APPLICATION_JSON)
                .content("{\"courseId\":" + courseId + ",\"title\":\"教材1\",\"content\":\"本文\"}"))
        .andExpect(status().isOk())
        .andExpect(jsonPath("$.title").value("教材1"))
        .andExpect(jsonPath("$.courseId").value(courseId))
        .andExpect(jsonPath("$.companyId").value(1));
  }
}
