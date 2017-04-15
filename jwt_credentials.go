package grpcutil

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	jwt "github.com/dgrijalva/jwt-go"

	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"

	"golang.org/x/net/context"
)

// JWTCredentials implements PerRPCCredentials for a jwt in the 'authorization' metadata field.
type JWTCredentials struct{}

var _ credentials.PerRPCCredentials = &JWTCredentials{}

var (
	defaultJWTSecret = "replacemereplacemereplacemereplacemereplace"
	defaultJWTEnvVar = "GRPCUTILS_JWT_SECRET"
)

func NewJWTCredentials() *JWTCredentials {
	return &JWTCredentials{}
}

var extractBearer = regexp.MustCompile("Bearer (.+)")

func jwtFromAuthorizationBearer(md metadata.MD, secret string) (*jwt.Token, error) {
	authn := md["Authorization"]
	if len(authn) == 0 {
		authn = md["authorization"]
	}
	if len(authn) != 1 {
		return nil, fmt.Errorf("unexpected number of parts in authorization header(%d)", len(authn))
	}
	jwtParts := extractBearer.FindStringSubmatch(authn[0])
	if len(jwtParts) != 2 {
		return nil, fmt.Errorf("unexpected number of parts in authorization header(%d)", len(jwtParts))
	}
	return jwt.Parse(jwtParts[1], func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
}

func (c *JWTCredentials) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	md, _ := metadata.FromContext(ctx)
	secret := os.Getenv(defaultJWTEnvVar)
	if secret == "" {
		secret = defaultJWTSecret
	}
	token, err := jwtFromAuthorizationBearer(md, secret)
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, errors.New("jwt: invalid token")
	}

	result := make(map[string]string)
	for k, v := range claims {
		val := ""
		switch v := v.(type) {
		case []interface{}:
			s := make([]string, 0)
			for _, item := range v {
				s = append(s, fmt.Sprint(item))
			}
			val = strings.Join(s, ",")
		default:
			val = fmt.Sprint(v)
		}
		result[k] = val
	}

	return result, err
}

func (c *JWTCredentials) RequireTransportSecurity() bool {
	return false
}
