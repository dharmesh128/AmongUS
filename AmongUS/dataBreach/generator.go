package dataBreach

import (
	"strings"
)

const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

// Encrypt encrypts plaintext using the Caesar cipher with the specified key
func Encrypt(plaintext string, key string) string {
	key = strings.ToUpper(key)
	var ciphertext strings.Builder

	for i, char := range plaintext {
		shift := int(key[i%len(key)] - 'A')
		if char >= 'A' && char <= 'Z' {
			index := (int(char-'A') + shift) % 26
			ciphertext.WriteByte(alphabet[index])
		} else {
			ciphertext.WriteRune(char) // Keep non-alphabetic characters unchanged
		}

		// Advance key index only for alphabetic characters
		if char >= 'A' && char <= 'Z' {
			keyIndex := (i / len(key)) % len(key)
			shift := int(key[keyIndex] - 'A')
			key = key[:keyIndex] + string(alphabet[(shift+1)%26]) + key[keyIndex+1:]
		}
	}

	return ciphertext.String()
}

// Decrypt decrypts plaintext using the Caesar cipher with the specified key
func Decrypt(ciphertext string, key string) string {
	key = strings.ToUpper(key)
	var plaintext strings.Builder

	for i, char := range ciphertext {
		shift := int(key[i%len(key)] - 'A')
		if char >= 'A' && char <= 'Z' {
			index := (int(char-'A') - shift + 26) % 26
			plaintext.WriteByte(alphabet[index])
		} else {
			plaintext.WriteRune(char) // Keep non-alphabetic characters unchanged
		}

		// Advance key index only for alphabetic characters
		if char >= 'A' && char <= 'Z' {
			keyIndex := (i / len(key)) % len(key)
			shift := int(key[keyIndex] - 'A')
			key = key[:keyIndex] + string(alphabet[(shift+1)%26]) + key[keyIndex+1:]
		}
	}

	return plaintext.String()
}
