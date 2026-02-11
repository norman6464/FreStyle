package com.example.FreStyle.repository;

import java.util.List;

import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import com.example.FreStyle.entity.CommunicationScore;

@Repository
public interface CommunicationScoreRepository extends JpaRepository<CommunicationScore, Integer> {

    List<CommunicationScore> findBySessionId(Integer sessionId);

    @Query("SELECT cs FROM CommunicationScore cs WHERE cs.user.id = :userId ORDER BY cs.createdAt DESC")
    List<CommunicationScore> findByUserIdOrderByCreatedAtDesc(@Param("userId") Integer userId);
}
