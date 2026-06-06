package com.normanblog.frestyle.dto;

import com.normanblog.frestyle.entity.Course;
import java.time.Instant;

/** コースの API レスポンス表現。 */
public record CourseResponse(
    Long id,
    Long companyId,
    Long createdByUserId,
    String title,
    String description,
    int sortOrder,
    boolean isPublished,
    Instant createdAt,
    Instant updatedAt) {

  public static CourseResponse from(Course course) {
    return new CourseResponse(
        course.getId(),
        course.getCompanyId(),
        course.getCreatedByUserId(),
        course.getTitle(),
        course.getDescription(),
        course.getSortOrder(),
        course.isPublished(),
        course.getCreatedAt(),
        course.getUpdatedAt());
  }
}
