package vcli

import (
	"fmt"
	"strconv"
)

// ApiTokenGenerate - generate API tokens, URL safe base64 encoded.
//
// vit-servicing-station-cli api-token generate [--n <Number of tokens to generate [default: 1]>] [--size <Size of the token [default: 10]>] | STDOUT
func ApiTokenGenerate(
	n int,
	size int,
) ([]byte, error) {
	arg := []string{"api-token", "generate"}
	if n > 0 {
		arg = append(arg, "--n", strconv.Itoa(n))
	}
	if size > 0 {
		arg = append(arg, "--size", strconv.Itoa(size))
	}

	return vcli(nil, arg...)
}

// ApiTokenAdd - add provided tokens to database.
//
// vit-servicing-station-cli api-token add --db-url <URL of the vit-servicing-station database to interact with> [--tokens list of tokens in URL safe base64.]
func ApiTokenAdd(
	stdin []byte,
	dbURL string,
	tokens []string,
) ([]byte, error) {
	if len(stdin) == 0 && len(tokens) == 0 {
		return nil, fmt.Errorf("%s : EMPTY and parameter missing : %s", "stdin", "tokens")
	}
	if dbURL == "" {
		return nil, fmt.Errorf("parameter missing : %s", "dbURL")
	}

	arg := []string{"api-token", "add", "--db-url", dbURL}
	for _, token := range tokens {
		arg = append(arg, "--tokens", token)
	}

	return vcli(stdin, arg...)
}
