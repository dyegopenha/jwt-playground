package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/dyegopenha/jwt-playground/internal/config/env"
	"github.com/dyegopenha/jwt-playground/internal/domain/entity"
	"github.com/dyegopenha/jwt-playground/internal/pkg/jwtutil"
	"github.com/dyegopenha/jwt-playground/internal/provider/cache"
)

type RefreshUseCase struct {
	c cache.Cache
	e *env.Env
	j *jwtutil.JWTUtil
}

func NewRefreshUseCase(
	c cache.Cache,
	e *env.Env,
	j *jwtutil.JWTUtil,
) *RefreshUseCase {
	return &RefreshUseCase{
		c: c,
		e: e,
		j: j,
	}
}

func (u *RefreshUseCase) Execute(
	ctx context.Context,
	refreshToken string,
) (accessToken string, newRefreshToken string, err error) {
	refreshSession := entity.RefreshSession{}
	ok, err := u.c.Scan(ctx, refreshToken, &refreshSession)
	if err != nil {
		return "", "", fmt.Errorf("failed to scan refresh token: %w", err)
	}
	if !ok || time.Now().After(refreshSession.ExpiresAt) {
		return "", "", errors.New("invalid or expired refresh token")
	}

	accessToken, newRefreshToken, err = u.j.IssueTokenPair(
		refreshSession.UserID,
		refreshSession.Role,
		u.e.AccessTokenTTL,
		u.e.RefreshTokenTTL,
	)
	if err != nil {
		return "", "", fmt.Errorf("failed to issue token pair: %w", err)
	}

	if err := u.c.Set(ctx, newRefreshToken, refreshSession, u.e.RefreshTokenTTL); err != nil {
		return "", "", fmt.Errorf("failed to set refresh token: %w", err)
	}

	if err := u.c.Delete(ctx, refreshToken); err != nil {
		return "", "", fmt.Errorf("failed to invalidate refresh token: %w", err)
	}

	return accessToken, newRefreshToken, nil
}
