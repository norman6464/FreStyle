package com.example.FreStyle.repository;

import java.util.List;
import java.util.Optional;

import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import com.example.FreStyle.entity.Friendship;

@Repository
public interface FriendshipRepository extends JpaRepository<Friendship, Integer> {

    List<Friendship> findByFollowerIdOrderByCreatedAtDesc(Integer followerId);

    List<Friendship> findByFollowingIdOrderByCreatedAtDesc(Integer followingId);

    Optional<Friendship> findByFollowerIdAndFollowingId(Integer followerId, Integer followingId);

    boolean existsByFollowerIdAndFollowingId(Integer followerId, Integer followingId);

    long countByFollowerId(Integer followerId);

    long countByFollowingId(Integer followingId);
}
