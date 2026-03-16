package telegram

import (
	"context"
	"fmt"
	"io/fs"
	"time"

	"telegram-api/internal/domain"

	"github.com/gotd/td/session"
	"github.com/gotd/td/session/tdesktop"
	"github.com/gotd/td/tg"
)

// validateTDataParams validates tdata import parameters.
func validateTDataParams(apiID int, apiHash string, tdataFiles map[string][]byte) error {
	if apiID <= 0 {
		return fmt.Errorf("invalid api_id")
	}
	if apiHash == "" {
		return fmt.Errorf("api_hash required")
	}
	if len(tdataFiles) == 0 {
		return fmt.Errorf("tdata files required")
	}
	return nil
}

// readTDataAccounts reads tdata accounts from files.
func readTDataAccounts(tdataFiles map[string][]byte) (tdesktop.Account, error) {
	tdFS := &memFS{files: tdataFiles}

	accounts, err := tdesktop.ReadFS(tdFS, nil)
	if err != nil {
		return tdesktop.Account{}, fmt.Errorf("%w: %v", domain.ErrTDataInvalid, err)
	}

	if len(accounts) == 0 {
		return tdesktop.Account{}, fmt.Errorf("%w: no accounts found", domain.ErrTDataInvalid)
	}

	return accounts[0], nil
}

// saveTDataSession saves tdata session to storage.
func (m *ClientManager) saveTDataSession(ctx context.Context, account tdesktop.Account, sessionID string) error {
	sessionData, err := session.TDesktopSession(account)
	if err != nil {
		return fmt.Errorf("%w: %v", domain.ErrTDataInvalid, err)
	}

	persistentStorage := NewPersistentSessionStorage(m.crypter, m.repo, sessionID)

	loader := session.Loader{
		Storage: persistentStorage,
	}

	if err := loader.Save(ctx, sessionData); err != nil {
		return fmt.Errorf("%w: %v", domain.ErrTDataInvalid, err)
	}

	return nil
}

// verifyTDataSession verifies the tdata session and returns user info.
func (m *ClientManager) verifyTDataSession(
	ctx context.Context,
	apiID int,
	apiHash, sessionName, sessionID string,
) (*TGUser, error) {
	persistentStorage := NewPersistentSessionStorage(m.crypter, m.repo, sessionID)
	client := m.newClient(apiID, apiHash, sessionName, persistentStorage)

	var user *TGUser

	err := client.Run(ctx, func(ctx context.Context) error {
		self, err := client.API().UsersGetFullUser(ctx, &tg.InputUserSelf{})
		if err != nil {
			return fmt.Errorf("verify session: %w", err)
		}

		u, ok := self.Users[0].(*tg.User)
		if !ok {
			return fmt.Errorf("unexpected user type")
		}

		user = &TGUser{ID: u.ID, Username: u.Username}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrTDataInvalid, err)
	}

	return user, nil
}

// ImportTData imports Telegram Desktop session from tdata files.
func (m *ClientManager) ImportTData(
	ctx context.Context,
	apiID int,
	apiHash string,
	sessionName string,
	sessionID string,
	tdataFiles map[string][]byte,
) (*TGUser, error) {
	if err := validateTDataParams(apiID, apiHash, tdataFiles); err != nil {
		return nil, err
	}

	account, err := readTDataAccounts(tdataFiles)
	if err != nil {
		return nil, err
	}

	if err := m.saveTDataSession(ctx, account, sessionID); err != nil {
		return nil, err
	}

	return m.verifyTDataSession(ctx, apiID, apiHash, sessionName, sessionID)
}

// memFS implements fs.FS for in-memory files.
type memFS struct {
	files map[string][]byte
}

func (m *memFS) Open(name string) (fs.File, error) {
	data, ok := m.files[name]
	if !ok {
		return nil, fmt.Errorf("file not found: %s", name)
	}
	return &memFile{name: name, data: data}, nil
}

// memFile implements fs.File for in-memory file.
type memFile struct {
	name string
	data []byte
	pos  int
}

func (f *memFile) Stat() (fs.FileInfo, error) {
	return &memFileInfo{name: f.name, size: int64(len(f.data))}, nil
}

func (f *memFile) Read(p []byte) (int, error) {
	if f.pos >= len(f.data) {
		return 0, fmt.Errorf("EOF")
	}
	n := copy(p, f.data[f.pos:])
	f.pos += n
	return n, nil
}

func (f *memFile) Close() error {
	return nil
}

// memFileInfo implements fs.FileInfo for in-memory file.
type memFileInfo struct {
	name string
	size int64
}

func (fi *memFileInfo) Name() string {
	return fi.name
}

func (fi *memFileInfo) Size() int64 {
	return fi.size
}

func (fi *memFileInfo) Mode() fs.FileMode {
	return 0
}

func (fi *memFileInfo) ModTime() time.Time {
	return time.Time{}
}

func (fi *memFileInfo) IsDir() bool {
	return false
}

func (fi *memFileInfo) Sys() interface{} {
	return nil
}
