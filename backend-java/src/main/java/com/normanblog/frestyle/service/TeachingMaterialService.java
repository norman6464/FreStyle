package com.normanblog.frestyle.service;

import com.normanblog.frestyle.dto.TeachingMaterialRequest;
import com.normanblog.frestyle.entity.Course;
import com.normanblog.frestyle.entity.TeachingMaterial;
import com.normanblog.frestyle.entity.User;
import com.normanblog.frestyle.repository.CourseRepository;
import com.normanblog.frestyle.repository.TeachingMaterialRepository;
import java.time.Instant;
import org.springframework.http.HttpStatus;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import org.springframework.web.server.ResponseStatusException;

/** 教材の閲覧・作成・更新・削除を担うサービス。所属コースの権限で認可する。 */
@Service
public class TeachingMaterialService {

  private static final int DEFAULT_ORDER = 100;

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
    TeachingMaterial material = findOrThrow(id);
    Course course = findCourseOrThrow(material.getCourseId());

    courseService.requireReadable(course, actor);
    if (!material.isPublished() && !CourseService.isManager(actor.getRole())) {
      throw new ResponseStatusException(HttpStatus.FORBIDDEN, "forbidden");
    }

    return material;
  }

  /** 教材を作成する。courseId 必須。所属コースの編集権が必要。 */
  @Transactional
  public TeachingMaterial create(User actor, TeachingMaterialRequest request) {
    courseService.requireManagerWithCompany(actor);
    if (request.courseId() == null) {
      throw new ResponseStatusException(HttpStatus.BAD_REQUEST, "course_id is required");
    }

    Course course = findCourseOrThrow(request.courseId());
    courseService.requireManage(course, actor);

    Instant now = Instant.now();
    TeachingMaterial material =
        TeachingMaterial.builder()
            .companyId(course.getCompanyId())
            .courseId(course.getId())
            .createdByUserId(actor.getId())
            .title(request.title())
            .content(request.content())
            .orderInCourse(request.orderInCourse() != null ? request.orderInCourse() : DEFAULT_ORDER)
            .isPublished(Boolean.TRUE.equals(request.isPublished()))
            .createdAt(now)
            .updatedAt(now)
            .build();

    return materials.save(material);
  }

  /** 教材を更新する。所属コースは変えない。編集権が無ければ 403、未存在は 404。 */
  @Transactional
  public TeachingMaterial update(Long id, User actor, TeachingMaterialRequest request) {
    TeachingMaterial material = findOrThrow(id);
    Course course = findCourseOrThrow(material.getCourseId());
    courseService.requireManage(course, actor);

    // 省略(null)されたフィールドは既存値を保持する(誤って空にしないため)。
    if (request.title() != null) {
      material.setTitle(request.title());
    }
    if (request.content() != null) {
      material.setContent(request.content());
    }
    if (request.orderInCourse() != null) {
      material.setOrderInCourse(request.orderInCourse());
    }
    if (request.isPublished() != null) {
      material.setPublished(request.isPublished());
    }
    material.setUpdatedAt(Instant.now());

    return materials.save(material);
  }

  /** 教材を削除する。編集権が無ければ 403、未存在は 404。 */
  @Transactional
  public void delete(Long id, User actor) {
    TeachingMaterial material = findOrThrow(id);
    Course course = findCourseOrThrow(material.getCourseId());
    courseService.requireManage(course, actor);

    materials.delete(material);
  }

  private TeachingMaterial findOrThrow(Long id) {
    return materials
        .findById(id)
        .orElseThrow(() -> new ResponseStatusException(HttpStatus.NOT_FOUND, "material_not_found"));
  }

  private Course findCourseOrThrow(Long courseId) {
    return courses
        .findById(courseId)
        .orElseThrow(() -> new ResponseStatusException(HttpStatus.NOT_FOUND, "course_not_found"));
  }
}
