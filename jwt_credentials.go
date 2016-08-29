package grpcutils

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"

	"github.com/coreos/go-oidc/jose"

	"golang.org/x/net/context"
)

type JWTCredentials struct{}

var _ credentials.PerRPCCredentials = &JWTCredentials{}

var (
	defaultJWTSecret = "replacemereplacemereplacemereplacemereplace"
	defaultJWTEnvVar = "GRPCUTILS_JWT_SECRET"
)

var ErrJWTExpired = errors.New("jwt: expired")

func NewJWTCredentials() *JWTCredentials {
	return &JWTCredentials{}
}

var extractBearer = regexp.MustCompile("Bearer (.+)")

func jwtFromAuthorizationBearer(md metadata.MD) (jose.JWT, error) {
	authn := md["Authorization"]
	if len(authn) == 0 {
		authn = md["authorization"]
	}
	if len(authn) != 1 {
		return jose.JWT{}, fmt.Errorf("unexpected number of parts in authorization header(%d)", len(authn))
	}
	jwtParts := extractBearer.FindStringSubmatch(authn[0])
	if len(jwtParts) != 2 {
		return jose.JWT{}, fmt.Errorf("unexpected number of parts in authorization header(%d)", len(jwtParts))
	}
	return jose.ParseJWT(jwtParts[1])
}

func (c *JWTCredentials) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	md, _ := metadata.FromContext(ctx)
	jwt, err := jwtFromAuthorizationBearer(md)
	if err != nil {
		return nil, err
	}
	claims, err := jwt.Claims()
	if err != nil {
		return nil, err
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
	secret := os.Getenv(defaultJWTEnvVar)
	if secret == "" {
		secret = defaultJWTSecret
	}
	verifier, err := jose.NewVerifierHMAC(jose.JWK{
		Secret: []byte(secret),
	})
	if err != nil {
		return nil, err
	}
	if exp, _, err := claims.TimeClaim("exp"); err != nil {
		return nil, err
	} else if exp.Before(time.Now()) {
		return nil, ErrJWTExpired
	}
	err = verifier.Verify(jwt.Signature, []byte(jwt.Data()))
	return result, err
}

func (c *JWTCredentials) RequireTransportSecurity() bool {
	return false
}
