package com.normanblog.frestyle.service;

import com.normanblog.frestyle.entity.Course;
import com.normanblog.frestyle.entity.Role;
import com.normanblog.frestyle.entity.TeachingMaterial;
import com.normanblog.frestyle.entity.User;
import com.normanblog.frestyle.repository.CourseRepository;
import com.normanblog.frestyle.repository.TeachingMaterialRepository;
import java.util.List;
import org.springframework.http.HttpStatus;
import org.springframework.stereotype.Service;
import org.springframework.web.server.ResponseStatusException;

/**
 * コースと、コース配下の教材一覧の閲覧を担うサービス。
 *
 * <p>アクセス制御はここで actor の company / role を見て行う。trainee は自社の published のみ、
 * 管理者(company_admin / super_admin)は draft も見られる。super_admin は会社を跨いで閲覧可。
 */
@Service
public class CourseService {

  private final CourseRepository courses;
  private final TeachingMaterialRepository materials;

  public CourseService(CourseRepository courses, TeachingMaterialRepository materials) {
    this.courses = courses;
    this.materials = materials;
  }

  /** actor の会社のコース一覧。trainee は published のみ。会社未所属なら空。 */
  public List<Course> list(User actor) {
    if (actor.getCompanyId() == null) {
      return List.of();
    }

    Long companyId = actor.getCompanyId();
    if (isManager(actor.getRole())) {
      return courses.findByCompanyIdOrderBySortOrderAscIdAsc(companyId);
    }

    return courses.findByCompanyIdAndIsPublishedTrueOrderBySortOrderAscIdAsc(companyId);
  }

  /** コース詳細。閲覧権が無ければ 403、存在しなければ 404。 */
  public Course get(Long id, User actor) {
    Course course = findOrThrow(id);
    requireReadable(course, actor);

    return course;
  }

  /** コース配下の教材一覧。コースの閲覧権を確認し、trainee には published のみ返す。 */
  public List<TeachingMaterial> listMaterials(Long courseId, User actor) {
    Course course = findOrThrow(courseId);
    requireReadable(course, actor);

    if (isManager(actor.getRole())) {
      return materials.findByCourseIdOrderByOrderInCourseAscIdAsc(courseId);
    }

    return materials.findByCourseIdAndIsPublishedTrueOrderByOrderInCourseAscIdAsc(courseId);
  }

  private Course findOrThrow(Long id) {
    return courses
        .findById(id)
        .orElseThrow(() -> new ResponseStatusException(HttpStatus.NOT_FOUND, "course_not_found"));
  }

  void requireReadable(Course course, User actor) {
    if (!canRead(course, actor)) {
      throw new ResponseStatusException(HttpStatus.FORBIDDEN, "forbidden");
    }
  }

  // super_admin は常に可。他社のコースは不可。未公開は管理者のみ可。
  boolean canRead(Course course, User actor) {
    if (Role.SUPER_ADMIN.equals(actor.getRole())) {
      return true;
    }
    if (!course.getCompanyId().equals(actor.getCompanyId())) {
      return false;
    }

    return course.isPublished() || isManager(actor.getRole());
  }

  static boolean isManager(String role) {
    return Role.COMPANY_ADMIN.equals(role) || Role.SUPER_ADMIN.equals(role);
  }
}
