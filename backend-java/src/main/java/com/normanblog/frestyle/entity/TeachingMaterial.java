package com.normanblog.frestyle.entity;

import jakarta.persistence.Column;
import jakarta.persistence.Entity;
import jakarta.persistence.GeneratedValue;
import jakarta.persistence.GenerationType;
import jakarta.persistence.Id;
import jakarta.persistence.Table;
import java.time.Instant;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Getter;
import lombok.NoArgsConstructor;
import lombok.Setter;

/** company_admin が作る教材。必ず 1 つの Course に所属する。本文は raw Markdown。 */
@Entity
@Table(name = "teaching_materials")
@Getter
@Setter
@NoArgsConstructor
@AllArgsConstructor
@Builder
public class TeachingMaterial {

  @Id
  @GeneratedValue(strategy = GenerationType.IDENTITY)
  private Long id;

  @Column(name = "company_id", nullable = false)
  private Long companyId;

  @Column(name = "course_id")
  private Long courseId;

  @Column(name = "created_by_user_id", nullable = false)
  private Long createdByUserId;

  private String title;

  @Column(columnDefinition = "text")
  private String content;

  // コース内の表示順(同値時は id 昇順)。
  @Column(name = "order_in_course")
  private int orderInCourse;

  @Column(name = "is_published")
  private boolean isPublished;

  @Column(name = "created_at")
  private Instant createdAt;

  @Column(name = "updated_at")
  private Instant updatedAt;
}
