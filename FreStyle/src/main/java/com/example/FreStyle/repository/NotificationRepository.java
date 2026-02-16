package com.example.FreStyle.repository;

import java.util.List;

import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import com.example.FreStyle.entity.Notification;

@Repository
public interface NotificationRepository extends JpaRepository<Notification, Integer> {

    List<Notification> findByUserIdOrderByCreatedAtDesc(Integer userId);

    List<Notification> findByUserIdAndIsReadFalseOrderByCreatedAtDesc(Integer userId);

    long countByUserIdAndIsReadFalse(Integer userId);
}
