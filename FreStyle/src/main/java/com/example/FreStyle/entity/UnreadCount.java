package com.example.FreStyle.entity;

import jakarta.persistence.*;
import lombok.*;

@Entity
@Table(name = "unread_counts",
    uniqueConstraints = @UniqueConstraint(name = "uk_unread_count", columnNames = {"user_id", "room_id"})
)
@Data
@NoArgsConstructor
@AllArgsConstructor
public class UnreadCount {

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Integer id;

    
    // FetchType.Lazyは遅延読み込みであり、実際必要になるまで関連エンティティを読み込まないという制約になる
    @ManyToOne(fetch = FetchType.LAZY)
    @JoinColumn(name = "user_id", nullable = false)
    private User user;
    
    @ManyToOne(fetch = FetchType.LAZY)
    @JoinColumn(name = "room_id", nullable = false)
    private ChatRoom room;
    
    @Column(nullable = false)
    private Integer count = 0;
  }
  
  // FetchType.EAGER（即時読み込み）になる
  // 即時読み込みはセッションが続かなくても関連エンティティが読み込まれるのでServiceでの@Transactionalをつけなくてもエラーが出ない
  // ただし、パフォーマンスはLazyに劣る