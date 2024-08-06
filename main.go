package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/charmbracelet/huh"

	"github.com/charmbracelet/huh/spinner"
)

type ApiResponse struct {
	Data map[string]float64 `json: "data"`
}

func main() {

	var currency1 string       //store the kind of currency in a string
	var currency2 string       //^
	var currencyAmount float64 //amount of currency user would convert (float for decimals)
	var currencyAmountStr string
	var confirmation bool
	var API_KEY string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Title("Please input your api key for freecurrencyapi.com").Value(&API_KEY),
		),
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Choose your currency").
				Options(

					huh.NewOption("$ USD", "USD"),
					huh.NewOption("$ CAD", "CAD"),
					huh.NewOption("€ Euro", "EUR"),
					huh.NewOption("¥ Yen", "JPY"),
				).
				Value(&currency1),
			huh.NewSelect[string]().
				Title("Choose the second currency you wish to convert to").
				Options(

					huh.NewOption("$ USD", "USD"),
					huh.NewOption("$ CAD", "CAD"),
					huh.NewOption("€ Euro", "EUR"),
					huh.NewOption("¥ Yen", "JPY"),
				).
				Value(&currency2),

			huh.NewInput().
				Title("What is the amount of money you'd like to convert?").
				Value(&currencyAmountStr).
				Validate(func(input string) error {
					_, err := strconv.ParseFloat(input, 64)
					if err != nil {
						return errors.New("You didn't input a valid number")
					}
					return nil
				}),

			huh.NewConfirm().Title("Does everything seem correct?").Value(&confirmation),
		),
	)

	err := form.Run()
	if err != nil {

		log.Fatal(err)
	}

	// Convert currencyAmountStr to currencyAmount after the form has been run
	currencyAmount, err = strconv.ParseFloat(currencyAmountStr, 64)
	if err != nil {
		log.Fatal("Failed to parse currency amount:", err)
	}

	if confirmation {

		s := spinner.New().Title("Processing your req...")
		done := make(chan bool)
		var convertedAmount float64

		go func() {

			url := fmt.Sprintf("https://api.freecurrencyapi.com/v1/latest?apikey=%s&currencies=%s&base_currency=%s", API_KEY, currency2, currency1)
			resp, err := http.Get(url)
			if err != nil {
				log.Fatal("Error making the request:", err)
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Fatal("Error reading the response:", err)

			}

			var response ApiResponse

			err = json.Unmarshal(body, &response)
			if err != nil {

				log.Fatal("Error parsing Json: ", err)
			}

			exchangeRate, ok := response.Data[currency2]
			if !ok {
				log.Fatalf("%s rate not found in response", currency2)
			}

			convertedAmount = currencyAmount * exchangeRate

			s.Run()
			done <- true
		}()
		<-done

		noteMsg := fmt.Sprintf("%.4f %s is equal to %.2f %s\n", currencyAmount, currency1, convertedAmount, currency2)
		resultForm := huh.NewForm(
			huh.NewGroup(
				huh.NewNote().Title("Conversion Result").Description(noteMsg),
			),
		)
		err = resultForm.Run()
		if err != nil {
			log.Fatal(err)
		}

	} else {

		fmt.Println("Please re-run and try again")
	}

}
