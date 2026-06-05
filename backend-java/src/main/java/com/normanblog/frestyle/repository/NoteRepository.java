package com.normanblog.frestyle.repository;

import com.normanblog.frestyle.entity.Note;
import java.util.List;
import org.springframework.data.jpa.repository.JpaRepository;

/** notes テーブルへのアクセスを担うリポジトリ。 */
public interface NoteRepository extends JpaRepository<Note, Long> {

  List<Note> findByUserIdOrderByIsPinnedDescUpdatedAtDesc(Long userId);
}
