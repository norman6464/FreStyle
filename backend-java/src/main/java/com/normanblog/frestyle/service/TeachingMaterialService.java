package com.normanblog.frestyle.service;

import com.normanblog.frestyle.entity.Course;
import com.normanblog.frestyle.entity.TeachingMaterial;
import com.normanblog.frestyle.entity.User;
import com.normanblog.frestyle.repository.CourseRepository;
import com.normanblog.frestyle.repository.TeachingMaterialRepository;
import org.springframework.http.HttpStatus;
import org.springframework.stereotype.Service;
import org.springframework.web.server.ResponseStatusException;

/** 教材単体の閲覧を担うサービス。所属コースの閲覧権 + 教材自身の公開状態で判定する。 */
@Service
public class TeachingMaterialService {

  private final TeachingMaterialRepository materials;
  private final CourseRepository courses;
  private final CourseService courseService;

  public TeachingMaterialService(
      TeachingMaterialRepository materials, CourseRepository courses, CourseService courseService) {
    this.materials = materials;
    this.courses = courses;
    this.courseService = courseService;
  }

  /** 教材詳細。所属コースが閲覧可能で、かつ公開済み(または管理者)でなければ 403。存在しなければ 404。 */
  public TeachingMaterial get(Long id, User actor) {
    TeachingMaterial material =
        materials
            .findById(id)
            .orElseThrow(
                () -> new ResponseStatusException(HttpStatus.NOT_FOUND, "material_not_found"));
    Course course =
        courses
            .findById(material.getCourseId())
            .orElseThrow(
                () -> new ResponseStatusException(HttpStatus.NOT_FOUND, "course_not_found"));

    courseService.requireReadable(course, actor);
    if (!material.isPublished() && !CourseService.isManager(actor.getRole())) {
      throw new ResponseStatusException(HttpStatus.FORBIDDEN, "forbidden");
    }

    return material;
  }
}
