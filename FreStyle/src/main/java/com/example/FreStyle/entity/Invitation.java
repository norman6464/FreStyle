package com.example.FreStyle.entity;

import java.time.LocalDateTime;

import jakarta.persistence.Column;
import jakarta.persistence.Entity;
import jakarta.persistence.GeneratedValue;
import jakarta.persistence.GenerationType;
import jakarta.persistence.Id;
import jakarta.persistence.Table;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

/**
 * 招待トークン。
 * CompanyAdmin が Trainee を招待する際に作成される。
 */
@Entity
@Table(name = "invitations")
@Data
@NoArgsConstructor
@AllArgsConstructor
public class Invitation {

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @Column(name = "company_id", nullable = false)
    private Long companyId;

    @Column(nullable = false, length = 254)
    private String email;

    @Column(nullable = false, length = 30)
    private String role;

    @Column(nullable = false, length = 64, unique = true)
    private String token;

    @Column(name = "invited_by")
    private Long invitedBy;

    @Column(name = "expires_at", nullable = false)
    private LocalDateTime expiresAt;

    @Column(name = "accepted_at")
    private LocalDateTime acceptedAt;

    @Column(name = "accepted_user_id")
    private Long acceptedUserId;

    @Column(name = "created_at", insertable = false, updatable = false)
    private LocalDateTime createdAt;
}
