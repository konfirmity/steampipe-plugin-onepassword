package onepassword

import (
	"context"
	"errors"
	"os"
	"sync"
	"time"

	"github.com/1Password/connect-sdk-go/connect"
	"github.com/1Password/connect-sdk-go/onepassword"
	opsdk "github.com/1password/onepassword-sdk-go"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
)

// vaultBackend is the vault and item source behind the tables, either a Connect server or the
// service-account SDK. Rows keep the Connect types so the table definitions work with both.
type vaultBackend interface {
	ListVaults(ctx context.Context) ([]onepassword.Vault, error)
	GetVault(ctx context.Context, id string) (*onepassword.Vault, error)
	ListItems(ctx context.Context, vaultID string) ([]itemOverview, error)
}

// itemOverview is the metadata of an item. The favorite, version, trashed and last_edited_by
// fields come only from Connect, so they are pointers that stay nil for service-account rows.
type itemOverview struct {
	ID           string
	VaultID      string
	Title        string
	Category     string
	State        string
	Tags         []string
	URLs         []onepassword.ItemURL
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Favorite     *bool
	Version      *int
	Trashed      *bool
	LastEditedBy *string
}

const defaultConnectURL = "http://localhost:8080"

var errMissingCredentials = errors.New("either 'service_account_token' or 'token' must be set in the connection configuration. Edit your connection configuration file and then restart Steampipe.")

type authMode int

const (
	authConnect authMode = iota
	authServiceAccount
)

type authSettings struct {
	mode  authMode
	token string
	url   string
}

// resolveAuth picks the credential: connection config beats environment variables, and within
// each source a service account token beats a Connect token.
func resolveAuth(cfg onepasswordConfig, getenv func(string) string) (authSettings, error) {
	switch {
	case nonEmpty(cfg.ServiceAccountToken):
		return authSettings{mode: authServiceAccount, token: *cfg.ServiceAccountToken}, nil
	case nonEmpty(cfg.Token):
		return connectSettings(*cfg.Token, cfg, getenv), nil
	case getenv("OP_SERVICE_ACCOUNT_TOKEN") != "":
		return authSettings{mode: authServiceAccount, token: getenv("OP_SERVICE_ACCOUNT_TOKEN")}, nil
	case getenv("OP_CONNECT_TOKEN") != "":
		return connectSettings(getenv("OP_CONNECT_TOKEN"), cfg, getenv), nil
	}
	return authSettings{}, errMissingCredentials
}

func connectSettings(token string, cfg onepasswordConfig, getenv func(string) string) authSettings {
	url := defaultConnectURL
	if nonEmpty(cfg.URL) {
		url = *cfg.URL
	} else if env := getenv("OP_CONNECT_HOST"); env != "" {
		url = env
	}
	return authSettings{mode: authConnect, token: token, url: url}
}

func nonEmpty(s *string) bool {
	return s != nil && *s != ""
}

var sdkClientMu sync.Mutex

const itemStateActive = "active"

const sdkBackendCacheKey = "onepassword_sdk_backend"

func getBackend(ctx context.Context, d *plugin.QueryData) (vaultBackend, error) {
	auth, err := resolveAuth(GetConfig(d.Connection), os.Getenv)
	if err != nil {
		return nil, err
	}
	if auth.mode == authConnect {
		return connectBackend{client: connect.NewClient(auth.url, auth.token)}, nil
	}

	// SDK client start-up loads a WASM core and authenticates, so one client serves the connection.
	sdkClientMu.Lock()
	defer sdkClientMu.Unlock()
	if cached, ok := d.ConnectionCache.Get(ctx, sdkBackendCacheKey); ok {
		return cached.(sdkBackend), nil
	}
	client, err := opsdk.NewClient(ctx,
		opsdk.WithServiceAccountToken(auth.token),
		opsdk.WithIntegrationInfo("Steampipe 1Password plugin", "v1"),
	)
	if err != nil {
		return nil, err
	}
	backend := sdkBackend{vaults: client.Vaults(), items: client.Items()}
	if err := d.ConnectionCache.Set(ctx, sdkBackendCacheKey, backend); err != nil {
		plugin.Logger(ctx).Warn("onepassword.getBackend", "cache_error", err)
	}
	return backend, nil
}

type connectBackend struct {
	client connect.Client
}

func (b connectBackend) ListVaults(_ context.Context) ([]onepassword.Vault, error) {
	return b.client.GetVaults()
}

func (b connectBackend) GetVault(_ context.Context, id string) (*onepassword.Vault, error) {
	return b.client.GetVault(id)
}

func (b connectBackend) ListItems(_ context.Context, vaultID string) ([]itemOverview, error) {
	items, err := b.client.GetItems(vaultID)
	if err != nil {
		return nil, err
	}
	overviews := make([]itemOverview, 0, len(items))
	for _, i := range items {
		overviews = append(overviews, itemFromConnect(i))
	}
	return overviews, nil
}

func itemFromConnect(i onepassword.Item) itemOverview {
	return itemOverview{
		ID:           i.ID,
		VaultID:      i.Vault.ID,
		Title:        i.Title,
		Category:     string(i.Category),
		State:        itemStateActive,
		Tags:         i.Tags,
		URLs:         i.URLs,
		CreatedAt:    i.CreatedAt,
		UpdatedAt:    i.UpdatedAt,
		Favorite:     &i.Favorite,
		Version:      &i.Version,
		Trashed:      &i.Trashed,
		LastEditedBy: &i.LastEditedBy,
	}
}

type sdkVaultsAPI interface {
	List(ctx context.Context, params ...opsdk.VaultListParams) ([]opsdk.VaultOverview, error)
	GetOverview(ctx context.Context, vaultID string) (opsdk.VaultOverview, error)
}

type sdkItemsAPI interface {
	List(ctx context.Context, vaultID string, filters ...opsdk.ItemListFilter) ([]opsdk.ItemOverview, error)
}

type sdkBackend struct {
	vaults sdkVaultsAPI
	items  sdkItemsAPI
}

func (b sdkBackend) ListVaults(ctx context.Context) ([]onepassword.Vault, error) {
	overviews, err := b.vaults.List(ctx)
	if err != nil {
		return nil, err
	}
	vaults := make([]onepassword.Vault, 0, len(overviews))
	for _, v := range overviews {
		vaults = append(vaults, vaultFromSDK(v))
	}
	return vaults, nil
}

func (b sdkBackend) GetVault(ctx context.Context, id string) (*onepassword.Vault, error) {
	overview, err := b.vaults.GetOverview(ctx, id)
	if err != nil {
		return nil, err
	}
	vault := vaultFromSDK(overview)
	return &vault, nil
}

func (b sdkBackend) ListItems(ctx context.Context, vaultID string) ([]itemOverview, error) {
	overviews, err := b.items.List(ctx, vaultID)
	if err != nil {
		return nil, err
	}
	items := make([]itemOverview, 0, len(overviews))
	for _, o := range overviews {
		items = append(items, itemFromSDKOverview(o))
	}
	return items, nil
}

var sdkVaultTypes = map[opsdk.VaultType]onepassword.VaultType{
	opsdk.VaultTypePersonal:    onepassword.PersonalVault,
	opsdk.VaultTypeEveryone:    onepassword.EveryoneVault,
	opsdk.VaultTypeTransfer:    onepassword.TransferVault,
	opsdk.VaultTypeUserCreated: onepassword.UserCreatedVault,
}

func vaultFromSDK(v opsdk.VaultOverview) onepassword.Vault {
	vaultType, ok := sdkVaultTypes[v.VaultType]
	if !ok {
		vaultType = onepassword.UnknownVault
	}
	return onepassword.Vault{
		ID:             v.ID,
		Name:           v.Title,
		Description:    v.Description,
		AttrVersion:    int(v.AttributeVersion),
		ContentVersoin: int(v.ContentVersion),
		Items:          int(v.ActiveItemCount),
		Type:           vaultType,
		CreatedAt:      v.CreatedAt,
		UpdatedAt:      v.UpdatedAt,
	}
}

// The SDK and Connect name categories differently; rows use the Connect names so category
// filters behave the same whichever credential is configured.
var sdkItemCategories = map[opsdk.ItemCategory]onepassword.ItemCategory{
	opsdk.ItemCategoryLogin:                onepassword.Login,
	opsdk.ItemCategoryPassword:             onepassword.Password,
	opsdk.ItemCategoryAPICredentials:       onepassword.ApiCredential,
	opsdk.ItemCategoryServer:               onepassword.Server,
	opsdk.ItemCategoryDatabase:             onepassword.Database,
	opsdk.ItemCategoryCreditCard:           onepassword.CreditCard,
	opsdk.ItemCategoryMembership:           onepassword.Membership,
	opsdk.ItemCategoryPassport:             onepassword.Passport,
	opsdk.ItemCategorySoftwareLicense:      onepassword.SoftwareLicense,
	opsdk.ItemCategoryOutdoorLicense:       onepassword.OutdoorLicense,
	opsdk.ItemCategorySecureNote:           onepassword.SecureNote,
	opsdk.ItemCategoryRouter:               onepassword.WirelessRouter,
	opsdk.ItemCategoryBankAccount:          onepassword.BankAccount,
	opsdk.ItemCategoryDriverLicense:        onepassword.DriverLicense,
	opsdk.ItemCategoryIdentity:             onepassword.Identity,
	opsdk.ItemCategoryRewards:              onepassword.RewardProgram,
	opsdk.ItemCategoryDocument:             onepassword.Document,
	opsdk.ItemCategoryEmail:                onepassword.EmailAccount,
	opsdk.ItemCategorySocialSecurityNumber: onepassword.SocialSecurityNumber,
	opsdk.ItemCategoryMedicalRecord:        onepassword.MedicalRecord,
	opsdk.ItemCategorySSHKey:               onepassword.SSHKey,
	opsdk.ItemCategoryCryptoWallet:         "CRYPTO_WALLET",
	opsdk.ItemCategoryPerson:               "PERSON",
}

func categoryFromSDK(c opsdk.ItemCategory) onepassword.ItemCategory {
	if mapped, ok := sdkItemCategories[c]; ok {
		return mapped
	}
	return onepassword.Custom
}

func urlsFromSDK(websites []opsdk.Website) []onepassword.ItemURL {
	if len(websites) == 0 {
		return nil
	}
	urls := make([]onepassword.ItemURL, 0, len(websites))
	for _, w := range websites {
		urls = append(urls, onepassword.ItemURL{URL: w.URL, Label: w.Label})
	}
	return urls
}

func itemFromSDKOverview(o opsdk.ItemOverview) itemOverview {
	return itemOverview{
		ID:        o.ID,
		VaultID:   o.VaultID,
		Title:     o.Title,
		Category:  string(categoryFromSDK(o.Category)),
		State:     string(o.State),
		Tags:      o.Tags,
		URLs:      urlsFromSDK(o.Websites),
		CreatedAt: o.CreatedAt,
		UpdatedAt: o.UpdatedAt,
	}
}
