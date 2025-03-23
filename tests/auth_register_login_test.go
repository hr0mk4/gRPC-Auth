package tests

import (
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/golang-jwt/jwt/v5"
	"github.com/hr0mk4/grpc_auth/tests/suite"
	authv1 "github.com/hr0mk4/protos_auth/gen/go/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	emptyAppId = 0
	appId      = 1
	appSecret  = "test-secret"
	passDefLen = 10
	deltaSec   = 1
)

func TestRegisterLogin_HappyPath(t *testing.T) {
	ctx, st := suite.New(t)

	email := gofakeit.Email()
	password := randomPassword()

	respReg, err := st.AuthClient.Register(ctx, &authv1.RegisterRequest{Email: email, Password: password})

	require.NoError(t, err)
	require.NotEmpty(t, respReg.GetUserId())

	respLog, err := st.AuthClient.LogIn(ctx, &authv1.LoginRequest{
		Email:    email,
		Password: password,
		AppId:    appId,
	})
	require.NoError(t, err)

	logTime := time.Now()

	token := respLog.GetToken()
	require.NotEmpty(t, token)

	parsedToken, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		return []byte(appSecret), nil
	})
	require.NoError(t, err)

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	require.True(t, ok)

	assert.Equal(t, respReg.GetUserId(), int64(claims["user_id"].(float64)))
	assert.Equal(t, appId, int(claims["app_id"].(float64)))
	assert.Equal(t, email, claims["email"].(string))

	assert.InDelta(t, logTime.Add(st.Cfg.TokenTTL).Unix(), int64(claims["exp"].(float64)), deltaSec)
}

func TestRegisterLogin_DoubleRegister(t *testing.T) {
	ctx, st := suite.New(t)

	email := gofakeit.Email()
	password := randomPassword()

	respReg, err := st.AuthClient.Register(ctx, &authv1.RegisterRequest{Email: email, Password: password})

	require.NoError(t, err)
	require.NotEmpty(t, respReg.GetUserId())

	respReg, err = st.AuthClient.Register(ctx, &authv1.RegisterRequest{Email: email, Password: password})

	require.Error(t, err)
	require.Nil(t, respReg)
}

func TestRegisterLogin_WrongPassword(t *testing.T) {
	ctx, st := suite.New(t)

	email := gofakeit.Email()
	password := randomPassword()

	respReg, err := st.AuthClient.Register(ctx, &authv1.RegisterRequest{Email: email, Password: password})

	require.NoError(t, err)
	require.NotEmpty(t, respReg.GetUserId())

	password = randomPassword()

	respLog, err := st.AuthClient.LogIn(ctx, &authv1.LoginRequest{
		Email:    email,
		Password: password,
		AppId:    appId,
	})
	require.Error(t, err)
	require.Nil(t, respLog)
}

func randomPassword() string {
	return gofakeit.Password(true, true, true, true, false, passDefLen)
}
