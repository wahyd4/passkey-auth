package auth

import (
	"net/http"

	"passkey-auth/internal/config"
	"passkey-auth/internal/database"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
)

type WebAuthn struct {
	web *webauthn.WebAuthn
	db  *database.DB
}

// WebAuthnUser implements the webauthn.User interface
type WebAuthnUser struct {
	user        *database.User
	credentials []*database.Credential
}

func (u *WebAuthnUser) SetUser(user *database.User) {
	u.user = user
}

func (u *WebAuthnUser) GetUser() *database.User {
	return u.user
}

func (u *WebAuthnUser) WebAuthnID() []byte {
	return []byte(u.user.Email)
}

func (u *WebAuthnUser) WebAuthnName() string {
	return u.user.Email
}

func (u *WebAuthnUser) WebAuthnDisplayName() string {
	return u.user.DisplayName
}

func (u *WebAuthnUser) WebAuthnIcon() string {
	return ""
}

func (u *WebAuthnUser) WebAuthnCredentials() []webauthn.Credential {
	var creds []webauthn.Credential
	for _, dbCred := range u.credentials {
		creds = append(creds, webauthn.Credential{
			ID:              dbCred.ID,
			PublicKey:       dbCred.PublicKey,
			AttestationType: dbCred.AttestationType,
			Authenticator: webauthn.Authenticator{
				AAGUID:       dbCred.AAGUID,
				SignCount:    dbCred.SignCount,
				CloneWarning: dbCred.CloneWarning,
			},
		})
	}
	return creds
}

func NewWebAuthn(cfg *config.Config) (*WebAuthn, error) {
	wconfig := &webauthn.Config{
		RPDisplayName: cfg.WebAuthn.RPDisplayName,
		RPID:          cfg.WebAuthn.RPID,
		RPOrigins:     cfg.WebAuthn.RPOrigins,
	}

	web, err := webauthn.New(wconfig)
	if err != nil {
		return nil, err
	}

	return &WebAuthn{
		web: web,
	}, nil
}

func (wa *WebAuthn) SetDB(db *database.DB) {
	wa.db = db
}

func (wa *WebAuthn) GetUserByEmail(email string) (*WebAuthnUser, error) {
	user, err := wa.db.GetUserByEmail(email)
	if err != nil {
		return nil, err
	}

	credentials, err := wa.db.GetCredentialsByUserID(user.ID)
	if err != nil {
		return nil, err
	}

	return &WebAuthnUser{
		user:        user,
		credentials: credentials,
	}, nil
}

func (wa *WebAuthn) BeginRegistration(user *WebAuthnUser) (*protocol.CredentialCreation, *webauthn.SessionData, error) {
	creation, sessionData, err := wa.web.BeginRegistration(user)
	return creation, sessionData, err
}

func (wa *WebAuthn) FinishRegistration(user *WebAuthnUser, sessionData webauthn.SessionData, response *http.Request) (*webauthn.Credential, error) {
	credential, err := wa.web.FinishRegistration(user, sessionData, response)
	return credential, err
}

func (wa *WebAuthn) BeginLogin(user *WebAuthnUser) (*protocol.CredentialAssertion, *webauthn.SessionData, error) {
	assertion, sessionData, err := wa.web.BeginLogin(user)
	return assertion, sessionData, err
}

func (wa *WebAuthn) FinishLogin(user *WebAuthnUser, sessionData webauthn.SessionData, response *http.Request) (*webauthn.Credential, error) {
	credential, err := wa.web.FinishLogin(user, sessionData, response)
	return credential, err
}

func (wa *WebAuthn) SaveCredential(userID int, cred *webauthn.Credential) error {
	dbCred := &database.Credential{
		ID:              cred.ID,
		UserID:          userID,
		PublicKey:       cred.PublicKey,
		AttestationType: cred.AttestationType,
		AAGUID:          cred.Authenticator.AAGUID,
		SignCount:       cred.Authenticator.SignCount,
		CloneWarning:    cred.Authenticator.CloneWarning,
	}

	return wa.db.SaveCredential(dbCred)
}

func (wa *WebAuthn) UpdateCredentialSignCount(credID []byte, signCount uint32) error {
	return wa.db.UpdateCredentialSignCount(credID, signCount)
}
