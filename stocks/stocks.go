package stocks

import (
	"BetterScorch/database"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

var Mutex sync.Mutex

func ModifyCompanyValue(name string, delta int) error {
	// Fetch the company's current value
	current, err := database.Get("Company", []string{"value"}, &database.DBValue{
		Name:  "pk_name",
		Value: name,
	})
	if err != nil {
		return err
	}

	intCurrent, _ := strconv.Atoi(current[0][0])

	// Update the company's value
	newValue := intCurrent + delta

	// If the new value is negative, throw an error
	if newValue < 0 {
		return fmt.Errorf("company value cannot be negative")
	}

	// Get all shareholder amounts for this company
	shareholders, err := database.Get("Shareholder", []string{"pkfk_user_owns", "amount"}, &database.DBValue{
		Name:  "pkfk_company_belongsTo",
		Value: name,
	})
	if err != nil {
		return err
	}

	// Calculate the scaling factor (new value / old total)
	scalingFactor := float64(newValue) / float64(intCurrent)
	fmt.Println(scalingFactor)

	// Update each shareholder's amount based on the scaling factor
	total := 0
	for _, shareholder := range shareholders {
		user := shareholder[0] // user ID is the first column
		amt, _ := strconv.Atoi(shareholder[1])

		// Calculate new amount for this shareholder
		newAmount := int(float64(amt) * scalingFactor)
		total += newAmount

		// Update the shareholder's amount in the database
		err := database.Update("Shareholder",
			[]*database.DBValue{{Name: "pkfk_user_owns", Value: user}, {Name: "pkfk_company_belongsTo", Value: name}},
			&database.DBValue{Name: "amount", Value: strconv.Itoa(newAmount)},
		)
		if err != nil {
			return err
		}
	}

	err = database.Update("Company",
		[]*database.DBValue{{Name: "pk_name", Value: name}},
		&database.DBValue{Name: "value", Value: strconv.Itoa(total)},
	)
	if err != nil {
		return err
	}

	return nil
}

func BuyShares(user string, company string, delta int) (int, error) {
	// Fetch the company's current value
	current, err := database.Get("Company", []string{"value"}, &database.DBValue{
		Name:  "pk_name",
		Value: company,
	})
	if err != nil {
		return -1, err
	}

	intCurrent, _ := strconv.Atoi(current[0][0])

	// Update the company's value
	newValue := intCurrent + delta

	// If the new value is negative, throw an error
	if newValue < 0 {
		return -1, fmt.Errorf("company value cannot be negative")
	}

	err = database.Update("Company",
		[]*database.DBValue{{Name: "pk_name", Value: company}},
		&database.DBValue{Name: "value", Value: strconv.Itoa(newValue)},
	)
	if err != nil {
		return -1, err
	}

	_, err = ModifyBalance(user, -delta)
	if err != nil {
		return -1, err
	}

	currentShares, err := database.Get("Shareholder", []string{"amount"}, []*database.DBValue{{Name: "pkfk_company_belongsTo", Value: company}, {Name: "pkfk_user_owns", Value: user}}...)
	if err != nil {
		return -1, err
	}

	var currentSharesInt int
	if len(currentShares) == 0 {
		currentSharesInt = 0
	} else {
		currentSharesInt, _ = strconv.Atoi(currentShares[0][0])
	}

	if currentSharesInt < -delta && delta < 0 {
		return -1, fmt.Errorf("You can't sell more stocks than you have")
	}

	err = database.InsertOrUpdate("Shareholder",
		[]*database.DBValue{{Name: "pkfk_company_belongsTo", Value: company}, {Name: "pkfk_user_owns", Value: user}},
		&database.DBValue{Name: "amount", Value: strconv.Itoa(currentSharesInt + delta)},
	)
	if err != nil {
		return -1, err
	}

	return currentSharesInt + delta, nil
}

func ModifyBalance(user string, amount int) (int, error) {
	currentVal, err := database.Get("ScorchCoin", []string{"balance"}, &database.DBValue{Name: "pk_user", Value: user})
	if err != nil {
		if strings.Contains(err.Error(), "index out of range") {
			return -1, fmt.Errorf("You are not registered in the economy. Please use /entereconomy first")
		}
		return -1, err
	}

	if len(currentVal) == 0 {
		return -1, fmt.Errorf("You are not registered in the economy. Please use /entereconomy first")
	}

	intVal, err := strconv.Atoi(currentVal[0][0])

	return amount + intVal, database.Update("ScorchCoin", []*database.DBValue{{Name: "pk_user", Value: user}}, &database.DBValue{Name: "balance", Value: strconv.Itoa(amount + intVal)})
}

func GetBalance(user string) (int, error) {
	currentVal, err := database.Get("ScorchCoin", []string{"balance"}, &database.DBValue{Name: "pk_user", Value: user})
	if err != nil {
		if strings.Contains(err.Error(), "index out of range") {
			return -1, fmt.Errorf("You are not registered in the economy. Please use /entereconomy first")
		}
		return -1, err
	}

	if len(currentVal) == 0 {
		return -1, fmt.Errorf("You are not registered in the economy. Please use /entereconomy first")
	}

	return strconv.Atoi(currentVal[0][0])

}

func GetPortfolio(user string) (map[string]int, error) {
	rows, err := database.Get("Shareholder", []string{"pkfk_company_belongsTo", "amount"}, &database.DBValue{
		Name:  "pkfk_user_owns",
		Value: user,
	})
	if err != nil {
		return nil, err
	}

	result := make(map[string]int)
	for _, row := range rows {
		company := row[0]
		amt, _ := strconv.Atoi(row[1])
		result[company] = amt
	}

	return result, nil
}

func GetCompanyValue(companyName string) (int, error) {
	result, err := database.Get("Company", []string{"value"}, &database.DBValue{
		Name:  "pk_name",
		Value: companyName,
	})
	if err != nil {
		return 0, err
	}
	if len(result) == 0 || len(result[0]) == 0 {
		return 0, fmt.Errorf("company not found")
	}

	companyValue, _ := strconv.Atoi(result[0][0])

	return companyValue, nil
}

func Enter(user string) error {
	_, err := database.Insert("ScorchCoin", &database.DBValue{Name: "pk_user", Value: user}, &database.DBValue{Name: "balance", Value: "420"})
	return err
}

func RegularHandler() {
	const maxPoints = 300

	executionLine := plotter.XYs{}
	reviveLine := plotter.XYs{}
	gambleLine := plotter.XYs{}

	for {
		now := float64(time.Now().Unix())

		// Update and append latest values
		if value, _ := GetCompanyValue("Execution Solutions LLC"); true {
			executionLine = append(executionLine, plotter.XY{X: now, Y: float64(value)})
			if len(executionLine) > maxPoints {
				executionLine = executionLine[len(executionLine)-maxPoints:]
			}
		}

		if value, _ := GetCompanyValue("Revival Technologies"); true {
			reviveLine = append(reviveLine, plotter.XY{X: now, Y: float64(value)})
			if len(reviveLine) > maxPoints {
				reviveLine = reviveLine[len(reviveLine)-maxPoints:]
			}
		}

		if value, _ := GetCompanyValue("Gambling Inc"); true {
			gambleLine = append(gambleLine, plotter.XY{X: now, Y: float64(value)})
			if len(gambleLine) > maxPoints {
				gambleLine = gambleLine[len(gambleLine)-maxPoints:]
			}
		}

		// Plot main chart
		p := plot.New()
		p.Title.Text = "Stonks"
		p.X.Label.Text = "Time"
		p.Y.Label.Text = "ScorchCoin"
		p.X.Tick.Marker = plot.TimeTicks{Format: "15:04"}

		plotutil.AddLinePoints(p,
			"Execution Solutions LLC", executionLine,
			"Revival Technologies", reviveLine,
		)

		// Plot gambling chart
		p2 := plot.New()
		p2.Title.Text = "Gamble stonks"
		p2.X.Label.Text = "Time"
		p2.Y.Label.Text = "ScorchCoin"
		p2.X.Tick.Marker = plot.TimeTicks{Format: "15:04"}

		plotutil.AddLinePoints(p2,
			"Gambling Inc", gambleLine,
		)

		if err := p.Save(10*vg.Inch, 4*vg.Inch, "multiline.png"); err != nil {
			panic(err)
		}
		if err := p2.Save(10*vg.Inch, 4*vg.Inch, "gamble.png"); err != nil {
			panic(err)
		}

		time.Sleep(15 * time.Minute)
	}
}
