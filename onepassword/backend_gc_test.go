package onepassword

import (
	"context"
	"errors"
	"runtime"
	"sync/atomic"
	"testing"
	"time"

	opsdk "github.com/1password/onepassword-sdk-go"
)

var errReleasedClient = errors.New("an internal error occurred, please contact 1Password: invalid client id")

// releasableItems fails like the SDK's Items handle once the owning client has been released.
type releasableItems struct {
	released *atomic.Bool
}

func (r releasableItems) List(context.Context, string, ...opsdk.ItemListFilter) ([]opsdk.ItemOverview, error) {
	if r.released.Load() {
		return nil, errReleasedClient
	}
	return []opsdk.ItemOverview{{ID: "i1", VaultID: "v1"}}, nil
}

// newReleasableClient builds a client whose finalizer marks its handles released, as the SDK's does.
func newReleasableClient() (*opsdk.Client, *atomic.Bool) {
	released := &atomic.Bool{}
	client := &opsdk.Client{
		VaultsAPI: nil,
		ItemsAPI:  releasableItemsAPI{releasableItems{released: released}},
	}
	runtime.SetFinalizer(client, func(*opsdk.Client) { released.Store(true) })
	return client, released
}

// releasableItemsAPI widens releasableItems to the SDK's ItemsAPI so it can sit on an opsdk.Client.
type releasableItemsAPI struct {
	releasableItems
}

func (releasableItemsAPI) Create(context.Context, opsdk.ItemCreateParams) (opsdk.Item, error) {
	panic("not used")
}
func (releasableItemsAPI) CreateAll(context.Context, string, []opsdk.ItemCreateParams) (opsdk.ItemsUpdateAllResponse, error) {
	panic("not used")
}
func (releasableItemsAPI) Get(context.Context, string, string) (opsdk.Item, error) {
	panic("not used")
}
func (releasableItemsAPI) GetAll(context.Context, string, []string) (opsdk.ItemsGetAllResponse, error) {
	panic("not used")
}
func (releasableItemsAPI) Put(context.Context, opsdk.Item) (opsdk.Item, error) { panic("not used") }
func (releasableItemsAPI) Delete(context.Context, string, string) error        { panic("not used") }
func (releasableItemsAPI) DeleteAll(context.Context, string, []string) (opsdk.ItemsDeleteAllResponse, error) {
	panic("not used")
}
func (releasableItemsAPI) Archive(context.Context, string, string) error { panic("not used") }
func (releasableItemsAPI) Shares() opsdk.ItemsSharesAPI                  { panic("not used") }
func (releasableItemsAPI) Files() opsdk.ItemsFilesAPI                    { panic("not used") }

// collectUntil runs the garbage collector until released is set or the attempts run out.
func collectUntil(released *atomic.Bool) {
	for i := 0; i < 50 && !released.Load(); i++ {
		runtime.GC()
		runtime.GC()
		time.Sleep(2 * time.Millisecond)
	}
}

func TestSDKBackendKeepsClientAliveAcrossGC(t *testing.T) {
	client, released := newReleasableClient()
	backend := newSDKBackend(client)
	client = nil
	_ = client

	collectUntil(released)

	if released.Load() {
		t.Fatal("SDK client was finalized while the cached backend still uses its handles")
	}
	if _, err := backend.ListItems(context.Background(), "v1"); err != nil {
		t.Fatalf("ListItems after GC: %v", err)
	}
	runtime.KeepAlive(backend)
}

func TestSDKHandlesAloneDoNotKeepClientAlive(t *testing.T) {
	client, released := newReleasableClient()
	handlesOnly := sdkBackend{items: client.Items()}
	client = nil
	_ = client

	collectUntil(released)

	if !released.Load() {
		t.Skip("garbage collector did not finalize the client; the control could not run")
	}
	if _, err := handlesOnly.ListItems(context.Background(), "v1"); !errors.Is(err, errReleasedClient) {
		t.Fatalf("ListItems = %v, want the released-client error", err)
	}
}
