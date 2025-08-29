package usecase

import (
	"context"
	"fmt"

	"github.com/dyegopenha/jwt-playground/internal/config/env"
	"github.com/dyegopenha/jwt-playground/internal/pkg/jwtutil"
	"github.com/dyegopenha/jwt-playground/internal/provider/cache"
)

type SignInUseCase struct {
	e *env.Env
	c cache.Cache
	j *jwtutil.JWTUtil
}

func NewSignInUseCase(
	e *env.Env,
	c cache.Cache,
	j *jwtutil.JWTUtil,
) *SignInUseCase {
	return &SignInUseCase{
		e: e,
		c: c,
		j: j,
	}
}

func (u *SignInUseCase) Execute(
	ctx context.Context,
	email, password string,
) (accessToken string, refreshToken string, err error) {
	// TODO: Get user from database and verify password
	user := map[string]string{
		"id":    "1",
		"email": "test@example.com",
		"role":  "admin",
	}

	accessToken, refreshToken, err = u.j.IssueTokenPair(
		user["id"],
		user["role"],
		u.e.AccessTokenTTL,
		u.e.RefreshTokenTTL,
	)
	if err != nil {
		return "", "", fmt.Errorf("failed to issue token pair: %w", err)
	}

	if err := u.c.Set(ctx, refreshToken, user, u.e.RefreshTokenTTL); err != nil {
		return "", "", fmt.Errorf("failed to set refresh token: %w", err)
	}

	return accessToken, refreshToken, nil
}
