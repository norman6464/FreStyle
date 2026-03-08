package com.example.FreStyle.repository;

import java.sql.Date;
import java.util.Optional;

import org.springframework.data.jpa.repository.JpaRepository;

import com.example.FreStyle.entity.WeeklyChallenge;

public interface WeeklyChallengeRepository extends JpaRepository<WeeklyChallenge, Integer> {
    Optional<WeeklyChallenge> findByWeekStartLessThanEqualAndWeekEndGreaterThanEqual(Date date, Date date2);
}
