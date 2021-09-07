package storageos

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	stosapiv2 "github.com/storageos/go-api/v2"
)

const (
	// secretUsernameKey is the key in the secret that holds the username value.
	secretUsernameKey = "username"

	// secretPasswordKey is the key in the secret that holds the password value.
	secretPasswordKey = "password"
)

//go:generate mockgen -build_flags=--mod=vendor -destination=mocks/mock_control_plane.go -package=mocks . ControlPlane
//go:generate mockgen -build_flags=--mod=vendor -destination=mocks/mock_identifier.go -package=mocks . Identifier
//go:generate mockgen -build_flags=--mod=vendor -destination=mocks/mock_object.go -package=mocks . Object

// ControlPlane is the subset of the StorageOS control plane ControlPlane that
// api-manager requires.  New methods should be added here as needed, then the
// mocks regenerated.
type ControlPlane interface {
	RefreshJwt(ctx context.Context) (stosapiv2.UserSession, *http.Response, error)
	AuthenticateUser(ctx context.Context, authUserData stosapiv2.AuthUserData) (stosapiv2.UserSession, *http.Response, error)
	ListNamespaces(ctx context.Context) ([]stosapiv2.Namespace, *http.Response, error)
	DeleteNamespace(ctx context.Context, id string, version string, localVarOptionals *stosapiv2.DeleteNamespaceOpts) (*http.Response, error)
	ListNodes(ctx context.Context) ([]stosapiv2.Node, *http.Response, error)
	UpdateNode(ctx context.Context, id string, updateNodeData stosapiv2.UpdateNodeData) (stosapiv2.Node, *http.Response, error)
	DeleteNode(ctx context.Context, id string, version string, localVarOptionals *stosapiv2.DeleteNodeOpts) (*http.Response, error)
	SetComputeOnly(ctx context.Context, id string, setComputeOnlyNodeData stosapiv2.SetComputeOnlyNodeData, localVarOptionals *stosapiv2.SetComputeOnlyOpts) (stosapiv2.Node, *http.Response, error)
	ListVolumes(ctx context.Context, namespaceID string) ([]stosapiv2.Volume, *http.Response, error)
	GetVolume(ctx context.Context, namespaceID string, id string) (stosapiv2.Volume, *http.Response, error)
	UpdateVolume(ctx context.Context, namespaceID string, id string, updateVolumeData stosapiv2.UpdateVolumeData, localVarOptionals *stosapiv2.UpdateVolumeOpts) (stosapiv2.Volume, *http.Response, error)
	SetReplicas(ctx context.Context, namespaceID string, id string, setReplicasRequest stosapiv2.SetReplicasRequest, localVarOptionals *stosapiv2.SetReplicasOpts) (stosapiv2.AcceptedMessage, *http.Response, error)
	SetFailureMode(ctx context.Context, namespaceID string, id string, setFailureModeRequest stosapiv2.SetFailureModeRequest, localVarOptionals *stosapiv2.SetFailureModeOpts) (stosapiv2.Volume, *http.Response, error)
	UpdateNFSVolumeMountEndpoint(ctx context.Context, namespaceID string, id string, nfsVolumeMountEndpoint stosapiv2.NfsVolumeMountEndpoint, localVarOptionals *stosapiv2.UpdateNFSVolumeMountEndpointOpts) (*http.Response, error)
}

// Identifier is a StorageOS object that has an identity.
type Identifier interface {
	GetID() string
	GetName() string
	GetNamespace() string
}

// Object is a StorageOS object with metadata.
type Object interface {
	Identifier
	GetLabels() map[string]string
	IsHealthy() bool
}

// Client provides access to the StorageOS API.
type Client struct {
	API       ControlPlane
	transport http.RoundTripper
	ctx       context.Context
	traced    bool
}

const (
	// DefaultPort is the default api port.
	DefaultPort = 5705

	// DefaultScheme is used for api endpoint.
	DefaultScheme = "http"

	// TLSScheme scheme can be used if the api endpoint has TLS enabled.
	TLSScheme = "https"
)

var (
	// ErrNotInitialized is returned if the API client was accessed before it
	// was initialised.
	ErrNotInitialized = errors.New("api client not initialized")
	// ErrNoAuthToken is returned when the API client did not get an error
	// during authentication but no valid auth token was returned.
	ErrNoAuthToken = errors.New("no token found in auth response")

	// HTTPTimeout is the time limit for requests made by the API Client. The
	// timeout includes connection time, any redirects, and reading the response
	// body. The timer remains running after Get, Head, Post, or Do return and
	// will interrupt reading of the Response.Body.
	HTTPTimeout = 10 * time.Second

	// AuthenticationTimeout is the time limit for authentication requests to
	// complete.  It should be longer than the HTTPTimeout.
	AuthenticationTimeout = 20 * time.Second

	// DefaultRequestTimeout is the default time limit for api requests to
	// complete.  It should be longer than the HTTPTimeout.
	DefaultRequestTimeout = 20 * time.Second
)

// NewTestAPIClient returns a client that uses the provided ControlPlane api
// client. Intended for tests that use a mocked StorageOS stosapiv2.  This avoids
// having to publically expose the api on the Client struct.
func NewTestAPIClient(api ControlPlane) *Client {
	return &Client{
		API:       api,
		transport: http.DefaultTransport,
		ctx:       context.TODO(),
		traced:    false,
	}
}

// New returns a pre-authenticated client for the StorageOS API.  The
// authentication token must be refreshed periodically using
// AuthenticateRefresh().
func New(username, password, endpoint string) (context.Context, *stosapiv2.APIClient, error) {
	transport := http.DefaultTransport
	return newAuthenticatedClient(username, password, endpoint, transport)
	// ctx, client, err := newAuthenticatedClient(username, password, endpoint, transport)
	// if err != nil {
	// 	return nil, err
	// }
	// return &Client{API: client.DefaultApi, transport: transport, ctx: ctx}, nil
}

func newAuthenticatedClient(username, password, endpoint string, transport http.RoundTripper) (context.Context, *stosapiv2.APIClient, error) {
	config := stosapiv2.NewConfiguration()

	if !strings.Contains(endpoint, "://") {
		endpoint = fmt.Sprintf("%s://%s", DefaultScheme, endpoint)
	}

	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, nil, err
	}

	config.Scheme = u.Scheme
	config.Host = u.Host
	if !strings.Contains(u.Host, ":") {
		config.Host = fmt.Sprintf("%s:%d", u.Host, DefaultPort)
	}

	httpc := &http.Client{
		Timeout:   HTTPTimeout,
		Transport: transport,
	}
	config.HTTPClient = httpc

	// Get a wrappered API client.
	client := stosapiv2.NewAPIClient(config)

	// Authenticate and return context with credentials and client.
	ctx, err := Authenticate(client, username, password)
	if err != nil {
		return nil, nil, err
	}

	return ctx, client, nil
}

// Authenticate against the API and set the authentication token in the client
// to be used for subsequent API requests.  The token must be refreshed
// periodically using AuthenticateRefresh().
func Authenticate(client *stosapiv2.APIClient, username, password string) (context.Context, error) {
	// Create context just for the login.
	ctx, cancel := context.WithTimeout(context.Background(), AuthenticationTimeout)
	defer cancel()

	// Initial basic auth to retrieve the jwt token.
	_, resp, err := client.DefaultApi.AuthenticateUser(ctx, stosapiv2.AuthUserData{
		Username: username,
		Password: password,
	})
	if err != nil {
		return nil, stosapiv2.MapAPIError(err, resp)
	}
	defer resp.Body.Close()

	// Set auth token in a new context for re-use.
	token := respAuthToken(resp)
	if token == "" {
		return nil, ErrNoAuthToken
	}
	return context.WithValue(context.Background(), stosapiv2.ContextAccessToken, token), nil
}

// AddToken adds the current authentication token to a given context.
func (c *Client) AddToken(ctx context.Context) context.Context {
	return context.WithValue(ctx, stosapiv2.ContextAccessToken, c.ctx.Value(stosapiv2.ContextAccessToken))
}

// respAuthToken is a helper to pull the auth token out of a HTTP Response.
func respAuthToken(resp *http.Response) string {
	if value := resp.Header.Get("Authorization"); value != "" {
		// "Bearer aaaabbbbcccdddeeeff"
		return strings.Split(value, " ")[1]
	}
	return ""
}

// ReadCredsFromMountedSecret reads the api username and password from a
// Kubernetes secret mounted at the given path.  If the username or password in
// the secret changes, the data in the mounted file will also change.
func ReadCredsFromMountedSecret(path string) (string, string, error) {
	username, err := readFromSecret(filepath.Join(path, secretUsernameKey))
	if err != nil {
		return "", "", err
	}
	password, err := readFromSecret(filepath.Join(path, secretPasswordKey))
	if err != nil {
		return "", "", err
	}
	return username, password, nil
}

// readFromSecret reads data from a secret from the given path.  The secret is
// expected to be mounted into the container by Kubernetes.
func readFromSecret(path string) (string, error) {
	secretBytes, readErr := ioutil.ReadFile(path)
	if readErr != nil {
		return "", fmt.Errorf("unable to read secret: %s, error: %s", path, readErr)
	}
	val := strings.TrimSpace(string(secretBytes))
	return val, nil
}
