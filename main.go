package main

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/sirupsen/logrus"
	ant "github.com/vibrantbyte/go-antpath/antpath"
	"net/http"
	"time"
)

var secret = []byte("test")
var matcher = ant.New()

//白名单
var blackList = []string{"/auth/token**", "/auth/auth**"}

const ForwardHeaderName = "X-Forwarded-Uri"

func main() {
	logrus.SetLevel(logrus.DebugLevel)
	http.HandleFunc("/token", func(writer http.ResponseWriter, request *http.Request) {
		logrus.Infof("访问路径:%s", request.RequestURI)
		request.ParseForm()
		uid := request.Form.Get("uid")
		claims := &jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
			Audience:  uid,
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signedString, err := token.SignedString(secret)
		if err != nil {
			fmt.Fprintln(writer, err.Error())
			return
		}
		fmt.Fprintln(writer, signedString)
		logrus.Infof("访问路径:%s 完成", "/token")
	})
	http.HandleFunc("/auth", func(writer http.ResponseWriter, request *http.Request) {
		logrus.Infof("访问路径:%s", request.RequestURI)
		//X-Forwarded-Uri
		uri := request.Header.Get(ForwardHeaderName)
		//如果在白名单内,直接放行
		for _, s := range blackList {
			match := matcher.Match(s, uri)
			if match {
				writer.WriteHeader(http.StatusOK)
				fmt.Fprintf(writer, "black list hit:%v", s)
				return
			}
		}
		value, ok := request.Header["Authorization"]
		if !ok {
			writer.WriteHeader(http.StatusForbidden)
			writer.Write([]byte("header上未找到Authorization"))
			return
		}
		parse, err := jwt.ParseWithClaims(value[0], &jwt.StandardClaims{}, func(token *jwt.Token) (i interface{}, err error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}
			return secret, nil
		})
		if err != nil {
			writer.WriteHeader(http.StatusForbidden)
			fmt.Fprintln(writer, err.Error())
			return
		}
		claims, ok := parse.Claims.(*jwt.StandardClaims)
		if !ok {
			writer.WriteHeader(http.StatusForbidden)
			fmt.Fprintln(writer, "token错误")
			return
		}
		if err := claims.Valid(); err != nil {
			writer.WriteHeader(http.StatusForbidden)
			fmt.Fprintln(writer, err.Error())
			return
		}
		//TODO:验证权限
		writer.Header().Add("Uid", claims.Audience)
		writer.Header().Set("abcd", "456")
		writer.WriteHeader(http.StatusOK)

		logrus.Infof("访问路径:%s 完成", request.RequestURI)
	})
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		logrus.Infof("访问路径:%s", request.RequestURI)
		for s, strings := range request.Header {
			logrus.Infof("key:%v,value:%v", s, strings)
			writer.Write([]byte(fmt.Sprintf("key:%v,value:%v \n", s, strings)))
		}
		logrus.Infof("访问路径:%s 完成", request.RequestURI)
	})
	http.ListenAndServe(":9999", nil)
}
