package auth

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type Auth struct {
	Issuer        string
	Audience      string
	Secret        string
	TokenExpiry   time.Duration
	RefreshExpiry time.Duration
	CookieDomain  string
	CookiePath    string
	CookieName    string
}

type JwtUser struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

type TokenPairs struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type Claims struct {
	jwt.RegisteredClaims
	ID    int    `json:"id"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

func (j *Auth) HashPassword(password string) (string, error) {
	hashedPw, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPw), nil
}

func (j *Auth) GenerateTokenPairAndCookie(user *JwtUser) (TokenPairs, *http.Cookie, error) {
	claims := &Claims{
		jwt.RegisteredClaims{
			Subject:   fmt.Sprint(user.ID),
			Audience:  []string{j.Audience},
			Issuer:    j.Issuer,
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(j.TokenExpiry)),
		},
		user.ID,
		user.Email,
		user.Role,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedAccessToken, err := token.SignedString([]byte(j.Secret))
	if err != nil {
		return TokenPairs{}, nil, err
	}

	refreshTokenClaims := &Claims{
		jwt.RegisteredClaims{
			Subject:   fmt.Sprint(user.ID),
			Issuer:    j.Issuer,
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(j.RefreshExpiry)),
		},
		user.ID,
		user.Email,
		user.Role,
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshTokenClaims)

	signedRefreshToken, err := refreshToken.SignedString([]byte(j.Secret))
	if err != nil {
		return TokenPairs{}, nil, err
	}

	tokenPairs := TokenPairs{
		AccessToken:  signedAccessToken,
		RefreshToken: signedRefreshToken,
	}

	isDebug := os.Getenv("ENV")
	var cookieDomain string
	if isDebug == "debug" {
		cookieDomain = os.Getenv("COOKIE_DOMAIN_DEBUG")
	} else {
		cookieDomain = os.Getenv("COOKIE_DOMAIN")
	}

	accessTokenCookie := &http.Cookie{
		Name:     "AccessCookie",
		Path:     "/",
		Value:    signedAccessToken,
		Expires:  time.Now().Add(j.TokenExpiry),
		MaxAge:   int(j.TokenExpiry.Seconds()),
		SameSite: http.SameSiteNoneMode,
		Domain:   cookieDomain,
		HttpOnly: true,
		Secure:   true,
	}

	return tokenPairs, accessTokenCookie, nil
}

func (j *Auth) GetRefreshCookie(refreshToken string) *http.Cookie {
	isDebug := os.Getenv("ENV")
	var cookieDomain string
	if isDebug == "debug" {
		cookieDomain = os.Getenv("COOKIE_DOMAIN_DEBUG")
	} else {
		cookieDomain = os.Getenv("COOKIE_DOMAIN")
	}

	cookie := &http.Cookie{
		Name:     "RefreshCookie",
		Path:     "/",
		Value:    refreshToken,
		Expires:  time.Now().Add(j.RefreshExpiry),
		MaxAge:   int(j.RefreshExpiry.Seconds()),
		SameSite: http.SameSiteNoneMode,
		Domain:   cookieDomain,
		HttpOnly: true,
		Secure:   true,
	}

	// fmt.Println("cookie: ", cookie)
	return cookie
}

func (j *Auth) GetExpiredRefreshCookie() *http.Cookie {
	isDebug := os.Getenv("ENV")
	var cookieDomain string
	if isDebug == "debug" {
		cookieDomain = os.Getenv("COOKIE_DOMAIN_DEBUG")
	} else {
		cookieDomain = os.Getenv("COOKIE_DOMAIN")
	}
	return &http.Cookie{
		Name:     "RefreshCookie",
		Path:     "/",
		Value:    "",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		SameSite: http.SameSiteNoneMode,
		Domain:   cookieDomain,
		HttpOnly: true,
		Secure:   true,
	}
}

func (j *Auth) GetTokenAndVerify(w http.ResponseWriter, r *http.Request) (string, *Claims, error) {
	cookie, err := r.Cookie("RefreshCookie")
	if err != nil {
		return "", nil, fmt.Errorf("no cookie: %v", err)
	}

	token := cookie.Value

	claims := &Claims{}

	_, err = jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("wrong signing method: %v", token.Header["alg"])
		}
		return []byte(j.Secret), nil
	})
	if err != nil {
		if strings.HasPrefix(err.Error(), "token is expired by") {
			return "", nil, errors.New("token expired")
		}
		return "", nil, err
	}

	if claims.Issuer != j.Issuer {
		return "", nil, errors.New("invalid issuer")
	}
	return token, claims, nil
}

func (j *Auth) GetUserRoleFromJWT(tokenStr string) (string, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(j.Secret), nil
	})

	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		role := claims["role"].(string)
		return role, nil
	}

	return "", fmt.Errorf("invalid token or role not found")
}
