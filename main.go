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
var commandMapper map[string]func(name string, arg string)

func Generate(args map[string]commando.ArgValue, flags map[string]commando.FlagValue) {
	cmd := args["identifier"].Value
	argument := args["argument"].Value

	// One month expiracy
	expirationTime := time.Now().Add(time.Minute * 60 * 24 * 30)
	claims := &Claims{
		Command: cmd,
		Arg:     argument,
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

		claims, _ = data.Claims.(jwt.MapClaims)
		arg, argOk := claims["arg"].(string)
		commandName, commandOk := claims["command"].(string)

		if !(argOk || commandOk) {
			return
		}

		if command, ok := commandMapper[commandName]; ok {
			command(commandName, arg)
		}
	})
	http.ListenAndServe(fmt.Sprintf(":%s", os.Getenv("SERVER_PORT")), nil)
}

func WakeOnLan(command string, macAddress string) {
	cmd := exec.Command("zsh", "-c", fmt.Sprintf("wakeonlan %s", macAddress))
	fmt.Println(cmd)
	if err := cmd.Run(); err != nil {
		fmt.Println(err)
	}
}

func DeployPortfolio(command string, pathToDeployScript string) {
	cmd := exec.Command("zsh", "-c", pathToDeployScript)
	fmt.Println(cmd)
	if err := cmd.Start(); err != nil {
		fmt.Println(err)
	}
}

func main() {
	godotenv.Load()
	jwtSecret = []byte(os.Getenv("JWT_SECRET"))
	commandMapper = make(map[string]func(string, string))

	commandMapper["WAKEUP"] = WakeOnLan
	commandMapper["DEPLOY"] = DeployPortfolio

	commando.
		SetExecutableName("ifttt-wol").
		SetDescription("Creates an http server able to wake an ethernet wired device")

	commando.
		Register("generate").
		AddArgument("identifier", "Identifier of the command", "").
		AddArgument("argument", "argument your would like to attach to the token", "").
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
