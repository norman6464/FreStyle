package com.normanblog.frestyle.repository;

import com.normanblog.frestyle.domain.Note;
import java.util.List;
import org.springframework.data.jpa.repository.JpaRepository;

/**
 * NoteRepository は notes テーブルへのアクセス境界(port)。
 * Spring Data JPA が実装(adapter)を自動生成する。FreStyle(Go) の
 * usecase/repository.NoteRepository に対応。
 */
public interface NoteRepository extends JpaRepository<Note, Long> {

  List<Note> findByUserIdOrderByIsPinnedDescUpdatedAtDesc(Long userId);
}
