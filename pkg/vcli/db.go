package vcli

import "fmt"

// DbInit - Initialize a DB with the proper migrations, DB file is created if not exists.
//
// vit-servicing-station-cli db init --db-url <URL of the vit-servicing-station database to interact with>
func DbInit(
	dbURL string,
) ([]byte, error) {
	if dbURL == "" {
		return nil, fmt.Errorf("parameter missing : %s", "dbURL")
	}

	arg := []string{"db", "init", "--db-url", dbURL}

	return vcli(nil, arg...)
}
