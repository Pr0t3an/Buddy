package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"github.com/akamensky/argparse"
	"github.com/common-nighthawk/go-figure"
	ct "github.com/daviddengcn/go-colortext"
	_ "github.com/mattn/go-sqlite3"
	"golang.design/x/clipboard"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
)

func dbreader(filter string, cp int) {

	if _, err := os.Stat("configs/buddy.sqllite"); err != nil {
		ct.Foreground(ct.Red, true)
		fmt.Println("File does not exist - Download from https://github.com/Pr0t3an/Buddy-Config")
		log.Fatal(err)
		ct.ResetColor()

	} else {
		ct.Foreground(ct.Green, true)
		fmt.Println("[+] Db file found")
		ct.ResetColor()
	}

	db, err := sql.Open("sqlite3", "configs/buddy.sqllite")

	if err != nil {
		ct.Foreground(ct.Red, true)
		fmt.Println("\n[!] Failed to open buddy.sqllite - do you want to download this?")
		ct.ResetColor()
		log.Fatal(err)
	}

	defer db.Close()

	rows, err := db.Query("SELECT * FROM checklists " + filter)

	if err != nil {
		log.Fatal(err)
	}

	defer rows.Close()

	for rows.Next() {

		var id int
		var Topic string
		var item string
		var tags string
		var Syntax string
		var subs string
		var sub string

		err = rows.Scan(&id, &Topic, &item, &tags, &Syntax, &subs, &sub)

		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("%s | %d | %s | %s\n", sub, id, item, Syntax)
		if cp > 0 {
			if subs == "Y" {
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

func main() {
	banner()
	parser := argparse.NewParser("Buddy", "Buddy - your friendly command line companion")
	//checklist item - all shows all, specifying name or search checklists
	t := parser.String("t", "tag", &argparse.Options{Required: false, Help: "Search by tag"})
	u := parser.String("u", "use", &argparse.Options{Required: false, Help: "Use # to copy syntax to clipboard"})
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
		ct.Foreground(ct.Green, true)
		fmt.Println("\n[+] Attempting to copy Syntax to Clipboard - " + *u)
		ct.ResetColor()
		intVar, err := strconv.Atoi(*u)
		if err == nil {
			dbreader("WHERE id = "+*u, intVar)
		}
	}
}
func banner() {
	myFigure := figure.NewColorFigure("Buddy", "small", "green", true)
	myFigure.Print()
}
