package com.example.FreStyle.repository;

import com.example.FreStyle.dto.PartnerRoomProjection;
import com.example.FreStyle.entity.RoomMember;
import com.example.FreStyle.entity.User;

import java.util.List;
import java.util.Optional;

import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

@Repository
public interface RoomMemberRepository extends JpaRepository<RoomMember, Integer> {

    /*
     * JPAでは外部キーでは_をつけてプロパティ名を書く必要がある
     */

    // ルームとユーザーの組み合わせが存在する（ユニーク制約チェック）
    boolean existsByRoom_IdAndUser_Id(Integer roomId, Integer userId);

    // ユーザーが所属する全チャットルームのIDを取得
    @Query("SELECT rm.room.id FROM RoomMember rm WHERE rm.user.id = :userId")
    List<Integer> findRoomIdByUserId(@Param("userId") Integer userId);

    // 当該ユーザー以外の参加しているルームを検索
    @Query("""
                SELECT rm2.user
                FROM RoomMember rm2
                WHERE rm2.room.id IN (
                    SELECT rm.room.id
                    FROM RoomMember rm
                    WHERE rm.user.id = :userId
                )
                AND rm2.user.id <> :userId
            """)
    List<User> findUsersByUserId(@Param("userId") Integer userId);

    /**
     * 自分が参加しているルームで、相手ユーザーとルームIDのペアを取得
     */
    @Query("""
            SELECT rm2.user.id AS userId, rm2.room.id AS roomId
            FROM RoomMember rm2
            WHERE rm2.room.id IN (
                SELECT rm.room.id FROM RoomMember rm WHERE rm.user.id = :userId
            )
            AND rm2.user.id <> :userId
            """)
    List<PartnerRoomProjection> findPartnerUserIdAndRoomIdByUserId(@Param("userId") Integer userId);

    /**
     * 指定ルームで自分以外のユーザーを取得
     */
    @Query("SELECT rm.user FROM RoomMember rm WHERE rm.room.id = :roomId AND rm.user.id <> :userId")
    Optional<User> findPartnerByRoomIdAndUserId(@Param("roomId") Integer roomId, @Param("userId") Integer userId);

    /**
     * 自分が参加しているルームで、自分以外のユーザー数をカウント（= 会話したことのあるユーザー数）
     */
    @Query("""
            SELECT COUNT(DISTINCT rm2.user.id)
            FROM RoomMember rm2
            WHERE rm2.room.id IN (
                SELECT rm.room.id FROM RoomMember rm WHERE rm.user.id = :userId
            )
            AND rm2.user.id <> :userId
            """)
    Long countDistinctPartnersByUserId(@Param("userId") Integer userId);

}
