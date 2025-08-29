//go:build wireinject
// +build wireinject

package server

import (
	"github.com/dyegopenha/jwt-playground/internal/app/server/handler"
	"github.com/dyegopenha/jwt-playground/internal/app/server/middleware"
	"github.com/dyegopenha/jwt-playground/internal/app/server/router"
	"github.com/dyegopenha/jwt-playground/internal/config/env"
	"github.com/dyegopenha/jwt-playground/internal/domain/usecase"
	"github.com/dyegopenha/jwt-playground/internal/pkg/jwtutil"
	"github.com/dyegopenha/jwt-playground/internal/pkg/validator"
	"github.com/dyegopenha/jwt-playground/internal/provider/cache"
	"github.com/dyegopenha/jwt-playground/internal/provider/cache/redis"
	"github.com/google/wire"
)

func New() *Server {
	wire.Build(
		env.NewEnv,

		jwtutil.NewJWTUtil,
		wire.Bind(new(validator.Validator), new(*validator.Validation)),
		validator.New,

		wire.Bind(new(cache.Cache), new(*redis.Redis)),
		redis.NewRedis,

		usecase.NewSignInUseCase,
		usecase.NewRefreshUseCase,

		middleware.NewMiddleware,

		handler.NewAuthHandler,
		handler.NewUserHandler,

		router.NewRouter,
		newServer,
	)

	return &Server{}
}
