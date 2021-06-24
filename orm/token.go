package orm

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/dgrijalva/jwt-go"
	"strings"
	"time"
)

const (
	//未设置密钥有效期， 默认设置
	jwtKey    = "aabbccddeeffgg00112233445566"
	jwtExpire = "5m"
	jwtIssuer = "gin-issuer"
)

var Token JWT

type WJTConf struct {
	Secret string `env:"WJT_SECRET"`
	Expir  string `env:"WJT_EXPIRE"`
	Issuer string `env:"WJT_ISSUER"`
}

//加密主体, 根据实际情况修改
type Claims struct {
	ID int64
	jwt.StandardClaims
}

//处理结构
type JWT struct {
	Claims    Claims
	expire    string
	jwtSecret []byte
}

func (obj *JWT) Init(conf WJTConf) {
	obj.SetSecret([]byte(conf.Secret))
	obj.SetExpire(conf.Expir)
	obj.SetIssuer(conf.Issuer)
}

// wjt 有效期   m 分钟    h  小时   s   秒
// 300s
func (obj *JWT) SetExpire(Expire string) {
	obj.expire = Expire
}
func (obj *JWT) GetExpire() time.Time {
	return time.Unix(obj.Claims.StandardClaims.ExpiresAt, 0)
}

func (obj *JWT) SetSecret(secret []byte) {
	obj.jwtSecret = secret
}
func (obj *JWT) GetSecret() []byte {
	return obj.jwtSecret
}

func (obj *JWT) SetIssuer(issuer string) {
	obj.Claims.Issuer = issuer
}
func (obj *JWT) GetIssuer() string {
	return obj.Claims.Issuer
}

// 生成token
func (obj *JWT) GenerateToken() (string, error) {

	if len(obj.jwtSecret) == 0 {
		obj.SetSecret([]byte(jwtKey))
	}
	if obj.expire == "" {
		obj.SetExpire(jwtExpire)
	}
	if obj.Claims.Issuer == "" {
		obj.Claims.Issuer = jwtIssuer
	}

	m, _ := time.ParseDuration(obj.expire)
	obj.Claims.StandardClaims.ExpiresAt = time.Now().Add(m).Unix()
	obj.Claims.IssuedAt = time.Now().Unix()

	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, obj.Claims)
	return tokenClaims.SignedString(obj.jwtSecret)
}

// 验证token
func (obj JWT) ParseToken(token string) (*Claims, error) {
	if token == "" {
		return nil, errors.New("token为空")
	}
	tokenClaims, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return obj.jwtSecret, nil
	})

	if tokenClaims != nil {
		if claims, ok := tokenClaims.Claims.(*Claims); ok && tokenClaims.Valid {
			return claims, nil
		}
	}
	return nil, err
}

//获取token主体信息, 不验证加token是否有效
func (obj JWT) GetPayload(token string) (*Claims, error) {
	if token == "" {
		return nil, errors.New("token为空")
	}

	tmpArray := strings.Split(token, ".")
	if len(tmpArray) != 3 {
		return nil, errors.New("token参数传递错误")
	}

	j, err := base64.RawURLEncoding.DecodeString(tmpArray[1])
	if err != nil {
		return nil, err
	}

	var claims Claims
	if err := json.Unmarshal(j, &claims); err != nil {
		return nil, err
	}

	return &claims, nil
}
