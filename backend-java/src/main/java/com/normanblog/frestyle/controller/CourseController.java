package com.normanblog.frestyle.controller;

import com.normanblog.frestyle.dto.CourseRequest;
import com.normanblog.frestyle.dto.CourseResponse;
import com.normanblog.frestyle.dto.TeachingMaterialResponse;
import com.normanblog.frestyle.entity.User;
import com.normanblog.frestyle.security.CurrentUserProvider;
import com.normanblog.frestyle.service.CourseService;
import java.util.List;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.DeleteMapping;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.PutMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

/** コース閲覧 API。current user の company / role で自動フィルタする。 */
@RestController
@RequestMapping("/api/v2/courses")
public class CourseController {

  private final CourseService courseService;
  private final CurrentUserProvider currentUser;

  public CourseController(CourseService courseService, CurrentUserProvider currentUser) {
    this.courseService = courseService;
    this.currentUser = currentUser;
  }

  @GetMapping
  public List<CourseResponse> list() {
    User actor = currentUser.require();

    return courseService.list(actor).stream().map(CourseResponse::from).toList();
  }

  @GetMapping("/{id}")
  public CourseResponse get(@PathVariable Long id) {
    User actor = currentUser.require();

    return CourseResponse.from(courseService.get(id, actor));
  }

  @GetMapping("/{id}/materials")
  public List<TeachingMaterialResponse> materials(@PathVariable Long id) {
    User actor = currentUser.require();

    return courseService.listMaterials(id, actor).stream()
        .map(TeachingMaterialResponse::from)
        .toList();
  }

  @PostMapping
  public CourseResponse create(@RequestBody CourseRequest request) {
    User actor = currentUser.require();

    return CourseResponse.from(courseService.create(actor, request));
  }

  @PutMapping("/{id}")
  public CourseResponse update(@PathVariable Long id, @RequestBody CourseRequest request) {
    User actor = currentUser.require();

    return CourseResponse.from(courseService.update(id, actor, request));
  }

  @DeleteMapping("/{id}")
  public ResponseEntity<Void> delete(@PathVariable Long id) {
    User actor = currentUser.require();
    courseService.delete(id, actor);

    return ResponseEntity.noContent().build();
  }
}
