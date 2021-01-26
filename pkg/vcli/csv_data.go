package vcli

import "fmt"

// CsvDataLoad - Load Funds, Voteplans and Proposals information into a SQLite3 ready file DB.
//
// vit-servicing-station-cli csv-data load
//        --db-url <db-url>          URL of the vit-servicing-station database to interact with
//        --funds <funds>            Path to the csv containing funds information
//        --proposals <proposals>    Path to the csv containing proposals information
//        --challenges <challenges>  Path to the csv containing challenges information
//        --voteplans <voteplans>    Path to the csv containing voteplans information
func CsvDataLoad(
	dbURL string,
	funds string,
	proposals string,
	challenges string,
	voteplans string,
) ([]byte, error) {
	if dbURL == "" {
		return nil, fmt.Errorf("parameter missing : %s", "dbURL")
	}
	if funds == "" {
		return nil, fmt.Errorf("parameter missing : %s", "funds")
	}
	if proposals == "" {
		return nil, fmt.Errorf("parameter missing : %s", "proposals")
	}
	if challenges == "" {
		return nil, fmt.Errorf("parameter missing : %s", "challenges")
	}
	if voteplans == "" {
		return nil, fmt.Errorf("parameter missing : %s", "voteplans")
	}

	arg := []string{"csv-data", "load",
		"--db-url", dbURL,
		"--funds", funds,
		"--proposals", proposals,
		"--challenges", challenges,
		"--voteplans", voteplans,
	}

	return vcli(nil, arg...)
}
