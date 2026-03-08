package com.example.FreStyle.repository;

import java.util.List;
import java.util.Optional;

import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.stereotype.Repository;

import com.example.FreStyle.entity.SharedSession;

@Repository
public interface SharedSessionRepository extends JpaRepository<SharedSession, Integer> {
    @Query("SELECT ss FROM SharedSession ss WHERE ss.isPublic = true ORDER BY ss.createdAt DESC")
    List<SharedSession> findPublicSessions();

    Optional<SharedSession> findBySessionId(Integer sessionId);

    List<SharedSession> findByUserIdOrderByCreatedAtDesc(Integer userId);
}
