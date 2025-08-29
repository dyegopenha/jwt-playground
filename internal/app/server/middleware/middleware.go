package middleware

import "github.com/dyegopenha/jwt-playground/internal/pkg/jwtutil"

type Middleware struct {
	j *jwtutil.JWTUtil
}

func NewMiddleware(j *jwtutil.JWTUtil) *Middleware {
	return &Middleware{
		j: j,
	}
}
