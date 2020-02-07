package api

import (
	"log"
	"net/http"

	"github.com/recoilme/pudge"
	guard "github.com/tomogoma/go-api-guard"
	typedError "github.com/tomogoma/go-typed-errors"
	"golang.org/x/net/context"
)

// Server represents the gRPC server
type Server struct {
}

// SayOk generates response ok to a Ping request
func (s *Server) SayOk(ctx context.Context, in *Empty) (*Ok, error) {
	//log.Printf("Receive message")
	return &Ok{Message: "ok"}, nil
}

//Key ...
type Key struct {
	Val []byte
}

//Value ...
func (k Key) Value() []byte {
	return k.Val
}

//KeyStore ...
type KeyStore struct {
	typedError.NotFoundErrCheck

	ExpInsAPIKErr        error
	ExpAPIKBUsrIDVal     Key
	ExpAPIKsBUsrIDValErr error
	RecInsAPIKUsrID      string
}

//APIKeyByUserIDVal ...
func (db *KeyStore) APIKeyByUserIDVal(userID string, key []byte) (guard.Key, error) {
	var err error
	var file = "db/tokens.db"

	err = pudge.Get(file, []byte(userID), &key)

	if err != nil {
		log.Println(err)
	}

	db.ExpAPIKBUsrIDVal = Key{Val: key}

	if db.ExpAPIKsBUsrIDValErr != nil {
		return nil, db.ExpAPIKsBUsrIDValErr
	}
	if db.ExpAPIKBUsrIDVal.Val == nil {
		return nil, typedError.NewNotFound("not found")
	}
	return db.ExpAPIKBUsrIDVal, db.ExpAPIKsBUsrIDValErr
}

//InsertAPIKey ...
func (db *KeyStore) InsertAPIKey(userID string, key []byte) (guard.Key, error) {
	var err error
	var file = "db/tokens.db"

	if db.ExpInsAPIKErr != nil {
		return nil, db.ExpInsAPIKErr
	}

	err = pudge.Set(file, []byte(userID), key)
	if err != nil {
		return nil, err
	}

	db.RecInsAPIKUsrID = userID
	return Key{Val: key}, db.ExpInsAPIKErr
}

var (
	userIDValidated string
	tokenGenerated  string
	err             error
)

//GenerateToken ...
func (s *Server) GenerateToken(ctx context.Context, CmdGenerateToken *CmdGenerateToken) (*ResGenerateToken, error) {
	db := &KeyStore{}

	g, _ := guard.NewGuard(
		db,
	)

	// Generate API key
	Key, err := g.NewAPIKey(CmdGenerateToken.UserID)
	if err != nil {
		tokenGenerated = ""
		log.Println(err)
	} else {
		tokenGenerated = string(Key.Value())
		log.Printf("API token generated: %v\n", tokenGenerated)
	}

	//return tokenGenerated, err
	return &ResGenerateToken{Token: tokenGenerated}, err
}

//ValidateToken ...
func (s *Server) ValidateToken(ctx context.Context, CmdValidateToken *CmdValidateToken) (*ResValidateToken, error) {

	db := &KeyStore{}

	g, _ := guard.NewGuard(
		db,
	)

	// Validate API Key
	userIDValidated, err := g.APIKeyValid(Key{Val: []byte(CmdValidateToken.Token)}.Value())

	if err != nil {
		log.Println(err)
	} else {
		log.Printf("API token for %v is valid\n", userIDValidated)
	}

	return &ResValidateToken{UserID: userIDValidated}, err
}

//Parser ...
func Parser(w http.ResponseWriter, r *http.Request) {
	//r.ParseForm()
}
