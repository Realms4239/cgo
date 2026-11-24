package api

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// NewAuditID — horodatage + 4 octets aléatoires : deux audits dans la même
// seconde ne collident pas et l'identifiant n'est pas devinable.
func NewAuditID() string {
	var b [4]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic("crypto/rand indisponible: " + err.Error())
	}
	return fmt.Sprintf("audit-%d-%s", time.Now().Unix(), hex.EncodeToString(b[:]))
}

// devNoAuth — opt-in explicite (--dev-no-auth) : désactive l'authentification
// sur loopback uniquement. Le fail closed sans jeton est imposé au démarrage
// par cmd/cgo/main.go (exit 2), jamais par une variable d'environnement.
var devNoAuth bool

// SetDevNoAuth — pose l'opt-in depuis cgo --serve avant la construction du
// serveur. Jamais activable par variable d'environnement : seul le drapeau
// explicite du binaire peut désactiver l'authentification.
func SetDevNoAuth(v bool) { devNoAuth = v }

const sessionCookie = "cgo_session"

// bootKey — clé tirée au hasard à chaque démarrage. Un cookie émis par une
// instance précédente ne survit pas à un restart, malgré son Max-Age client.
var bootKey [32]byte

func init() {
	if _, err := rand.Read(bootKey[:]); err != nil {
		panic("crypto/rand indisponible pour la clé de session: " + err.Error())
	}
}

// newSessionToken — jti aléatoire (16 octets) signé par bootKey, au format
// « jti.signature ». Le MAC porte sur le jti hex, comme dans validSession.
func newSessionToken() string {
	var nonce [16]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		panic("crypto/rand indisponible pour le jti de session: " + err.Error())
	}
	jti := hex.EncodeToString(nonce[:])
	mac := hmac.New(sha256.New, bootKey[:])
	_, _ = mac.Write([]byte(jti))
	return jti + "." + hex.EncodeToString(mac.Sum(nil))
}

// revokedJTI — sessions révoquées dans ce processus, sous mutex car les
// handlers HTTP sont concurrents. Un jeton révoqué ne redevient jamais valide.
var (
	revokedMu  sync.Mutex
	revokedJTI = map[string]bool{}
)

// revokeSession — retire un jeton des sessions acceptées, définitivement.
func revokeSession(token string) {
	if token == "" {
		return
	}
	revokedMu.Lock()
	revokedJTI[token] = true
	revokedMu.Unlock()
}

// validSession — signature en temps constant puis table des révocations.
// Indépendant de CGO_API_TOKEN : l'inertie sans secret est imposée par
// authenticated(), qui refuse toute requête avant d'appeler ce validateur.
func validSession(token string) bool {
	jti, sig, ok := strings.Cut(token, ".")
	if !ok {
		return false
	}
	mac := hmac.New(sha256.New, bootKey[:])
	_, _ = mac.Write([]byte(jti))
	if subtle.ConstantTimeCompare([]byte(sig), []byte(hex.EncodeToString(mac.Sum(nil)))) != 1 {
		return false
	}
	revokedMu.Lock()
	defer revokedMu.Unlock()
	return !revokedJTI[token]
}

// bearerOK — même comparaison, en-tête Authorization: Bearer.
func bearerOK(r *http.Request, want string) bool {
	got, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
	return ok && subtle.ConstantTimeCompare([]byte(got), []byte(want)) == 1
}

// authenticated — Bearer ou cookie session. Sans jeton configuré, REFUS
// partout : seul le mode dev explicite (--dev-no-auth) laisse passer le
// loopback. Aucun contournement implicite, y compris depuis le loopback.
func authenticated(r *http.Request) bool {
	want := os.Getenv("CGO_API_TOKEN")
	if want == "" {
		return devNoAuth && isLoopback(r.RemoteAddr)
	}
	if bearerOK(r, want) {
		return true
	}
	if c, err := r.Cookie(sessionCookie); err == nil && validSession(c.Value) {
		return true
	}
	return false
}

// Principal — identité d'un appelant authentifié. OperatorID est l'empreinte
// du jeton : deux jetons distincts sont deux propriétaires distincts, sans
// table d'utilisateurs à maintenir.
type Principal struct {
	OperatorID string
	Role       string // "admin" ou "operator"
}

// principalFrom — un appelant authentifié avec le jeton de service est admin ;
// un appelant non authentifié n'a aucune identité (les gardes refusent alors).
func principalFrom(r *http.Request) Principal {
	if !authenticated(r) {
		return Principal{}
	}
	token := os.Getenv("CGO_API_TOKEN")
	sum := sha256.Sum256([]byte(token))
	return Principal{OperatorID: hex.EncodeToString(sum[:])[:16], Role: "admin"}
}

// authorizeResource — lecture/mutation d'une ressource possédée. Les rapports
// historiques sans owner_id restent lisibles ; leur mutation reste réservée à
// l'admin (contrôlée par authorizeMutation via requireOwnerMutate).
func authorizeResource(p Principal, ownerID string) bool {
	if p.Role == "admin" {
		return true
	}
	return ownerID == "" || ownerID == p.OperatorID
}

// authorizeMutation — une ressource sans propriétaire n'est modifiable que par
// l'admin ; un opérateur ne modifie que la sienne.
func authorizeMutation(p Principal, ownerID string) bool {
	if p.Role == "admin" {
		return true
	}
	return ownerID != "" && ownerID == p.OperatorID
}

// principal — identité lue par les gardes de propriété. Le seam est nil en
// production : principalFrom reste la seule source, donc aucun appelant réel
// ne change de rôle ni de droits.
func (s *Server) principal(r *http.Request) Principal {
	if s.principalFn != nil {
		return s.principalFn(r)
	}
	return principalFrom(r)
}

// requireOwner — garde de lecture : le propriétaire (ou l'admin) passe, un
// autre opérateur reçoit 403. Les requireAuth existants restent en place.
func (s *Server) requireOwner(next http.HandlerFunc, ownerOf func(*http.Request) string) http.HandlerFunc {
	return s.ownerGuard(next, ownerOf, authorizeResource)
}

// requireOwnerMutate — garde de mutation : en plus, une ressource sans
// propriétaire est refusée à tout non-admin.
func (s *Server) requireOwnerMutate(next http.HandlerFunc, ownerOf func(*http.Request) string) http.HandlerFunc {
	return s.ownerGuard(next, ownerOf, authorizeMutation)
}

func (s *Server) ownerGuard(next http.HandlerFunc, ownerOf func(*http.Request) string, allow func(Principal, string) bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !allow(s.principal(r), ownerOf(r)) {
			writeErr(w, r, "accès refusé à cette ressource", http.StatusForbidden)
			return
		}
		next(w, r)
	}
}

// requireAuth — mutations : Bearer ou cookie, sinon 401. Sans jeton, seul le
// mode dev explicite (--dev-no-auth) laisse passer le loopback.
func requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if authenticated(r) {
			next(w, r)
			return
		}
		w.Header().Set("WWW-Authenticate", "Bearer")
		writeErr(w, r, "authentification requise", http.StatusUnauthorized)
	}
}

func isLoopback(addr string) bool {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		host = addr
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}
