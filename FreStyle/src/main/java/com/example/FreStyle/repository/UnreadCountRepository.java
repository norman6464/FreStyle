package com.example.FreStyle.repository;

import com.example.FreStyle.entity.UnreadCount;
import com.example.FreStyle.entity.ChatRoom;
import com.example.FreStyle.entity.User;

import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.Optional;

@Repository
public interface UnreadCountRepository extends JpaRepository<UnreadCount, Integer> {

    // user と room による検索
    Optional<UnreadCount> findByUserAndRoom(User user, ChatRoom room);

    // count の合計を取得したいなどカスタムメソッドも追加可能
}
