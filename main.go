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
	//api response is structured as: data: so looking for that
	Data map[string]float64 `json: "data"`
}

func main() {

	//variables for use in program
	var currency1 string       //store the kind of currency in a string
	var currency2 string       //^
	var currencyAmount float64 //amount of currency user would convert (float for decimals)
	var currencyAmountStr string
	var confirmation bool //
	var API_KEY string //store api key inputed by user temp (maybe semi-permanent solution later?)


	//creating new huh form
	form := huh.NewForm(
		//group to handle api key input 
		huh.NewGroup(
			huh.NewInput().Title("Please input your api key for freecurrencyapi.com").Value(&API_KEY),
		),
		//group for business logic of TUI
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Choose your currency").
				Options( //can add more options here

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
				//had some troubles getting float64 to work, this is my work around
				//validate that the string can be converted into a number, if not then throw bad input
				Validate(func(input string) error {
					//strconv library to convert to float
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

	//everything hinges on if user clicks that everything seems correct, if they click no then no http req is sent to API
	if confirmation {
		//create HUH spinner class for user experience
		s := spinner.New().Title("Processing your req...")
		done := make(chan bool) //make new go routine channel to display spinner while handling api in back
		var convertedAmount float64

		go func() {
			// store url in a variable, %s %s %s are the three variables taken from the user (currency, currency2, api key)
			url := fmt.Sprintf("https://api.freecurrencyapi.com/v1/latest?apikey=%s&currencies=%s&base_currency=%s", API_KEY, currency2, currency1)
			
			//send off the get req
			resp, err := http.Get(url)
			if err != nil {
				log.Fatal("Error making the request:", err)
			}
			//defer makes it so that once everything else is finished it closes the response
			defer resp.Body.Close()
			//create body variable to store response body
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Fatal("Error reading the response:", err)

			}

			var response ApiResponse
			//parse response from apiresponse and put it in map from apiresponsestruct
			err = json.Unmarshal(body, &response)
			if err != nil {

				log.Fatal("Error parsing Json: ", err)
			}

			//if everything ok, then store the conversion rate in the exchangerate variable
			//it should bne noted that its looking for the float64 value with the key of the 
			//response (so if i wanted to convert to JPY, then this is looking for the float value that has key JPY

			exchangeRate, ok := response.Data[currency2]
			if !ok {
				log.Fatalf("%s rate not found in response", currency2)
			}

			//take currency amount and times it by the exchange rate
			convertedAmount = currencyAmount * exchangeRate

			//run spinner 
			s.Run()
			//when done quit the routine and stop the spinner
			done <- true
		}()
		<-done

		noteMsg := fmt.Sprintf("%.4f %s is equal to %.2f %s\n", currencyAmount, currency1, convertedAmount, currency2)
		//create new form to display result in a nice format
		resultForm := huh.NewForm(
			huh.NewGroup(
				huh.NewNote().Title("Conversion Result").Description(noteMsg),
			),
		)
		//we gotta run this after the api response is done
		err = resultForm.Run()
		if err != nil {
			log.Fatal(err)
		}

	} else {

		fmt.Println("Please re-run and try again")
	}

}
