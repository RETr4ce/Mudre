package cmd

import (
	"database/sql"
	"encoding/json"
	"encoding/xml"
	"regexp"
	"strings"

	"github.com/RETr4ce/mudre/tools"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	RootCmd.AddCommand(crackmesCmd)
	crackmesCmd.AddCommand(crackmesPostCmd, crackmesPullCmd) // you can add subdommands to the root command.

}

//Createcommand crackmes
var crackmesCmd = &cobra.Command{
	Use:   "crackmes",
	Short: "Module for parsing the crackmes website",
}

//Subcommand to post the upcoming data to Discord
var crackmesPostCmd = &cobra.Command{
	Use:   "post",
	Short: "Post to Discord",
	Long:  "Posting the latest crackmes to the Discord channel",
	Run:   crackmesPost,
}

//Subcommand to parse Events from CTFTime website
var crackmesPullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull latest events and writeups",
	Long:  "Pull and cache latest events and writeups into a SQLite database",
	Run:   crackmesPull,
}

type crackmesDataEvents struct {
	Channel struct {
		Item []struct {
			Text        string `xml:",chardata"`
			Title       string `xml:"title"`
			Link        string `xml:"link"`
			Description string `xml:"description"`
			Author      string `xml:"author"`
			Category    string `xml:"category"`
			Guid        string `xml:"guid"`
			PubDate     string `xml:"pubDate"`
		} `xml:"item"`
	} `xml:"channel"`
}

//Pulling data from crackmes and pushing it into the SQLite database
func crackmesPull(cmd *cobra.Command, args []string) {
	InfoLogger.Println("[-] Start pulling data")
	InfoLogger.Println("[+] Connecting to sqlite database")
	database, _ := sql.Open("sqlite3", viper.GetString("database-path"))
	InfoLogger.Println("[+] Prepairing SQL statements")
	tx, _ := database.Begin()

	defer database.Close()

	//Getting the RSS feed from Crackmes and start pulling the last 50 submissions
	body, _ := tools.GetDataFromUrl(viper.GetString("crackmes.crackmes-latest"))

	var dataObjects []crackmesDataEvents
	err := xml.Unmarshal(body, &dataObjects)
	if err != nil {
		ErrorLogger.Println("[+] ", err)
	}

	for _, value := range dataObjects[0].Channel.Item {
		//Quick and dirty by keeping a unique ID in the database and return failed if ID is already in the database.
		_, err := tx.Exec("INSERT INTO crackmesLatest (link, title, description, author, category, guid,pubDate, pushDiscord) VALUES (?, ?, ?, ?, ?, ?, ?, ?)", value.Link, value.Title, value.Description, value.Author, value.Category, value.Guid, value.PubDate, 0)

		if err != nil {
			WarningLogger.Println("[*] Data already excist in database: ", value.Title)
		}
	}
	InfoLogger.Println("[+] Executing SQL statements")
	tx.Commit()
	InfoLogger.Println("[-] done")
}

func crackmesPost(cmd *cobra.Command, args []string) {
	InfoLogger.Println("[-] Starting upcoming events")

	InfoLogger.Println("[+] Connecting to sqlite database")
	database, _ := sql.Open("sqlite3", viper.GetString("database-path"))

	InfoLogger.Println("[+] Prepairing SQL statements")
	tx, _ := database.Begin()

	stmt, err := tx.Prepare("UPDATE crackmesLatest SET pushDiscord = 1 WHERE id = ?")

	if err != nil {
		ErrorLogger.Println("[*] ", err)
	}

	rows, err := tx.Query("select * from crackmesLatest where pushDiscord = 0;")

	if err != nil {
		ErrorLogger.Println("[*] ", err)
	}

	defer rows.Close()

	for rows.Next() {
		var id int
		var link string
		var title string
		var description string
		var author string
		var category string
		var guid string
		var pubDate string
		var pushDiscord bool

		err = rows.Scan(&id, &link, &title, &description, &author, &category, &guid, &pubDate, &pushDiscord)

		if err != nil {
			ErrorLogger.Println("[*] ", err)
		}

		expression := regexp.MustCompile(`(?:\[)(.*?)\s-\s(.*?)\s-\s(.*?)(?:\])`)
		regexResults := expression.FindStringSubmatch(title)

		var jsonData = json.RawMessage(`
		{
			"username": "Crackmes",
			"avatar_url": "` + viper.GetString("crackmes.crackmes-avatar") + `",
			"embeds": [{
			"author": {
			"name": "` + strings.Replace(title, regexResults[0], "", -1) + `"
		},
		"description": "` + link + `",
		"color": 10181046,
		"fields": [
			{
				"name": "category",
				"value": "` + regexResults[1] + `",
				"inline": true
			},
			{
				"name": "Language",
				"value": "` + regexResults[2] + `",
				"inline": true
			},
			{
				"name": "Difficulty",
				"value": "` + regexResults[3] + `",
				"inline": true
			}
			],
		"footer": {
			"text": "New crackme | Author: ` + author + ` | ` + pubDate + `"
				}
			}]
		}`)
		tools.PostToDiscord(jsonData, viper.GetString("crackmes.discord-webhook"))

		_, err := stmt.Exec(id)

		if err != nil {
			ErrorLogger.Println("[*] ", err)
		}

	}
	InfoLogger.Println("[+] Executing SQL statements")
	tx.Commit()
	InfoLogger.Println("[-] Done")
}
