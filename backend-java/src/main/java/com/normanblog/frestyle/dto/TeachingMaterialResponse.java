package com.normanblog.frestyle.dto;

import com.normanblog.frestyle.entity.TeachingMaterial;
import java.time.Instant;

/** 教材の API レスポンス表現。 */
public record TeachingMaterialResponse(
    Long id,
    Long companyId,
    Long courseId,
    Long createdByUserId,
    String title,
    String content,
    int orderInCourse,
    boolean isPublished,
    Instant createdAt,
    Instant updatedAt) {

  public static TeachingMaterialResponse from(TeachingMaterial material) {
    return new TeachingMaterialResponse(
        material.getId(),
        material.getCompanyId(),
        material.getCourseId(),
        material.getCreatedByUserId(),
        material.getTitle(),
        material.getContent(),
        material.getOrderInCourse(),
        material.isPublished(),
        material.getCreatedAt(),
        material.getUpdatedAt());
  }
}
