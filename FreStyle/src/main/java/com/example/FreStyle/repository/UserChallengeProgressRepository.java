package com.example.FreStyle.repository;

import java.util.Optional;

import org.springframework.data.jpa.repository.JpaRepository;

import com.example.FreStyle.entity.UserChallengeProgress;

public interface UserChallengeProgressRepository extends JpaRepository<UserChallengeProgress, Integer> {
    Optional<UserChallengeProgress> findByUserIdAndChallengeId(Integer userId, Integer challengeId);
}
