package cmd

// TODO: 	* updateCmd to update records every 12 hours with the latest information - https://ctftime.org/event/list/upcoming
//				Check ID, if not in the upcoming list then delete record else update record.

import (
	"database/sql"
	"encoding/json"
	"encoding/xml"
	"strconv"
	"time"

	"github.com/RETr4ce/mudre/tools"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

//init function to the root program and add logging
func init() {
	RootCmd.AddCommand(ctftimeCmd)
	ctftimeCmd.AddCommand(ctftimeWriteupCmd, ctftimePullCmd, ctftimeUpcomingCmd, ctftimeUpdateCmd) // you can add subdommands to the root command.
}

//Createcommand ctftime
var ctftimeCmd = &cobra.Command{
	Use:   "ctftime",
	Short: "Module for parsing CTFTime website",
}

//Subcommand to post the upcoming data to Discord
var ctftimeUpcomingCmd = &cobra.Command{
	Use:   "upcoming",
	Short: "Posting upcoming",
	Long:  "Posting upcoming events 2 days prior to Discord",
	Run:   ctftUpcoming,
}

//Subcommand to post the writeups data to Discord
var ctftimeWriteupCmd = &cobra.Command{
	Use:   "writeup",
	Short: "Posting writeups",
	Long:  "Posting the latest event writeups to Discord",
	Run:   ctftWriteup,
}

//Subcommand to parse Events from CTFTime website
var ctftimePullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull latest events and writeups",
	Long:  "Pull and cache latest events and writeups into a SQLite database",
	Run:   ctftPull,
}

var ctftimeUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update latest events times",
	Long:  "Update with the latest pushed dates and times with the database",
	Run:   ctftUpdate,
}

//Checks if pushDiscord is not 1 before posting it to Discord. Safetycheck for duplicates
//Looks two days up ahead before posting it to Discord.
//This will give the user enough time to prepare or signup for the CTF. Except for prequalifications.
func ctftUpcoming(cmd *cobra.Command, args []string) {
	InfoLogger.Println("[-] Starting upcoming events")

	InfoLogger.Println("[+] Connecting to sqlite database")
	database, _ := sql.Open("sqlite3", viper.GetString("database-path"))

	InfoLogger.Println("[+] Prepairing SQL statements")
	tx, _ := database.Begin()

	stmt, err := tx.Prepare("UPDATE ctftimeEvents SET pushDiscord = 1 WHERE id = ?")

	if err != nil {
		ErrorLogger.Println("[*] ", err)
	}

	rows, err := tx.Query("select * from ctftimeEvents where DATE(ctftimeEvents.start) = DATE('now', '+2 day') and pushDiscord = 0;")

	if err != nil {
		ErrorLogger.Println("[*] ", err)
	}

	defer rows.Close()

	for rows.Next() {
		var Id int32
		var Title string
		var Link string
		var Start time.Time
		var Finish time.Time
		var Description string
		var Format string
		var Logo string
		var Restrictions string
		var Onsite bool
		var pushDiscord bool

		err = rows.Scan(&Id, &Title, &Link, &Start, &Finish, &Description, &Format, &Logo, &Restrictions, &Onsite, &pushDiscord)
		if err != nil {
			ErrorLogger.Println("[*] ", err)
		}
		var jsonData = json.RawMessage(`
		{
			"username": "CTFTime",
			"avatar_url": "` + viper.GetString("ctftime.ctftime-avatar") + `",
			"embeds": [{
				"color": 2021216,
				"title": "` + Title + `",
				"description": "URL: ` + Link + `",
				"thumbnail": {
					"url": "` + Logo + `"
				},
				"fields": [{
						"name": "Start",
						"value": "` + Start.Format(time.ANSIC) + `",
						"inline": true
					},
					{
						"name": "Finish",
						"value": "` + Finish.Format(time.ANSIC) + `",
						"inline": true
					}
				],
				"footer": {
					"text": "Upcoming events | Format: ` + Format + ` | Restrictions: ` + Restrictions + ` | On-site: ` + strconv.FormatBool(Onsite) + `"
				}
			}]
		}`)
		tools.PostToDiscord(jsonData, viper.GetString("ctftime.discord-webhook"))

		_, err := stmt.Exec(Id)

		if err != nil {
			ErrorLogger.Println("[*] ", err)
		}
	}
	InfoLogger.Println("[+] Executing SQL statements")
	tx.Commit()
	InfoLogger.Println("[-] Done")
}

//Checks if pushDiscord is not 1 before posting it to Discord. Safetycheck for duplicates
func ctftWriteup(cmd *cobra.Command, args []string) {
	InfoLogger.Println("[-] Starting latest writeups")

	InfoLogger.Println("[+] Connecting to sqlite database")
	database, _ := sql.Open("sqlite3", viper.GetString("database-path"))

	InfoLogger.Println("[+] Prepairing SQL statements")
	tx, _ := database.Begin()

	stmt, err := tx.Prepare("UPDATE ctftimeWriteup SET pushDiscord = 1 WHERE id = ?")

	if err != nil {
		ErrorLogger.Println("[*] ", err)
	}

	rows, err := tx.Query("select * from ctftimeWriteup where pushDiscord = 0;")

	if err != nil {
		ErrorLogger.Println("[*] ", err)
	}

	defer rows.Close()

	for rows.Next() {

		var id int
		var link string
		var title string
		var originalUrl string
		var lastBuild time.Time
		var pushDiscord bool

		err = rows.Scan(&id, &link, &title, &originalUrl, &lastBuild, &pushDiscord)

		if err != nil {
			ErrorLogger.Println("[*] ", err)
		}

		var jsonData = json.RawMessage(`{
				"username": "CTFTime",
				"avatar_url": "` + viper.GetString("ctftime.ctftime-avatar") + `",
				"embeds": [{
					"color": 10181046,
					"fields": [{
							"name": "` + title + `",
							"value": "` + link + `"
						}
					],
					"footer": {
						"text": "New writeup | ` + lastBuild.Format(time.ANSIC) + `"
					}
				}]
			}`)
		tools.PostToDiscord(jsonData, viper.GetString("ctftime.discord-webhook"))

		_, err := stmt.Exec(id)

		if err != nil {
			ErrorLogger.Println("[*] ", err)
		}

	}
	InfoLogger.Println("[+] Executing SQL statements")
	tx.Commit()
	InfoLogger.Println("[-] Done")

}

type dataEvent struct {
	Title        string    `json:"title"`
	Link         string    `json:"ctftime_url"`
	Start        time.Time `json:"start"`
	Finish       time.Time `json:"finish"`
	Description  string    `json:"description"`
	Format       string    `json:"format"`
	Id           int32     `json:"id"`
	Logo         string    `json:"logo"`
	Restrictions string    `json:"restrictions"`
	Onsite       bool      `json:"onsite"`
}

type dataWriteup struct {
	Channel struct {
		Item []struct {
			Title       string `xml:"title"`
			Link        string `xml:"link"`
			OriginalURL string `xml:"original_url"`
		} `xml:"item"`
	} `xml:"channel"`
}

//Pull the data from CTFTime and store it in the sqlite database.
//Use crontab to execute the command like once every 3 hours.
func ctftPull(cmd *cobra.Command, args []string) {

	InfoLogger.Println("[-] Start pulling data")
	InfoLogger.Println("[+] Connecting to sqlite database")
	database, _ := sql.Open("sqlite3", viper.GetString("database-path"))
	InfoLogger.Println("[+] Prepairing SQL statements")
	tx, _ := database.Begin()

	defer database.Close()

	pullEvent := func() {
		body, _ := tools.GetDataFromUrl(viper.GetString("ctftime.ctftime-event"))

		//unmarshal into a slice
		var dataObjects []dataEvent
		err := json.Unmarshal(body, &dataObjects)
		if err != nil {
			ErrorLogger.Println("[+] ", err)
		}

		for _, value := range dataObjects {

			//Quick and dirty by keeping a unique ID in the database and return failed if ID is already in the database.
			_, err := tx.Exec("INSERT INTO ctftimeEvents (id, title, link, start, finish, description, format, logo, restrictions, onsite, pushDiscord) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", value.Id, value.Title, value.Link, value.Start, value.Finish, value.Description, value.Format, value.Logo, value.Restrictions, value.Onsite, 0)

			if err != nil {
				WarningLogger.Println("[*] Data already excist in database: ", value.Title)
			}
		}
	}

	pullWriteup := func() {
		body, _ := tools.GetDataFromUrl(viper.GetString("ctftime.ctftime-writeup"))

		//unmarshal into a slice
		var dataObjects []dataWriteup
		err := xml.Unmarshal(body, &dataObjects)
		if err != nil {
			ErrorLogger.Println("[+] ", err)
		}

		for _, value := range dataObjects[0].Channel.Item {

			//Quick and dirty by keeping a unique ID in the database and return failed if ID is already in the database.
			_, err := tx.Exec("INSERT INTO ctftimeWriteup (link, title, originalUrl, lastBuild, pushDiscord) VALUES (?, ?, ?, ?, ?)", value.Link, value.Title, value.OriginalURL, time.Now(), 0)

			if err != nil {
				WarningLogger.Println("[*] Data already excist in database: ", value.Title)
			}
		}
	}

	InfoLogger.Println("[+] Pulling events")
	pullEvent()
	InfoLogger.Println("[+] Pulling writeup")
	pullWriteup()

	InfoLogger.Println("[+] Executing SQL statements")
	tx.Commit()
	InfoLogger.Println("[-] done")
}

func ctftUpdate(cmd *cobra.Command, args []string) {

	InfoLogger.Println("[-] Start updating data")
	InfoLogger.Println("[+] Connecting to sqlite database")
	database, _ := sql.Open("sqlite3", viper.GetString("database-path"))
	InfoLogger.Println("[+] Prepairing SQL statements")
	tx, _ := database.Begin()

	//Compare the dates and return the differences
	stmtCompare, _ := tx.Prepare("SELECT JULIANDAY(?) - JULIANDAY(start) AS date_difference FROM ctftimeEvents WHERE id=?")
	//If date is not 0 update the row with the latest dates
	stmtUpdate, _ := tx.Prepare("UPDATE ctftimeEvents SET start = ?, finish = ? WHERE id = ?")

	body, _ := tools.GetDataFromUrl(viper.GetString("ctftime.ctftime-event"))

	//Unmarshal into a slice
	var dataObjects []dataEvent
	err := json.Unmarshal(body, &dataObjects)

	if err != nil {
		ErrorLogger.Println("[+] ", err)
	}

	for _, value := range dataObjects {
		var date_difference float64

		//query row and compare the time differences
		err := stmtCompare.QueryRow(value.Start, value.Id).Scan(&date_difference)

		if err != nil {
			ErrorLogger.Println("[*] ", err)
		}

		// if time differences is not 0 then update with the latest times from ctftime
		if date_difference != 0 {
			_, err = stmtUpdate.Exec(value.Start, value.Finish, value.Id)

			if err != nil {
				ErrorLogger.Println("[*] ", err)
			}
		}
	}
	tx.Commit()
	InfoLogger.Println("[-] done")
}
