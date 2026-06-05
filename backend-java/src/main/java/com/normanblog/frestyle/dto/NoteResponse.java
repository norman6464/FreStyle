package com.normanblog.frestyle.dto;

import com.normanblog.frestyle.entity.Note;
import java.time.Instant;

/**
 * ノートの API レスポンス表現。
 *
 * <p>エンティティを直接シリアライズすると、DB の都合(列追加やリレーション)が API 契約に
 * 漏れたり、隠したいフィールドまで露出してしまう。公開する項目をこの record で固定し、
 * 永続化の構造と API の構造を独立して変更できるようにするため分離している。
 */
public record NoteResponse(
    Long id,
    Long userId,
    String title,
    String content,
    boolean isPublic,
    boolean isPinned,
    Instant createdAt,
    Instant updatedAt) {

  public static NoteResponse from(Note note) {
    return new NoteResponse(
        note.getId(),
        note.getUserId(),
        note.getTitle(),
        note.getContent(),
        note.isPublic(),
        note.isPinned(),
        note.getCreatedAt(),
        note.getUpdatedAt());
  }
}
