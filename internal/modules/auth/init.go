package auth

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"

	"gateway/webhook/internal/commons/async"
	"gateway/webhook/internal/commons/jwt"
	"gateway/webhook/internal/database"
	"gateway/webhook/internal/infrastructure/email"
	"gateway/webhook/internal/middleware"
	"gateway/webhook/internal/modules/users"
)

type InitConfig struct {
	AccessTokenTTL   time.Duration
	RefreshTokenTTL  time.Duration
	ResetTokenTTL    time.Duration
	ResetMaxAttempts int

	LoginRatePerWindow    int
	LoginWindow           time.Duration
	RegisterRatePerWindow int
	RegisterWindow        time.Duration
	RefreshRatePerWindow  int
	RefreshWindow         time.Duration
}

func Init(
	router *gin.RouterGroup,
	db *database.DB,
	log *slog.Logger,
	issuer *jwt.Issuer,
	usersRepo users.Repository,
	// organizersRepo organizers.Repository,
	mailer email.Sender,
	runner *async.Runner,
	cfg InitConfig,
	_ gin.HandlerFunc,
	_ gin.HandlerFunc,
) gin.HandlerFunc {
	repo := NewRepository(db.Primary)
	svc := NewService(db, repo, usersRepo, issuer, mailer, runner, log, Config{
		AccessTokenTTL:   cfg.AccessTokenTTL,
		RefreshTokenTTL:  cfg.RefreshTokenTTL,
		ResetTokenTTL:    cfg.ResetTokenTTL,
		ResetMaxAttempts: cfg.ResetMaxAttempts,
	})
	h := NewHandler(svc)

	loginLimiter := newLimiter(cfg.LoginRatePerWindow, cfg.LoginWindow)
	registerLimiter := newLimiter(cfg.RegisterRatePerWindow, cfg.RegisterWindow)
	refreshLimiter := newLimiter(cfg.RefreshRatePerWindow, cfg.RefreshWindow)

	rl := func(l *limiter) gin.HandlerFunc {
		return middleware.RateLimitByIP(l.Allow, "auth.too_many_requests", "too many requests")
	}

	authGroup := router.Group("/auth")
	{
		authGroup.POST("/register", rl(registerLimiter), h.Register)
		authGroup.POST("/login", rl(loginLimiter), h.Login)
		authGroup.POST("/refresh", rl(refreshLimiter), h.Refresh)
		authGroup.POST("/logout", h.Logout)
		authGroup.POST("/password/forgot", h.ForgotPassword)
		authGroup.POST("/password/reset", h.ResetPassword)
	}

	return RequireAuth(issuer)
}
