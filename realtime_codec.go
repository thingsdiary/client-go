package client

import (
	"time"

	"github.com/pkg/errors"
	"github.com/samber/mo"

	v1 "github.com/thingsdiary/client-go/proto"
)

func protoTime(t interface{ AsTime() time.Time }) time.Time {
	if t == nil {
		return time.Time{}
	}

	return t.AsTime()
}

func protoString(value mo.Option[string]) *string {
	if !value.IsPresent() {
		return nil
	}

	v := value.MustGet()

	return &v
}

func protoBool(v bool) *bool { return &v }

func decryptProtoDetails(
	encryption *v1.Encryption,
	details *v1.EncryptedData,
	diaryKey []byte,
	name string,
) ([]byte, error) {
	if encryption == nil {
		return nil, errors.New("encryption is required")
	}
	if details == nil {
		return nil, errors.Errorf("%s is required", name)
	}

	entityKey, err := decryptWithSymmetricKey(
		encryption.EncryptedKeyNonce,
		encryption.EncryptedKeyData,
		diaryKey,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decrypt entity key")
	}

	decryptedDetails, err := decryptWithSymmetricKey(
		details.Nonce,
		details.Data,
		entityKey,
	)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to decrypt %s", name)
	}

	return decryptedDetails, nil
}

func validateProtoEncryption(value *v1.Encryption) error {
	if value == nil {
		return errors.New("encryption is required")
	}
	if value.DiaryKeyId == "" {
		return errors.New("diary key id is required")
	}
	if len(value.EncryptedKeyNonce) == 0 {
		return errors.New("encrypted key nonce is required")
	}
	if len(value.EncryptedKeyData) == 0 {
		return errors.New("encrypted key data is required")
	}

	return nil
}

func validateProtoEncryptedData(value *v1.EncryptedData, name string) error {
	if value == nil {
		return errors.Errorf("%s is required", name)
	}
	if len(value.Nonce) == 0 {
		return errors.Errorf("%s nonce is required", name)
	}
	if len(value.Data) == 0 {
		return errors.Errorf("%s data is required", name)
	}

	return nil
}
