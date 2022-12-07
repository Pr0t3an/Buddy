package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"github.com/akamensky/argparse"
	"github.com/common-nighthawk/go-figure"
	ct "github.com/daviddengcn/go-colortext"
	_ "github.com/mattn/go-sqlite3"
	"github.com/olekukonko/tablewriter"
	"golang.design/x/clipboard"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var whereami string
var dbpath string
var setflag bool = false
var subflag bool = false
var id int
var Topic string
var item string
var tags string
var Syntax string
var subs string
var sub string

func executeshell(vCommand string) (vOutput string) {

	out, err := exec.Command("bash", "-c", vCommand).Output()
	if err != nil {
		fmt.Println("error in execution")
		//log.Fatal(err)

	}
	vOutput = string(out)
	return

}

func dbreader(filter string, cp int) {
	if _, err := os.Stat(dbpath); err != nil {
		ct.Foreground(ct.Red, true)
		fmt.Print("[!] File does not exist - Download from https://github.com/Pr0t3an/Buddy-Config" + " [y/n]: ")
		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		response = strings.ToLower(strings.TrimSpace(response))
		if response == "y" || response == "yes" {
			fmt.Println("[-] Attempting to install Predfetch - Prefetch Parser")
			var clonestring string = "git clone https://github.com/Pr0t3an/predfetch.git " + filepath.Dir(resolve(whereami))
			executeshell(clonestring)
			fmt.Println("[-] Re-checking DB...")
		}
		if response == "n" || response == "no" {
			log.Fatal(err)
			ct.ResetColor()

		} else {
			ct.Foreground(ct.Green, true)
			fmt.Println("[+] Db file found")
			ct.ResetColor()
		}
	}
	db, err := sql.Open("sqlite3", dbpath)
	if err != nil {
		ct.Foreground(ct.Red, true)
		fmt.Println("\n[!] Failed to open buddy.sqllite")
		ct.ResetColor()
		log.Fatal(err)
	}

	defer db.Close()

	rows, err := db.Query("SELECT * FROM checklists " + filter)

	if err != nil {
		log.Fatal(err)
	}

	defer rows.Close()

	table := tablewriter.NewWriter(os.Stdout)
	table.SetRowLine(true)
	table.SetHeader([]string{"ID", "Name", "Tags", "Syntax"})

	for rows.Next() {

		err = rows.Scan(&id, &Topic, &item, &tags, &Syntax, &subs, &sub)
		table.Append([]string{strconv.Itoa(id), Topic, tags, Syntax})
		if subs == "Y" {
			subflag = true
		}
		if err != nil {
			log.Fatal(err)
		}
	}
	table.Render()
	if setflag == true {
		if cp > 0 {
			if subflag == true {
				ct.Foreground(ct.Yellow, true)
				fmt.Println("\n[-] Need to set command variable(s)")
				ct.ResetColor()
				re := regexp.MustCompile(`<@(.*?)@>`)
				//fmt.Println(re.FindAllString(string1, -1))
				for _, match := range re.FindAllString(Syntax, -1) {
					fmt.Println(match)
					reader := bufio.NewReader(os.Stdin)
					fmt.Print("Enter text: ")
					text, _ := reader.ReadString('\n')
					Syntax = strings.Replace(Syntax, match, text, -1)
					Syntax = strings.Replace(Syntax, "\n", "", -1)
				}

			}
			clipboard.Write(clipboard.FmtText, []byte(Syntax))
			ct.Foreground(ct.Green, true)
			fmt.Println("\n[+] Copied Syntax to Clipboard - " + Syntax)
			ct.ResetColor()

		}
	}
}

//lazy

func getsetdbpath() {
	fs, _ := os.Executable()
	exeDir, err := filepath.Abs(fs)
	if err != nil {
		ct.Background(ct.Red, true)
		fmt.Println("[!] Error getting path to executable")
		ct.ResetColor()
	}

	// Construct the path to the database file
	fileInfo, err := os.Lstat(exeDir)

	//above finds path and is fine - use exeDir
	if fileInfo.Mode()&os.ModeSymlink == os.ModeSymlink {
		whereami = filepath.Dir(resolve(exeDir))

	} else {
		whereami = filepath.Dir(exeDir)
	}

	dbpath = whereami + "/config/buddy.sqllite"
}

func resolve(p string) string {
	cmd := exec.Command("readlink", "-fn", p)
	out, _ := cmd.Output()
	return (string(out))
}

// Define a struct to represent the data in the checklists table
type checklist struct {
	id     int
	Topic  string
	item   string
	tags   string
	Syntax string
	subs   string
	Sub    string
}

// Open a connection to the SQLite database
func openDB() *sql.DB {
	// Replace "database.db" with the path to your SQLite database file
	db, err := sql.Open("sqlite3", dbpath)
	if err != nil {
		log.Fatal(err)
	}

	return db
}

// Create a new record in the checklists table
func create(db *sql.DB, data checklist) error {
	// Use an INSERT statement to add the data to the checklists table
	_, err := db.Exec("INSERT INTO checklists (Topic, item, tags, Syntax, subs, Sub) VALUES (?, ?, ?, ?, ?, ?)",
		data.Topic, data.item, data.tags, data.Syntax, data.subs, data.Sub)

	return err
}

func updatedb() {
	// Open a connection to the SQLite database
	db := openDB()
	defer db.Close()

	// Prompt the user for the data to insert into the checklists table
	reader := bufio.NewReader(os.Stdin)
	ct.Foreground(ct.Green, true)
	fmt.Print("Enter Topic (general category): ")
	ct.ResetColor()

	fmt.Print("Enter Sub Topic: ")

	sub, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}
	sub = strings.TrimSpace(sub)

	topic, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}
	topic = strings.TrimSpace(topic)

	fmt.Print("Description: ")
	item, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}
	item = strings.TrimSpace(item)

	fmt.Print("Enter tags (comma separated): ")
	tags, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}
	tags = strings.TrimSpace(tags)

	fmt.Print("Enter Syntax (use <@var@> for variables): ")
	syntaxStr, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}
	syntaxStr = strings.TrimSpace(syntaxStr)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print("Does this command have subs? (Y/N): ")
	subs, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}
	subs = strings.TrimSpace(subs)

	// Create a checklist struct with the user-provided data
	data := checklist{
		Topic:  topic,
		item:   item,
		tags:   tags,
		Syntax: syntaxStr,
		subs:   subs,
		Sub:    sub,
	}

	// Insert the data into the checklists table
	err = create(db, data)
	if err != nil {
		log.Fatal(err)
	}
	ct.Foreground(ct.Green, true)
	fmt.Println("[+] Record added to Buddies DB")
	ct.ResetColor()
}
func delete(db *sql.DB, id int) error {
	// Use a DELETE statement to remove the record with the specified ID from the checklists table
	_, err := db.Exec("DELETE FROM checklists WHERE id = ?", id)

	return err
}

func deletebytopic() {
	db := openDB()
	defer db.Close()
	ct.Foreground(ct.Green, true)
	// Prompt the user for the ID of the record to delete
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter ID of record to delete: ")
	idStr, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}
	idStr = strings.TrimSpace(idStr)
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Fatal(err)
	}

	// Prompt the user for confirmation before deleting the record
	fmt.Print("Are you sure you want to delete this record? [y/n]: ")
	response, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}
	response = strings.ToLower(strings.TrimSpace(response))

	// If the user confirms, delete the record from the database
	if response == "y" || response == "yes" {
		err = delete(db, id)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Record successfully deleted from the database")
	}
	ct.ResetColor()
}

func main() {
	banner()
	getsetdbpath()
	fmt.Println("Path to db file is " + dbpath + " Dont forget to back it up")
	parser := argparse.NewParser("Buddy", "Buddy - your friendly command line companion")
	//checklist item - all shows all, specifying name or search checklists
	t := parser.String("t", "tag", &argparse.Options{Required: false, Help: "Search by tag"})
	u := parser.String("u", "use", &argparse.Options{Required: false, Help: "Use # to copy syntax to clipboard"})
	var addrecord *bool = parser.Flag("a", "add", &argparse.Options{Required: false, Help: "Guided - Add Record"})
	var deleterecord *bool = parser.Flag("d", "delete", &argparse.Options{Required: false, Help: "Guided - Delete Record by ID"})

	err := parser.Parse(os.Args)
	if err != nil {
		fmt.Print(parser.Usage(err))
	}
	//get checklist flags
	if len(*t) > 0 {
		ct.Foreground(ct.Green, true)
		fmt.Println("\n[+] Searching Tag - " + *t)
		ct.ResetColor()
		if *t == "all" {
			dbreader("", 0)
		} else {
			dbreader("WHERE tags LIKE '%"+*t+"%'", 0)
		}
	}

	if len(*u) > 0 {

		setflag = true
		ct.Foreground(ct.Green, true)
		fmt.Println("\n[+] Attempting to copy Syntax to Clipboard - " + *u)
		ct.ResetColor()
		intVar, err := strconv.Atoi(*u)
		if err == nil {
			dbreader("WHERE id = "+*u, intVar)
		}
	}

	if *deleterecord {
		deletebytopic()
	}

	if *addrecord {
		updatedb()
	}
}
func banner() {
	myFigure := figure.NewColorFigure("Buddy", "small", "green", true)
	myFigure.Print()
}
