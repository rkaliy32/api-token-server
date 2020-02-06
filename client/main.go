package main

import (
	"context"
	"fmt"
	"log"
	"os"

	dbapi "github.com/rkaliy32/api-token-server/api"
	"google.golang.org/grpc"
)

var (
	userIDValidated string
	tokenGenerated  string
	err             error
)

//generateAPIToken ...
func generateAPIToken(userID string) (string, error) {
	var conn *grpc.ClientConn
	var err error

	conn, err = grpc.Dial(":7777", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %s", err)
	}
	defer conn.Close()

	c := dbapi.NewOkdbClient(conn)

	// SayOk
	response, err := c.SayOk(context.Background(), &dbapi.Empty{})
	if err != nil {
		log.Fatalf("Error when calling db server: %s", err)
	}
	log.Printf("Response from server: %s", response.Message)

	// Generate Token
	var result *dbapi.ResGenerateToken
	result, err = c.GenerateToken(context.Background(), &dbapi.CmdGenerateToken{UserID: userID})
	if err != nil {
		log.Println(err)
		result = &dbapi.ResGenerateToken{}
	}

	return result.Token, err
}

//validateAPIToken ...
func validateAPIToken(apiToken string) (string, error) {
	var conn *grpc.ClientConn
	var err error

	conn, err = grpc.Dial(":7777", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %s", err)
	}
	defer conn.Close()

	c := dbapi.NewOkdbClient(conn)

	// SayOk
	response, err := c.SayOk(context.Background(), &dbapi.Empty{})
	if err != nil {
		log.Fatalf("Error when calling db server: %s", err)
	}
	log.Printf("Response from server: %s", response.Message)

	// Validate Token
	var result *dbapi.ResValidateToken
	result, err = c.ValidateToken(context.Background(), &dbapi.CmdValidateToken{Token: apiToken})
	if err != nil {
		log.Println(err)
		result = &dbapi.ResValidateToken{}
	}

	return result.UserID, err

}

//main ...
func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage:\n", os.Args[0], "generate {UserID}")
		fmt.Println(os.Args[0], "validate {APIToken}")
		return
	}
	if os.Args[1] == "generate" {
		tokenGenerated, err = generateAPIToken(os.Args[2])
		if err == nil {
			fmt.Printf("token generated %v\n", tokenGenerated)
		} else {
			fmt.Println(err)
		}
	} else if os.Args[1] == "validate" {
		userIDValidated, err = validateAPIToken(os.Args[2])
		if err == nil {
			fmt.Printf("API key is validated for user %v\n", userIDValidated)
		} else {
			fmt.Printf("Invalid API key")
		}
	} else {
		fmt.Printf("Command error\n")
		fmt.Println("Usage:\n", os.Args[0], "generate {UserID}")
		fmt.Println(os.Args[0], "validate {APIToken}")
		return
	}
}
