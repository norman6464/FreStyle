package com.example.FreStyle.repository;

import java.util.Optional;

import org.springframework.data.jpa.repository.JpaRepository;

import com.example.FreStyle.entity.SessionNote;

public interface SessionNoteRepository extends JpaRepository<SessionNote, Integer> {
    Optional<SessionNote> findByUserIdAndSessionId(Integer userId, Integer sessionId);
}
