package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/ulule/deepcopier"
)

type transaction struct {
	Date      time.Time
	Amount    float64
	Recurring string
}
type Config struct {
	Balance      float64
	BalanceDate  time.Time
	Transactions map[string]transaction
}

type Txns []transaction

func (s Txns) Len() int      { return len(s) }
func (s Txns) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

type ByDate struct{ Txns }

func (s ByDate) Less(i, j int) bool { return s.Txns[i].Date.Before(s.Txns[j].Date) }

var config Config

func sameDay(day, check time.Time) bool {
	return check.Format("2006-01-02") == day.Format("2006-01-02")
}

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("please specify config file")
		os.Exit(1)
	}

	if _, err := toml.DecodeFile(args[0], &config); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("Balance: $%.2f\n", config.Balance)
	fmt.Printf("Balance Date: %s\n", config.BalanceDate.Format("2006-01-02"))
	workingtxns := []transaction{}

	for transactionName, transaction := range config.Transactions {
		fmt.Printf("Transaction: %s (%s, $%.2f, %s)\n", transactionName, transaction.Date.Format("2006-01-02"), transaction.Amount, transaction.Recurring)
		transaction.Date = dateAfter(transaction.Recurring, transaction.Date, config.BalanceDate)
		workingtxns = append(workingtxns, transaction)
	}

	sort.Sort(ByDate{workingtxns})

	workingconfig := &Config{}
	deepcopier.Copy(config).To(workingconfig)
	endDate := config.BalanceDate.AddDate(2, 0, 0)
	fmt.Printf("End Date: %s\n", endDate.Format("2006-01-02"))
	var csvText bytes.Buffer

	for workingconfig.BalanceDate.Before(endDate) {
		//fmt.Printf("Working Date: %s\n", workingconfig.BalanceDate.Format("2006-01-02"))

		for sameDay(workingtxns[0].Date, workingconfig.BalanceDate) {
			workingconfig.Balance += workingtxns[0].Amount
			//fmt.Printf("Adding: %.2f\n", workingtxns[0].Amount)
			workingtxns[0].Date = nextDate(workingtxns[0].Recurring, workingtxns[0].Date)
			sort.Sort(ByDate{workingtxns})
		}

		fmt.Printf("%s - %.2f\n", workingconfig.BalanceDate.Format("2006-01-02"), workingconfig.Balance)
		fmt.Fprintf(&csvText, "%s,%.2f\n", workingconfig.BalanceDate.Format("2006-01-02"), workingconfig.Balance)
		workingconfig.BalanceDate = workingconfig.BalanceDate.AddDate(0, 0, 1)
	}

	fo, err := os.Create("output.csv")
	if err != nil {
		panic(err)
	}
	defer fo.Close()

	if _, err := fo.Write(csvText.Bytes()); err != nil {
		panic(err)
	}

}

func nextDate(recurring string, original time.Time) time.Time {

	returnDate := original

	switch strings.ToLower(recurring) {
	case "daily":
		returnDate = original.AddDate(0, 0, 1)
	case "weekly":
		returnDate = original.AddDate(0, 0, 7)
	case "fortnightly":
		returnDate = original.AddDate(0, 0, 14)
	case "monthly":
		returnDate = original.AddDate(0, 1, 0)
	case "quarterly":
		returnDate = original.AddDate(0, 3, 0)
	case "yearly":
		returnDate = original.AddDate(1, 0, 0)
	default:
		panic(fmt.Sprintf("No recurring available for %s", recurring))
	}

	return returnDate
}

func dateAfter(recurring string, original, after time.Time) time.Time {
	returnDate := original
	for returnDate.Before(after) {
		returnDate = nextDate(recurring, returnDate)
	}
	return returnDate
}
