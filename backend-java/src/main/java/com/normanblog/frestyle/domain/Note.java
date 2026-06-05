package com.normanblog.frestyle.domain;

import com.fasterxml.jackson.annotation.JsonAutoDetect;
import com.fasterxml.jackson.annotation.JsonProperty;
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

/** Note はユーザーの学習ノート。FreStyle(Go) の domain.Note と 1:1 対応。 */
// Lombok の boolean getter(isPublic()→"public" 等)が二重シリアライズされるのを防ぐため、
// JSON はフィールドのみから生成する(getter は無視)。フィールド名は @JsonProperty で Go 版に合わせる。
@JsonAutoDetect(
    fieldVisibility = JsonAutoDetect.Visibility.ANY,
    getterVisibility = JsonAutoDetect.Visibility.NONE,
    isGetterVisibility = JsonAutoDetect.Visibility.NONE)
@Entity
@Table(name = "notes")
@Getter
@Setter
@NoArgsConstructor
@AllArgsConstructor
@Builder
public class Note {

  @Id
  @GeneratedValue(strategy = GenerationType.IDENTITY)
  private Long id;

  @Column(name = "user_id")
  private Long userId;

  private String title;

  @Column(columnDefinition = "text")
  private String content;

  // Lombok の boolean getter は "is" を落として public/pinned になるため、
  // Go 版(json:"isPublic"/"isPinned")と互換にするよう明示する。
  @JsonProperty("isPublic")
  @Column(name = "is_public")
  private boolean isPublic;

  @JsonProperty("isPinned")
  @Column(name = "is_pinned")
  private boolean isPinned;

  @Column(name = "created_at")
  private Instant createdAt;

  @Column(name = "updated_at")
  private Instant updatedAt;
}
