package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/joho/godotenv"
	"github.com/thatisuday/commando"
)

type Claims struct {
	Command string `json:"command"`
	Arg     string `json:"arg"`
	jwt.StandardClaims
}

type TokenPayload struct {
	Token string `json:"token"`
}

var jwtSecret []byte

func Generate(args map[string]commando.ArgValue, flags map[string]commando.FlagValue) {
	macAddress := args["mac"].Value

	// One month expiracy
	expirationTime := time.Now().Add(time.Minute * 60 * 24 * 30)
	claims := &Claims{
		Command: "WAKEUP",
		Arg:     macAddress,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)

	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(tokenString)
}

func Server(args map[string]commando.ArgValue, flags map[string]commando.FlagValue) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var token TokenPayload

		err := json.NewDecoder(r.Body).Decode(&token)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		claims := jwt.MapClaims{}
		data, err := jwt.ParseWithClaims(token.Token, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if claims, ok := data.Claims.(jwt.MapClaims); ok && data.Valid {
			if str, ok := claims["command"].(string); ok && str == "WAKEUP" {
				if str, ok := claims["arg"].(string); ok {
					cmd := exec.Command("zsh", "-c", fmt.Sprintf("wakeonlan %s", str))
					if err := cmd.Run(); err != nil {
						fmt.Println(err)
					}
				}
			}
		} else {
			fmt.Println(err)
		}
	})
	http.ListenAndServe(fmt.Sprintf(":%s", os.Getenv("SERVER_PORT")), nil)
}

func main() {
	godotenv.Load()
	jwtSecret = []byte(os.Getenv("JWT_SECRET"))

	commando.
		SetExecutableName("ifttt-wol").
		SetDescription("Creates an http server able to wake an ethernet wired device")

	commando.
		Register("generate").
		AddArgument("mac", "MAC address of the device it should wake up", "").
		SetDescription("Generates a JWT that will be accepted to use the WakeOnLAN utility").
		SetShortDescription("Generates a JWT that will be accepted to use the WakeOnLAN utility").
		SetAction(Generate)

	commando.
		Register("server").
		SetDescription("Runs the server that intercepts the Webhook").
		SetShortDescription("Runs the server that intercepts the Webhook").
		SetAction(Server)

	commando.Parse(nil)
}
