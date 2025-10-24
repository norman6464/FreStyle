package com.example.FreStyle.repository;

import com.example.FreStyle.entity.ChatRoom;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

@Repository
public interface ChatRoomRepository extends JpaRepository<ChatRoom, Integer> {

    @Query("""
            SELECT rm1.room.id
            FROM RoomMember rm1
            JOIN RoomMember rm2 ON rm1.room.id = rm2.room.id
            WHERE rm1.user.id = :userId1 AND rm2.user.id = :userId2
            """)
    Integer findRoomIdByUserIds(@Param("userId1") Integer userId1, @Param("userId2") Integer userId2);
}
