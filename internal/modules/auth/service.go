package auth

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"gateway/webhook/internal/commons/async"
	apperr "gateway/webhook/internal/commons/errors"
	"gateway/webhook/internal/commons/hashes"
	"gateway/webhook/internal/commons/jwt"
	"gateway/webhook/internal/database"
	"gateway/webhook/internal/infrastructure/email"
	// "gateway/webhook/internal/modules/organizers"
	"gateway/webhook/internal/modules/users"
)

// Config carries the tunables the auth service needs. Kept as its own type
// so the service can be tested without a full app config.
type Config struct {
	AccessTokenTTL   time.Duration
	RefreshTokenTTL  time.Duration
	ResetTokenTTL    time.Duration // now 10 minutes
	ResetMaxAttempts int           // new — e.g. 5
	ResetURLBase     string        // keep, but now unused; can drop
}

type Service struct {
	db         *database.DB
	repo       Repository
	users      users.Repository
	issuer     *jwt.Issuer
	mailer     email.Sender
	runner     *async.Runner
	log        *slog.Logger
	cfg        Config
	now        func() time.Time // injectable for tests
}

func NewService(
	db *database.DB,
	repo Repository,
	usersRepo users.Repository,
	issuer *jwt.Issuer,
	mailer email.Sender,
	runner *async.Runner,
	log *slog.Logger,
	cfg Config,
) *Service {
	return &Service{
		db:         db,
		repo:       repo,
		users:      usersRepo,
		issuer:     issuer,
		mailer:     mailer,
		runner:     runner,
		log:        log,
		cfg:        cfg,
		now:        func() time.Time { return time.Now().UTC() },
	}
}

// ---------------------------------------------------------------- Register

func (s *Service) Register(ctx context.Context, in RegisterInput, ua, ip string) (*TokenPair, *users.User, error) {
	emailNorm := strings.ToLower(strings.TrimSpace(in.Email))

	// Pre-check for a friendly error. The DB unique index is the real guard.
	if _, err := s.users.FindByEmail(ctx, emailNorm); err == nil {
		return nil, nil, apperr.NewService("auth", apperr.KindConflict, "auth.email_taken", "email already registered", nil)
	} else if !isRepoNotFound(err) {
		return nil, nil, apperr.NewService("auth", apperr.KindInternal, "auth.lookup_failed", "could not check email", err)
	}

	pwHash, err := hashes.HashPassword(in.Password)
	if err != nil {
		return nil, nil, apperr.NewService("auth", apperr.KindInternal, "auth.hash_failed", "could not hash password", err)
	}

	u := &users.User{
		ID:           uuid.New(),
		Email:        emailNorm,
		PasswordHash: pwHash,
		Role:         users.Role(in.Role),
		Status:       users.StatusActive,
		FullName:     strings.TrimSpace(in.FullName),
		Phone:        strings.TrimSpace(in.Phone),
	}

	var pair *TokenPair
	err = database.WithTx(ctx, s.db, func(tx *gorm.DB) error {
		if err := s.users.Create(ctx, tx, u); err != nil {
			return err
		}

		// the users module is the sole owner of the users table.
		p, err := s.createSessionAndTokens(ctx, tx, u, ua, ip)
		if err != nil {
			return err
		}
		pair = p
		return nil
	})
	if err != nil {
		if isUniqueViolation(err) {
			return nil, nil, apperr.NewService("auth", apperr.KindConflict, "auth.email_taken", "email already registered", err)
		}
		return nil, nil, apperr.NewService("auth", apperr.KindInternal, "auth.register_failed", "could not register", err)
	}
	return pair, u, nil
}

// ------------------------------------------------------------------- Login

func (s *Service) Login(ctx context.Context, in LoginInput, ua, ip string) (*TokenPair, *users.User, error) {
	u, err := s.users.FindByEmail(ctx, strings.ToLower(strings.TrimSpace(in.Email)))
	if err != nil {
		// Return the same error for unknown email and wrong password —
		// don't leak which one is wrong.
		if isRepoNotFound(err) {
			return nil, nil, apperr.NewService("auth", apperr.KindUnauthorized, "auth.invalid_credentials", "invalid credentials", nil)
		}
		return nil, nil, apperr.NewService("auth", apperr.KindInternal, "auth.lookup_failed", "could not load user", err)
	}

	if u.Status == users.StatusSuspended {
		return nil, nil, apperr.NewService("auth", apperr.KindForbidden, "auth.suspended", "account suspended", nil)
	}
	if u.Status == users.StatusDeleted {
		return nil, nil, apperr.NewService("auth", apperr.KindUnauthorized, "auth.invalid_credentials", "invalid credentials", nil)
	}

	if err := hashes.ComparePassword(u.PasswordHash, in.Password); err != nil {
		return nil, nil, apperr.NewService("auth", apperr.KindUnauthorized, "auth.invalid_credentials", "invalid credentials", nil)
	}

	var pair *TokenPair
	err = database.WithTx(ctx, s.db, func(tx *gorm.DB) error {
		p, err := s.createSessionAndTokens(ctx, tx, u, ua, ip)
		if err != nil {
			return err
		}
		pair = p
		return nil
	})
	if err != nil {
		return nil, nil, apperr.NewService("auth", apperr.KindInternal, "auth.login_failed", "could not start session", err)
	}
	return pair, u, nil
}

// ----------------------------------------------------------------- Refresh

func (s *Service) Refresh(ctx context.Context, in RefreshInput) (*TokenPair, error) {
	hash := hashes.HashToken(in.RefreshToken)

	sess, err := s.repo.FindSessionByTokenHash(ctx, hash)
	if err != nil {
		if isRepoNotFound(err) {
			return nil, apperr.NewService("auth", apperr.KindUnauthorized, "auth.invalid_refresh", "invalid refresh token", nil)
		}
		return nil, apperr.NewService("auth", apperr.KindInternal, "auth.refresh_failed", "could not load session", err)
	}

	now := s.now()
	if sess.RevokedAt != nil {
		// Reuse of a rotated token. Best practice is to revoke the whole
		// session family; we do the session-scoped version for simplicity.
		return nil, apperr.NewService("auth", apperr.KindUnauthorized, "auth.revoked_refresh", "refresh token revoked", nil)
	}
	if now.After(sess.ExpiresAt) {
		return nil, apperr.NewService("auth", apperr.KindUnauthorized, "auth.expired_refresh", "refresh token expired", nil)
	}

	u, err := s.users.FindByID(ctx, sess.UserID)
	if err != nil {
		return nil, apperr.NewService("auth", apperr.KindInternal, "auth.lookup_failed", "could not load user", err)
	}
	if u.Status != users.StatusActive {
		return nil, apperr.NewService("auth", apperr.KindForbidden, "auth.account_inactive", "account is not active", nil)
	}

	var pair *TokenPair
	err = database.WithTx(ctx, s.db, func(tx *gorm.DB) error {
		// Revoke the old session first — this is the rotation.
		if err := tx.WithContext(ctx).
			Model(&Session{}).
			Where("id = ? AND revoked_at IS NULL", sess.ID).
			Update("revoked_at", now).Error; err != nil {
			return err
		}
		p, err := s.createSessionAndTokens(ctx, tx, u, sess.UserAgent, sess.IPAddress)
		if err != nil {
			return err
		}
		pair = p
		return nil
	})
	if err != nil {
		return nil, apperr.NewService("auth", apperr.KindInternal, "auth.refresh_failed", "could not rotate session", err)
	}
	return pair, nil
}

// ------------------------------------------------------------------ Logout

func (s *Service) Logout(ctx context.Context, in LogoutInput) error {
	hash := hashes.HashToken(in.RefreshToken)
	sess, err := s.repo.FindSessionByTokenHash(ctx, hash)
	if err != nil {
		// Idempotent: unknown token is a no-op success. Prevents probing.
		if isRepoNotFound(err) {
			return nil
		}
		return apperr.NewService("auth", apperr.KindInternal, "auth.logout_failed", "could not log out", err)
	}
	if err := s.repo.RevokeSession(ctx, sess.ID); err != nil {
		return apperr.NewService("auth", apperr.KindInternal, "auth.logout_failed", "could not log out", err)
	}
	return nil
}

// ------------------------------------------------------- Password reset

func (s *Service) RequestPasswordReset(ctx context.Context, in ForgotPasswordInput) error {
	u, err := s.users.FindByEmail(ctx, strings.ToLower(strings.TrimSpace(in.Email)))
	if err != nil {
		// Do not leak whether the email exists. Unknown email returns success
		// so an attacker can't enumerate accounts by watching the response.
		if isRepoNotFound(err) {
			return nil
		}
		return apperr.NewService("auth", apperr.KindInternal, "auth.reset_request_failed", "could not process request", err)
	}

	// Never let a suspended or deleted account trigger a reset. Return the
	// same generic success the "not found" case returns.
	if u.Status != users.StatusActive {
		return nil
	}

	code, err := hashes.NewNumericCode(6)
	if err != nil {
		return apperr.NewService("auth", apperr.KindInternal, "auth.reset_request_failed", "could not generate code", err)
	}
	codeHash := hashes.HashToken(code)

	// Invalidate any active reset row for this user, then insert the new one.
	// The partial unique index (prt_one_active_per_user) enforces this at the
	// DB level too, but doing it here makes the intent explicit and keeps the
	// error message meaningful if the index ever trips.
	err = database.WithTx(ctx, s.db, func(tx *gorm.DB) error {
		if err := s.repo.InvalidateActiveResetTokens(ctx, u.ID); err != nil {
			return err
		}
		row := &PasswordResetToken{
			ID:        uuid.New(),
			UserID:    u.ID,
			CodeHash:  codeHash,
			ExpiresAt: s.now().Add(s.cfg.ResetTokenTTL),
		}
		return s.repo.CreateResetToken(ctx, row)
	})
	if err != nil {
		return apperr.NewService("auth", apperr.KindInternal, "auth.reset_request_failed", "could not persist code", err)
	}

	body := "Your Webhook password reset code is: " + code +
		"\n\nIt expires in " + s.cfg.ResetTokenTTL.String() + "."

	// Fire-and-forget: the response does not depend on the email actually
	// being delivered. The token is already persisted, so a failed send is
	// recoverable by requesting a new code. The Runner guarantees a panic
	// inside the mailer cannot crash the process.
	bgCtx := async.Detach(ctx)
	s.runner.Go(bgCtx, "auth.password_reset_email", func(ctx context.Context) {
		if err := s.mailer.Send(ctx, u.Email, "Your Webhook password reset code", body); err != nil {
			s.log.WarnContext(ctx, "password_reset_email_failed",
				slog.String("user_id", u.ID.String()),
				slog.String("email", u.Email),
				slog.Any("err", err),
			)
		}
	})

	return nil
}

func (s *Service) ResetPassword(ctx context.Context, in ResetPasswordInput) error {
	emailNorm := strings.ToLower(strings.TrimSpace(in.Email))

	u, err := s.users.FindByEmail(ctx, emailNorm)
	if err != nil {
		if isRepoNotFound(err) {
			return apperr.NewService("auth", apperr.KindValidation, "auth.invalid_code", "invalid or expired code", nil)
		}
		return apperr.NewService("auth", apperr.KindInternal, "auth.reset_failed", "could not load user", err)
	}

	// 1. Find the active row for this user.
	row, err := s.repo.FindActiveResetTokenByUser(ctx, u.ID)
	if err != nil {
		if isRepoNotFound(err) {
			return apperr.NewService("auth", apperr.KindValidation, "auth.invalid_code", "invalid or expired code", nil)
		}
		return apperr.NewService("auth", apperr.KindInternal, "auth.reset_failed", "could not load token", err)
	}

	now := s.now()
	if now.After(row.ExpiresAt) {
		return apperr.NewService("auth", apperr.KindValidation, "auth.invalid_code", "invalid or expired code", nil)
	}

	// 2. Hard cap on attempts — check before comparing.
	if row.Attempts >= s.cfg.ResetMaxAttempts {
		_ = s.repo.InvalidateActiveResetTokens(ctx, u.ID)
		return apperr.NewService("auth", apperr.KindTooManyRequests, "auth.code_attempts_exhausted", "too many incorrect attempts; request a new code", nil)
	}

	// 3. Compare. On mismatch, bump attempts and return the same generic error.
	if row.CodeHash != hashes.HashToken(in.Code) {
		_ = s.repo.IncrementResetAttempts(ctx, row.ID)
		return apperr.NewService("auth", apperr.KindValidation, "auth.invalid_code", "invalid or expired code", nil)
	}

	// 4. Match — apply the change.
	newHash, err := hashes.HashPassword(in.NewPassword)
	if err != nil {
		return apperr.NewService("auth", apperr.KindInternal, "auth.reset_failed", "could not hash password", err)
	}
	u.PasswordHash = newHash

	return database.WithTx(ctx, s.db, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Save(u).Error; err != nil {
			return err
		}
		if err := tx.WithContext(ctx).
			Model(&PasswordResetToken{}).
			Where("id = ? AND used_at IS NULL", row.ID).
			Update("used_at", now).Error; err != nil {
			return err
		}
		return tx.WithContext(ctx).
			Model(&Session{}).
			Where("user_id = ? AND revoked_at IS NULL", u.ID).
			Update("revoked_at", now).Error
	})
}

// ---------------------------------------------------- internal helpers

// createSessionAndTokens does two things atomically: inserts a new session
// row with the hashed refresh token, and mints the access token.
// It must be called inside a transaction (tx non-nil).
func (s *Service) createSessionAndTokens(ctx context.Context, tx *gorm.DB, u *users.User, ua, ip string) (*TokenPair, error) {
	now := s.now()

	refresh, refreshHash, err := hashes.NewOpaqueToken(32)
	if err != nil {
		return nil, err
	}

	sess := &Session{
		ID:               uuid.New(),
		UserID:           u.ID,
		RefreshTokenHash: refreshHash,
		UserAgent:        truncate(ua, 255),
		IPAddress:        truncate(ip, 45),
		ExpiresAt:        now.Add(s.cfg.RefreshTokenTTL),
	}
	if err := tx.WithContext(ctx).Create(sess).Error; err != nil {
		return nil, err
	}

	access, exp, err := s.issuer.Issue(u.ID.String(), string(u.Role), sess.ID.String(), now)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  access,
		RefreshToken: refresh,
		TokenType:    "Bearer",
		ExpiresAt:    exp,
	}, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

func isRepoNotFound(err error) bool {
	var re *apperr.RepoError
	if errors.As(err, &re) {
		return re.Constraint == "not_found"
	}
	return false
}

func isUniqueViolation(err error) bool {
	// GORM's pgx driver exposes pgconn.PgError; we keep this string-based to
	// avoid a hard dependency on pgconn here.
	return err != nil && strings.Contains(err.Error(), "SQLSTATE 23505")
}
