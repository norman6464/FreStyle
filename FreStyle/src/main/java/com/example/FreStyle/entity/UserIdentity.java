package com.example.FreStyle.entity;

import lombok.*;
import java.sql.Timestamp;
import jakarta.persistence.*;

@Entity
@Table(name = "user_identities",
       uniqueConstraints = {@UniqueConstraint(columnNames = {"provider", "provider_sub"})})
@Data
@NoArgsConstructor
@AllArgsConstructor
public class UserIdentity {

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Integer id;

    @ManyToOne(fetch = FetchType.LAZY)
    @JoinColumn(name = "user_id", nullable = false)
    private User user;

    @Column(length = 50, nullable = false)
    private String provider; // e.g., "password", "google-oidc", "cognito-oidc"

    @Column(name = "provider_sub", length = 255, nullable = false)
    private String providerSub;

    @Column(name = "created_at", insertable = false, updatable = false)
    private Timestamp createdAt;
}
