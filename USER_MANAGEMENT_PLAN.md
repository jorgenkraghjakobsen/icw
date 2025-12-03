# ICW User Management & Password Reset - Next Level

## Current State Analysis

### Existing MAW User Management

**Current capabilities** (from icw-maw):
```go
AddUserToRepo(repo, user, password)      // Add user with password
RemoveUserFromRepo(repo, user)           // Remove user
ListUsers()                              // List all users
ListRepoUsers(repo)                      // List users per repo
```

**Current limitations**:
- ❌ No self-service password reset
- ❌ No email verification
- ❌ No user roles/permissions
- ❌ Passwords stored in SASL DB (no visibility)
- ❌ No invitation system
- ❌ No audit logging
- ❌ No password policies

## Vision: Next-Level User Management

```
┌────────────────────────────────────────────────────────────┐
│           Modern User Management System                     │
├────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌─────────────────────┐      ┌─────────────────────┐     │
│  │  User Portal        │      │  Admin Portal       │     │
│  │  ───────────        │      │  ────────────       │     │
│  │  • Profile          │      │  • User Management  │     │
│  │  • Reset Password   │      │  • Repo Access      │     │
│  │  • 2FA Setup        │      │  • Audit Logs       │     │
│  │  • API Keys         │      │  • Reports          │     │
│  └──────────┬──────────┘      └──────────┬──────────┘     │
│             │                            │                 │
│             └────────────┬───────────────┘                 │
│                          v                                 │
│              ┌──────────────────────┐                      │
│              │  Auth Service        │                      │
│              │  ────────────        │                      │
│              │  • JWT Tokens        │                      │
│              │  • Password Hashing  │                      │
│              │  • Email Service     │                      │
│              │  • 2FA (TOTP)        │                      │
│              │  • Audit Log         │                      │
│              └──────────┬───────────┘                      │
│                         │                                  │
│            ┌────────────┴────────────┐                     │
│            v                         v                     │
│  ┌─────────────────┐      ┌─────────────────┐            │
│  │  User Database  │      │  SASL/SVN Auth  │            │
│  │  PostgreSQL     │      │  (existing)     │            │
│  └─────────────────┘      └─────────────────┘            │
│                                                             │
└────────────────────────────────────────────────────────────┘
```

## Feature Breakdown

### Level 1: Self-Service Password Reset (Essential)

**User Flow**:
```
1. User visits: https://g9.example.com/icw/reset-password
2. Enters email/username
3. Receives reset link via email
4. Clicks link (token-based, 1-hour expiry)
5. Sets new password
6. Password synced to SASL DB
```

**Implementation**:
```go
type PasswordResetToken struct {
    ID        string
    UserID    string
    Email     string
    Token     string    // Random secure token
    ExpiresAt time.Time // 1 hour from creation
    Used      bool
}

func RequestPasswordReset(email string) error {
    // 1. Find user
    user := findUserByEmail(email)

    // 2. Generate secure token
    token := generateSecureToken()

    // 3. Store in database
    saveResetToken(user, token, time.Now().Add(1*time.Hour))

    // 4. Send email
    resetURL := fmt.Sprintf("https://g9/icw/reset/%s", token)
    sendEmail(user.Email, "Password Reset", resetURL)

    return nil
}

func ResetPassword(token, newPassword string) error {
    // 1. Validate token
    resetToken := findResetToken(token)
    if resetToken.Used || time.Now().After(resetToken.ExpiresAt) {
        return errors.New("invalid or expired token")
    }

    // 2. Validate password strength
    if !isStrongPassword(newPassword) {
        return errors.New("password too weak")
    }

    // 3. Update SASL DB (sync with existing auth)
    updateSASLPassword(resetToken.UserID, newPassword)

    // 4. Mark token as used
    markTokenUsed(token)

    // 5. Log the action
    auditLog("password_reset", resetToken.UserID)

    return nil
}
```

**Database Schema**:
```sql
CREATE TABLE users (
    id              UUID PRIMARY KEY,
    username        VARCHAR(50) UNIQUE NOT NULL,
    email           VARCHAR(255) UNIQUE NOT NULL,
    email_verified  BOOLEAN DEFAULT FALSE,
    created_at      TIMESTAMP DEFAULT NOW(),
    updated_at      TIMESTAMP DEFAULT NOW()
);

CREATE TABLE password_reset_tokens (
    id              UUID PRIMARY KEY,
    user_id         UUID REFERENCES users(id),
    token           VARCHAR(64) UNIQUE NOT NULL,
    expires_at      TIMESTAMP NOT NULL,
    used            BOOLEAN DEFAULT FALSE,
    created_at      TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_reset_token ON password_reset_tokens(token) WHERE NOT used;
```

**Email Template**:
```html
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; }
        .button {
            background-color: #4CAF50;
            color: white;
            padding: 14px 20px;
            text-decoration: none;
            border-radius: 4px;
        }
    </style>
</head>
<body>
    <h2>Password Reset Request</h2>
    <p>Hi {{.Username}},</p>
    <p>You requested a password reset for your ICW account.</p>
    <p>Click the button below to reset your password:</p>
    <p>
        <a href="{{.ResetURL}}" class="button">Reset Password</a>
    </p>
    <p>This link expires in 1 hour.</p>
    <p>If you didn't request this, please ignore this email.</p>
    <p>- ICW Team</p>
</body>
</html>
```

**Effort**: ~3-4 days

### Level 2: User Invitation System (High Value)

**Admin Flow**:
```
1. Admin invites user: email + repos they'll access
2. System sends invitation email
3. User clicks invite link
4. User creates account + sets password
5. Automatic access to assigned repos
```

**Implementation**:
```go
type Invitation struct {
    ID          string
    Email       string
    InvitedBy   string
    Repos       []string  // Repositories they'll have access to
    Role        UserRole  // Admin, Developer, ReadOnly
    Token       string
    ExpiresAt   time.Time // 7 days
    Accepted    bool
}

func InviteUser(email string, repos []string, role UserRole, invitedBy string) error {
    // 1. Check if user exists
    if userExists(email) {
        return errors.New("user already exists")
    }

    // 2. Create invitation
    invitation := Invitation{
        ID:        uuid.New().String(),
        Email:     email,
        Repos:     repos,
        Role:      role,
        Token:     generateSecureToken(),
        InvitedBy: invitedBy,
        ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
    }

    // 3. Save invitation
    saveInvitation(invitation)

    // 4. Send email
    inviteURL := fmt.Sprintf("https://g9/icw/accept-invite/%s", invitation.Token)
    sendInviteEmail(email, inviteURL, repos)

    return nil
}

func AcceptInvitation(token, username, password string) error {
    // 1. Validate invitation
    invite := findInvitation(token)
    if invite.Accepted || time.Now().After(invite.ExpiresAt) {
        return errors.New("invalid or expired invitation")
    }

    // 2. Create user
    user := createUser(username, invite.Email, password)

    // 3. Add to all invited repos
    for _, repo := range invite.Repos {
        backend.AddUserToRepo(repo, username, password)
    }

    // 4. Mark invitation accepted
    markInvitationAccepted(token, user.ID)

    // 5. Send welcome email
    sendWelcomeEmail(user)

    return nil
}
```

**Database Schema**:
```sql
CREATE TABLE invitations (
    id              UUID PRIMARY KEY,
    email           VARCHAR(255) NOT NULL,
    invited_by      UUID REFERENCES users(id),
    role            VARCHAR(20) NOT NULL,
    token           VARCHAR(64) UNIQUE NOT NULL,
    expires_at      TIMESTAMP NOT NULL,
    accepted        BOOLEAN DEFAULT FALSE,
    accepted_by     UUID REFERENCES users(id),
    accepted_at     TIMESTAMP,
    created_at      TIMESTAMP DEFAULT NOW()
);

CREATE TABLE invitation_repos (
    invitation_id   UUID REFERENCES invitations(id),
    repo_name       VARCHAR(100) NOT NULL,
    PRIMARY KEY (invitation_id, repo_name)
);
```

**Effort**: ~4-5 days

### Level 3: Role-Based Access Control (RBAC)

**Roles**:
```go
type UserRole string

const (
    RoleAdmin      UserRole = "admin"       // Full access to everything
    RoleDeveloper  UserRole = "developer"   // Read/Write to assigned repos
    RoleReadOnly   UserRole = "readonly"    // Read-only access
    RoleGuest      UserRole = "guest"       // Limited, temporary access
)

type Permission string

const (
    PermRepoCreate   Permission = "repo:create"
    PermRepoDelete   Permission = "repo:delete"
    PermRepoRead     Permission = "repo:read"
    PermRepoWrite    Permission = "repo:write"
    PermUserManage   Permission = "user:manage"
    PermUserInvite   Permission = "user:invite"
)

var RolePermissions = map[UserRole][]Permission{
    RoleAdmin: {
        PermRepoCreate, PermRepoDelete, PermRepoRead, PermRepoWrite,
        PermUserManage, PermUserInvite,
    },
    RoleDeveloper: {
        PermRepoRead, PermRepoWrite,
    },
    RoleReadOnly: {
        PermRepoRead,
    },
    RoleGuest: {
        PermRepoRead, // Temporary, specific repos only
    },
}
```

**Database Schema**:
```sql
CREATE TABLE user_roles (
    user_id         UUID REFERENCES users(id),
    repo_name       VARCHAR(100),
    role            VARCHAR(20) NOT NULL,
    granted_by      UUID REFERENCES users(id),
    granted_at      TIMESTAMP DEFAULT NOW(),
    expires_at      TIMESTAMP, -- For guest access
    PRIMARY KEY (user_id, repo_name)
);
```

**Middleware**:
```go
func RequirePermission(perm Permission) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            user := getUserFromContext(r.Context())
            repo := getRepoFromRequest(r)

            if !user.HasPermission(perm, repo) {
                http.Error(w, "Forbidden", http.StatusForbidden)
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}

// Usage:
http.Handle("/repo/delete", RequirePermission(PermRepoDelete)(deleteRepoHandler))
```

**Effort**: ~5-6 days

### Level 4: Two-Factor Authentication (2FA)

**TOTP-based 2FA** (Google Authenticator compatible):

```go
type TwoFactorAuth struct {
    UserID      string
    Secret      string    // TOTP secret
    Enabled     bool
    BackupCodes []string  // 10 single-use backup codes
    CreatedAt   time.Time
}

func EnableTwoFactor(userID string) (*TwoFactorAuth, error) {
    // 1. Generate TOTP secret
    secret, err := totp.Generate(totp.GenerateOpts{
        Issuer:      "ICW",
        AccountName: getUserEmail(userID),
    })

    // 2. Generate backup codes
    backupCodes := generateBackupCodes(10)

    // 3. Store (disabled until verified)
    tfa := &TwoFactorAuth{
        UserID:      userID,
        Secret:      secret.Secret(),
        Enabled:     false,
        BackupCodes: hashBackupCodes(backupCodes),
    }
    saveTwoFactorAuth(tfa)

    // 4. Return QR code and backup codes
    return tfa, nil
}

func VerifyTwoFactorSetup(userID, code string) error {
    tfa := getTwoFactorAuth(userID)

    // Verify TOTP code
    if !totp.Validate(code, tfa.Secret) {
        return errors.New("invalid code")
    }

    // Enable 2FA
    tfa.Enabled = true
    updateTwoFactorAuth(tfa)

    auditLog("2fa_enabled", userID)
    return nil
}

func LoginWithTwoFactor(username, password, totpCode string) (string, error) {
    // 1. Verify username/password
    user := authenticateUser(username, password)
    if user == nil {
        return "", errors.New("invalid credentials")
    }

    // 2. Check if 2FA enabled
    tfa := getTwoFactorAuth(user.ID)
    if !tfa.Enabled {
        // No 2FA, return JWT
        return generateJWT(user), nil
    }

    // 3. Verify TOTP code
    if !totp.Validate(totpCode, tfa.Secret) {
        // Check backup codes
        if !verifyBackupCode(user.ID, totpCode) {
            return "", errors.New("invalid 2FA code")
        }
    }

    // 4. Return JWT
    return generateJWT(user), nil
}
```

**Database Schema**:
```sql
CREATE TABLE two_factor_auth (
    user_id         UUID PRIMARY KEY REFERENCES users(id),
    secret          VARCHAR(255) NOT NULL,
    enabled         BOOLEAN DEFAULT FALSE,
    created_at      TIMESTAMP DEFAULT NOW()
);

CREATE TABLE backup_codes (
    user_id         UUID REFERENCES users(id),
    code_hash       VARCHAR(64) NOT NULL,
    used            BOOLEAN DEFAULT FALSE,
    used_at         TIMESTAMP
);
```

**Effort**: ~4-5 days

### Level 5: API Keys for Automation

**For CI/CD and automated access**:

```go
type APIKey struct {
    ID          string
    UserID      string
    Name        string    // "Jenkins CI", "GitHub Actions"
    Key         string    // Hashed
    KeyPrefix   string    // First 8 chars (for display)
    Repos       []string  // Repos this key can access
    Permissions []Permission
    LastUsedAt  time.Time
    ExpiresAt   time.Time
    CreatedAt   time.Time
}

func CreateAPIKey(userID, name string, repos []string, expiresIn time.Duration) (*APIKey, error) {
    // 1. Generate key
    rawKey := generateAPIKey() // e.g., "icw_aBcDeFgHiJkLmNoPqRsTuVwXyZ123456"
    keyHash := hashAPIKey(rawKey)

    // 2. Create API key
    apiKey := &APIKey{
        ID:          uuid.New().String(),
        UserID:      userID,
        Name:        name,
        Key:         keyHash,
        KeyPrefix:   rawKey[:8],
        Repos:       repos,
        Permissions: []Permission{PermRepoRead, PermRepoWrite},
        ExpiresAt:   time.Now().Add(expiresIn),
        CreatedAt:   time.Now(),
    }

    saveAPIKey(apiKey)
    auditLog("api_key_created", userID, map[string]interface{}{"key_name": name})

    // 3. Return raw key (only shown once!)
    apiKey.Key = rawKey
    return apiKey, nil
}

func AuthenticateAPIKey(keyString string) (*User, error) {
    keyHash := hashAPIKey(keyString)
    apiKey := findAPIKeyByHash(keyHash)

    if apiKey == nil || time.Now().After(apiKey.ExpiresAt) {
        return nil, errors.New("invalid or expired API key")
    }

    // Update last used
    apiKey.LastUsedAt = time.Now()
    updateAPIKey(apiKey)

    // Return user with limited permissions
    user := getUser(apiKey.UserID)
    user.APIKeyPermissions = apiKey.Permissions
    user.APIKeyRepos = apiKey.Repos

    return user, nil
}
```

**Usage**:
```bash
# CI/CD usage
export ICW_API_KEY="icw_aBcDeFgHiJkLmNoPqRsTuVwXyZ123456"
icw checkout digital/top --api-key=$ICW_API_KEY

# Or via auth header
curl -H "Authorization: Bearer icw_aBcDeFgHiJkLmNoPqRsTuVwXyZ123456" \
     https://g9/icw/api/repos/cp4/components
```

**Effort**: ~3-4 days

### Level 6: Audit Logging

**Track everything**:

```go
type AuditLog struct {
    ID          string
    UserID      string
    Action      string    // "login", "password_reset", "repo_create", etc.
    Resource    string    // "repo:cp4", "user:john"
    IPAddress   string
    UserAgent   string
    Success     bool
    ErrorMsg    string
    Metadata    map[string]interface{}
    Timestamp   time.Time
}

func AuditLog(action, userID, resource string, success bool, metadata map[string]interface{}) {
    log := AuditLog{
        ID:        uuid.New().String(),
        UserID:    userID,
        Action:    action,
        Resource:  resource,
        Success:   success,
        Metadata:  metadata,
        Timestamp: time.Now(),
    }

    saveAuditLog(log)

    // Also log to file for compliance
    logToFile(log)
}

// Usage throughout code:
func LoginHandler(w http.ResponseWriter, r *http.Request) {
    user, err := authenticate(username, password)

    if err != nil {
        AuditLog("login_failed", username, "", false, map[string]interface{}{
            "ip":         r.RemoteAddr,
            "user_agent": r.UserAgent(),
            "reason":     err.Error(),
        })
        http.Error(w, "Invalid credentials", 401)
        return
    }

    AuditLog("login_success", user.ID, "", true, map[string]interface{}{
        "ip":         r.RemoteAddr,
        "user_agent": r.UserAgent(),
    })

    // ... proceed with login
}
```

**Database Schema**:
```sql
CREATE TABLE audit_logs (
    id              UUID PRIMARY KEY,
    user_id         UUID,
    action          VARCHAR(50) NOT NULL,
    resource        VARCHAR(255),
    ip_address      INET,
    user_agent      TEXT,
    success         BOOLEAN NOT NULL,
    error_message   TEXT,
    metadata        JSONB,
    timestamp       TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_audit_user ON audit_logs(user_id, timestamp);
CREATE INDEX idx_audit_action ON audit_logs(action, timestamp);
CREATE INDEX idx_audit_resource ON audit_logs(resource, timestamp);
```

**Admin Dashboard**:
```
Recent Activity:
─────────────────────────────────────────────────────────
[2024-12-03 14:23] john@example.com    login_success
[2024-12-03 14:25] john@example.com    repo_checkout    cp4/digital/top
[2024-12-03 14:30] admin@example.com   user_invited     jane@example.com
[2024-12-03 14:45] jane@example.com    invite_accepted
[2024-12-03 15:00] john@example.com    password_reset
```

**Effort**: ~2-3 days

### Level 7: Email Service Integration

**SMTP Configuration**:

```go
type EmailConfig struct {
    SMTPHost     string // smtp.gmail.com
    SMTPPort     int    // 587
    Username     string
    Password     string
    FromAddress  string
    FromName     string
}

type EmailService struct {
    config EmailConfig
    smtp   *gomail.Dialer
}

func NewEmailService(config EmailConfig) *EmailService {
    return &EmailService{
        config: config,
        smtp:   gomail.NewDialer(config.SMTPHost, config.SMTPPort, config.Username, config.Password),
    }
}

func (s *EmailService) SendPasswordReset(to, resetURL string) error {
    tmpl := loadTemplate("password_reset.html")

    body, err := executeTemplate(tmpl, map[string]interface{}{
        "ResetURL": resetURL,
    })

    return s.sendEmail(to, "Password Reset Request", body)
}

func (s *EmailService) SendInvitation(to, inviteURL string, repos []string) error {
    tmpl := loadTemplate("invitation.html")

    body, err := executeTemplate(tmpl, map[string]interface{}{
        "InviteURL": inviteURL,
        "Repos":     repos,
    })

    return s.sendEmail(to, "You've been invited to ICW", body)
}

func (s *EmailService) sendEmail(to, subject, body string) error {
    m := gomail.NewMessage()
    m.SetHeader("From", fmt.Sprintf("%s <%s>", s.config.FromName, s.config.FromAddress))
    m.SetHeader("To", to)
    m.SetHeader("Subject", subject)
    m.SetBody("text/html", body)

    return s.smtp.DialAndSend(m)
}
```

**Environment Configuration**:
```bash
# .env
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=icw@yourcompany.com
SMTP_PASSWORD=your-app-password
EMAIL_FROM=icw@yourcompany.com
EMAIL_FROM_NAME=ICW System
```

**Effort**: ~2 days

## Architecture Overview

### Technology Stack

```
Frontend:
- HTML/CSS/JavaScript (or React for advanced UI)
- Bootstrap 5 for styling
- Chart.js for admin dashboard

Backend:
- Go (extends existing MAW server)
- PostgreSQL for user data
- Existing SASL DB for SVN auth (sync)

Libraries:
- JWT: github.com/golang-jwt/jwt
- TOTP: github.com/pquerna/otp
- Email: gopkg.in/gomail.v2
- Password: golang.org/x/crypto/bcrypt
- Database: github.com/lib/pq
```

### Database Migration Plan

**Phase 1**: Add PostgreSQL alongside SASL
```
SASL DB (existing)     PostgreSQL (new)
─────────────────      ──────────────────
SVN passwords    <─┐   Users table
                   └──> • username
                        • email
                        • metadata
                        • roles
```

**Phase 2**: Sync passwords both ways
- User sets password → update both DBs
- Keep SASL as source of truth for SVN
- PostgreSQL for web auth & metadata

### Security Considerations

1. **Password Storage**:
   - PostgreSQL: bcrypt (cost 12)
   - SASL: Keep existing format
   - Sync on change

2. **Tokens**:
   - JWT with RS256 (asymmetric)
   - Short-lived (15 min)
   - Refresh tokens (7 days)

3. **Rate Limiting**:
   - Login attempts: 5 per 15 min
   - Password reset: 3 per hour
   - API calls: 1000 per hour

4. **HTTPS Only**:
   - Force HTTPS for all web traffic
   - HSTS headers
   - Secure cookies

## Implementation Roadmap

### Sprint 1: Foundation (Week 1-2)
- [ ] PostgreSQL setup
- [ ] User database schema
- [ ] Email service integration
- [ ] Basic JWT authentication
- [ ] Password reset flow

**Deliverable**: Users can reset passwords

### Sprint 2: Invitations (Week 3)
- [ ] Invitation system
- [ ] Invitation emails
- [ ] Accept invitation flow
- [ ] Auto-repo assignment

**Deliverable**: Admins can invite users

### Sprint 3: RBAC (Week 4)
- [ ] Role system
- [ ] Permission middleware
- [ ] Admin dashboard
- [ ] User management UI

**Deliverable**: Role-based access control

### Sprint 4: Security (Week 5-6)
- [ ] 2FA/TOTP implementation
- [ ] Backup codes
- [ ] API keys
- [ ] Rate limiting

**Deliverable**: Enhanced security features

### Sprint 5: Audit & Polish (Week 7)
- [ ] Audit logging
- [ ] Admin reports
- [ ] Email templates
- [ ] Documentation
- [ ] Testing

**Deliverable**: Production-ready system

## Effort Summary

| Feature                    | Effort      | Priority |
|----------------------------|-------------|----------|
| Password Reset             | 3-4 days    | Critical |
| User Invitations           | 4-5 days    | High     |
| RBAC                       | 5-6 days    | High     |
| 2FA/TOTP                   | 4-5 days    | Medium   |
| API Keys                   | 3-4 days    | Medium   |
| Audit Logging              | 2-3 days    | Medium   |
| Email Service              | 2 days      | Critical |
| Admin Dashboard            | 5-7 days    | High     |
| Testing & Documentation    | 3-5 days    | High     |
| **Total**                  | **~7 weeks**| -        |

## MVP: What to Build First?

**Minimum Viable Product (2-3 weeks)**:
1. ✅ Password Reset (self-service)
2. ✅ User Invitations (email-based)
3. ✅ Basic Role System (admin/user)
4. ✅ Email Service
5. ✅ Simple Admin UI

**Later Enhancements**:
- 2FA (add when needed)
- API Keys (add when CI/CD needed)
- Advanced RBAC (add if required)
- Audit Logging (add for compliance)

## Cost Analysis

### Infrastructure
- **PostgreSQL**: Free (self-hosted on g9)
- **Email Service**:
  - Gmail/G Suite: Free for low volume
  - SendGrid: $15/month for 40k emails
  - AWS SES: $0.10 per 1000 emails

### Development Time
- **MVP**: 2-3 weeks (1 developer)
- **Full System**: 7 weeks (1 developer)
- **With 2 developers**: 4 weeks

## Success Metrics

- ✅ Users can reset passwords themselves (no admin intervention)
- ✅ New user onboarding < 5 minutes
- ✅ Admin can manage 100+ users efficiently
- ✅ Zero password-related security incidents
- ✅ Audit trail for compliance
- ✅ 99.9% email delivery rate

## Next Steps

1. **Review this plan** - Agree on scope (MVP vs Full)
2. **Setup PostgreSQL** on g9
3. **Configure email service** (SMTP credentials)
4. **Start Sprint 1** - Password reset
5. **Iterate** - Demo after each sprint

## Questions?

1. **MVP or Full system?** - What's the timeline?
2. **Email service preference?** - Gmail, SendGrid, or AWS SES?
3. **2FA requirement?** - Is this needed for compliance?
4. **Single developer or team?** - Affects timeline
5. **Existing Active Directory/LDAP?** - Could integrate instead

---

Ready to start? Let me know which features are highest priority and we can begin implementation!
