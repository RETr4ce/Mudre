package cmd

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/RETr4ce/mudre/tools"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	RootCmd.AddCommand(githubCmd)
	githubCmd.AddCommand(githubPullCmd, githubPushCmd) // you can add subdommands to the root command.
}

//Createcommand crackmes
var githubCmd = &cobra.Command{
	Use:   "github",
	Short: "Module for parsing the latest starred github from user account",
}

//Subcommand to post the upcoming data to Discord
var githubPushCmd = &cobra.Command{
	Use:   "push",
	Short: "push to Discord",
	Long:  "push the latest starred github to the Discord channel",
	Run:   githubPush,
}

//Subcommand to parse Events from CTFTime website
var githubPullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull latest starred",
	Long:  "Pull and cache latest starred into a SQLite database",
	Run:   githubPull,
}

type githubDataEvents struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	HTMLURL     string    `json:"html_url"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	Language    string    `json:"language"`
}

func githubPull(cmd *cobra.Command, args []string) {
	InfoLogger.Println("[-] Start pulling data")
	InfoLogger.Println("[+] Connecting to sqlite database")
	InfoLogger.Println("[+] ", viper.GetString("database-path"))
	database, _ := sql.Open("sqlite3", viper.GetString("database-path"))

	InfoLogger.Println("[+] Prepairing SQL statements")
	tx, _ := database.Begin()

	//Getting the RSS feed from Crackmes and start pulling the last 50 submissions
	body, _ := tools.GetDataFromUrl(viper.GetString("github.github-starred"))

	var dataObjects []githubDataEvents
	err := json.Unmarshal(body, &dataObjects)
	if err != nil {
		ErrorLogger.Println("[+] ", err)
	}

	for _, value := range dataObjects {
		//Quick and dirty by keeping a unique ID in the database and return failed if ID is already in the database.
		_, err := tx.Exec("INSERT INTO github (id, name, htmlurl, description, createdAt, language, pushDiscord) VALUES (?, ?, ?, ?, ?, ?, ?)", value.ID, value.Name, value.HTMLURL, value.Description, value.CreatedAt, value.Language, 0)

		if err != nil {
			WarningLogger.Println("[*] Data already excist in database: ", value.Name)
		}
	}
	InfoLogger.Println("[+] Executing SQL statements")
	tx.Commit()
	InfoLogger.Println("[-] done")
}

func githubPush(cmd *cobra.Command, args []string) {
	InfoLogger.Println("[-] Starting upcoming events")

	InfoLogger.Println("[+] Connecting to sqlite database")
	database, _ := sql.Open("sqlite3", viper.GetString("database-path"))

	InfoLogger.Println("[+] Prepairing SQL statements")
	tx, _ := database.Begin()

	stmt, err := tx.Prepare("UPDATE github SET pushDiscord = 1 WHERE id = ?")

	if err != nil {
		ErrorLogger.Println("[*] ", err)
	}

	rows, err := tx.Query("select * from github where pushDiscord = 0;")

	if err != nil {
		ErrorLogger.Println("[*] ", err)
	}

	defer rows.Close()

	for rows.Next() {
		var id int
		var name string
		var htmlUrl string
		var description string
		var createdAt string
		var language string
		var pushDiscord bool

		err = rows.Scan(&id, &name, &htmlUrl, &description, &createdAt, &language, &pushDiscord)

		if err != nil {
			ErrorLogger.Println("[*] ", err)
		}

		var jsonData = json.RawMessage(`
		{
			"username": "Github",
			"avatar_url": "` + viper.GetString("github.github-avatar") + `",
			"embeds": [{
			"title": "` + htmlUrl + `",
			"url": "` + htmlUrl + `",
			"author": {
			"name": "` + name + `"
		},
		"description": "` + description + `",
		"color": 10181046,

		"footer": {
			"text": "Language: ` + language + ` | ` + createdAt + `"
				}
			}]
		}`)
		tools.PostToDiscord(jsonData, viper.GetString("github.discord-webhook"))

		_, err := stmt.Exec(id)

		if err != nil {
			ErrorLogger.Println("[*] ", err)
		}

	}
	InfoLogger.Println("[+] Executing SQL statements")
	tx.Commit()
	InfoLogger.Println("[-] Done")
}
