package privkey

import (
	"errors"
	"os"
	"sync"
)

var keyBytesProvider = func(sshKeyPath *string) ([]byte, error) {
	if sshKeyPath != nil {
		key, err := os.ReadFile(*sshKeyPath)
		if err != nil {
			return nil, errors.New("unable to read private key: " + *sshKeyPath)
		}
		return key, nil
	} else {
		return nil, errors.New("unable to identify private key location (nil config)")
	}
}

var providerMutex sync.RWMutex

// GetBytes returns the SSH private key bytes.
// On iOS, it calls the provider function that was set by the iOS app
// using SetKeyBytesProvider or SetKeyBytes.
func GetBytes(sshKeyPath *string) ([]byte, error) {
	providerMutex.RLock()
	provider := keyBytesProvider
	providerMutex.RUnlock()

	if provider == nil {
		return nil, errors.New("key bytes provider not set - call SetKeyBytes first")
	}

	keyBytes, err := provider(sshKeyPath)
	if err != nil {
		return nil, errors.New("unable to get key bytes from provider: " + err.Error())
	}

	if len(keyBytes) == 0 {
		return nil, errors.New("provider returned empty key bytes")
	}

	return keyBytes, nil
}

// SetBytes allows the iOS app to directly set the SSH key bytes.
func SetBytes(keyBytes []byte) {
	providerMutex.Lock()
	defer providerMutex.Unlock()
	keyBytesProvider = func(*string) ([]byte, error) {
		return keyBytes, nil
	}
}
