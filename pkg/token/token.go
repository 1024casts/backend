package token

import (
	"errors"
	"fmt"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

var (
	// ErrMissingHeader means the `Authorization` header was empty.
	ErrMissingHeader = errors.New("The length of the `Authorization` header is zero.")
)

// Context is the context of the JSON web token.
type Context struct {
	ID       uint64
	Username string
}

// secretFunc validates the secret format.
func secretFunc(secret string) jwt.Keyfunc {
	return func(token *jwt.Token) (interface{}, error) {
		// Make sure the `alg` is what we except.
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}

		return []byte(secret), nil
	}
}

// Parse validates the token with the specified secret,
// and returns the context if the token was valid.
func Parse(tokenString string, secret string) (*Context, error) {
	ctx := &Context{}

	// Parse the token.
	token, err := jwt.Parse(tokenString, secretFunc(secret))

	// Parse error.
	if err != nil {
		return ctx, err

		// Read the token if it's valid.
	} else if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		ctx.ID = uint64(claims["id"].(float64))
		ctx.Username = claims["username"].(string)
		return ctx, nil

		// Other errors.
	} else {
		return ctx, err
	}
}

// ParseRequest gets the token from the header and
// pass it to the Parse function to parses the token.
func ParseRequest(c *gin.Context) (*Context, error) {
	header := c.Request.Header.Get("Authorization")

	// Load the jwt secret from config
	secret := viper.GetString("jwt_secret")

	if len(header) == 0 {
		return &Context{}, ErrMissingHeader
	}

	var t string
	// Parse the header to get the token part.
	fmt.Sscanf(header, "Bearer %s", &t)
	return Parse(t, secret)
}

// Sign signs the context with the specified secret.
func Sign(c *gin.Context, ctx Context, secret string) (tokenString string, err error) {
	// Load the jwt secret from the Gin config if the secret isn't specified.
	if secret == "" {
		secret = viper.GetString("jwt_secret")
	}
	// The token content.
	//iss: （Issuer）签发者
	//iat: （Issued At）签发时间，用Unix时间戳表示
	//exp: （Expiration Time）过期时间，用Unix时间戳表示
	//aud: （Audience）接收该JWT的一方
	//sub: （Subject）该JWT的主题
	//nbf: （Not Before）不要早于这个时间
	//jti: （JWT ID）用于标识JWT的唯一ID
	var hour = viper.GetDuration("jwt_duration")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":       ctx.ID,
		"username": ctx.Username,
		"nbf":      time.Now().Unix(),
		"iat":      time.Now().Unix(),
		"exp":      time.Now().Add(time.Hour * hour).Unix(),
	})
	// Sign the token with the specified secret.
	tokenString, err = token.SignedString([]byte(secret))

	return
}
