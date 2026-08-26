package security

import (
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	passwordHashAlgorithm = "argon2id"
	passwordHashVersion   = argon2.Version
	passwordMemoryKiB     = 64 * 1024
	passwordIterations    = 3
	passwordParallelism   = 2
	passwordSaltLength    = 16
	passwordKeyLength     = 32

	passwordMaxMemoryKiB   = 256 * 1024
	passwordMaxIterations  = 16
	passwordMaxParallelism = 8
	passwordMaxSaltLength  = 64
	passwordMaxKeyLength   = 64
)

func HashPassword(password string) (string, error) {
	salt, err := secureRandomBytes(passwordSaltLength)
	if err != nil {
		return "", err
	}

	hash := argon2.IDKey([]byte(password), salt, passwordIterations, passwordMemoryKiB, passwordParallelism, passwordKeyLength)
	encoded := fmt.Sprintf("$%s$v=%d$m=%d,t=%d,p=%d$%s$%s",
		passwordHashAlgorithm,
		passwordHashVersion,
		passwordMemoryKiB,
		passwordIterations,
		passwordParallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	)

	return encoded, nil
}

func VerifyPassword(password, encoded string) (bool, error) {
	params, salt, expectedHash, err := parsePasswordHash(encoded)
	if err != nil {
		return false, err
	}

	derived := argon2.IDKey([]byte(password), salt, params.iterations, params.memoryKiB, params.parallelism, uint32(len(expectedHash)))
	if subtle.ConstantTimeCompare(derived, expectedHash) != 1 {
		return false, nil
	}

	return true, nil
}

type passwordHashParams struct {
	memoryKiB   uint32
	iterations  uint32
	parallelism uint8
}

func parsePasswordHash(encoded string) (passwordHashParams, []byte, []byte, error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[0] != "" {
		return passwordHashParams{}, nil, nil, errors.New("invalid password hash format")
	}
	if parts[1] != passwordHashAlgorithm {
		return passwordHashParams{}, nil, nil, errors.New("invalid password hash algorithm")
	}

	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return passwordHashParams{}, nil, nil, errors.New("invalid password hash version")
	}
	if version != passwordHashVersion {
		return passwordHashParams{}, nil, nil, errors.New("unsupported password hash version")
	}

	var params passwordHashParams
	for _, item := range strings.Split(parts[3], ",") {
		name, value, found := strings.Cut(item, "=")
		if !found {
			return passwordHashParams{}, nil, nil, errors.New("invalid password hash parameters")
		}

		parsed, err := strconv.ParseUint(value, 10, 32)
		if err != nil {
			return passwordHashParams{}, nil, nil, errors.New("invalid password hash parameters")
		}

		switch name {
		case "m":
			if parsed == 0 || parsed > passwordMaxMemoryKiB {
				return passwordHashParams{}, nil, nil, errors.New("unsupported password hash memory")
			}
			params.memoryKiB = uint32(parsed)
		case "t":
			if parsed == 0 || parsed > passwordMaxIterations {
				return passwordHashParams{}, nil, nil, errors.New("unsupported password hash iterations")
			}
			params.iterations = uint32(parsed)
		case "p":
			if parsed == 0 || parsed > passwordMaxParallelism {
				return passwordHashParams{}, nil, nil, errors.New("unsupported password hash parallelism")
			}
			params.parallelism = uint8(parsed)
		default:
			return passwordHashParams{}, nil, nil, errors.New("invalid password hash parameters")
		}
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return passwordHashParams{}, nil, nil, errors.New("invalid password hash salt")
	}
	if len(salt) != passwordSaltLength || len(salt) > passwordMaxSaltLength {
		return passwordHashParams{}, nil, nil, errors.New("unsupported password hash salt length")
	}

	expectedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return passwordHashParams{}, nil, nil, errors.New("invalid password hash value")
	}
	if len(expectedHash) != passwordKeyLength || len(expectedHash) > passwordMaxKeyLength {
		return passwordHashParams{}, nil, nil, errors.New("unsupported password hash length")
	}

	if params.memoryKiB == 0 || params.iterations == 0 || params.parallelism == 0 {
		return passwordHashParams{}, nil, nil, errors.New("incomplete password hash parameters")
	}

	return params, salt, expectedHash, nil
}
