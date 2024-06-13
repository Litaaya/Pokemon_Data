// Lưu ý, để chạy được file vui lòng tạo một folder pokedex, tạo main.go, cop đống code này vô rồi bật terminal lên
// Gõ: cd pokedex -> xong rồi gõ, mod init pokedex
// Sau đó tải 2 gói github ở dưới với câu lệnh: go get github.com/chromedp/chromedp và câu go get github.com/gocolly/colly
// Cấu trúc của file sẽ như vầy:
// pokedex
// -go.mod
// -go.sum
// -main.go
// pokedex.json      -----> file này sẽ đc tạo khi chạy câu lệnh: go run main.go để bắt đầu đào data

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/chromedp/chromedp" //gói để điều khiển trình duyệt Chrome không có giao diện (headless)
	"github.com/gocolly/colly" //Gói để lấy data
)

// Chỗ này để định nghĩa struct

type Pokemon struct {
	Name   string   `json:"name"`
	Types  []string `json:"types"`
	Number string   `json:"number"`
	Stats  Stats    `json:"stats"`
	Exp    string   `json:"exp"`
}

type Stats struct {
	HP     int `json:"hp"`
	Attack int `json:"attack"`
	Defense int `json:"defense"`
	Speed  int `json:"speed"`
	SpAtk  int `json:"sp_atk"`
	SpDef  int `json:"sp_def"`
}

func main() {
	// Tạo context
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	// Kéo dài thời gian giờ cho các thao tác -> đảm bảo các thao tác sẽ ko vượt quá 90s
	ctx, cancel = context.WithTimeout(ctx, 90*time.Second)
	defer cancel()

	//Tạo biến chứa danh sách pokemon
	var pokemonList []Pokemon

	//CHỗ này để quyết định mình sẽ lấy về bao nhiêu pokemon, chỉnh i để quyết định coi đào bao nhiêu pokemon
	//Dùng chromedp.Run để điều khiển trình duyệt, tải trang và trích xuất dữ liệu từ trang https://pokedex.org
	for i := 1; i <= 10; i++ {
		var pokemon Pokemon
		var numberStr, hpStr, attackStr, defenseStr, speedStr, spAtkStr, spDefStr string
		err := chromedp.Run(ctx, 
			chromedp.Navigate(fmt.Sprintf("https://pokedex.org/#/pokemon/%d", i)), // Điều hướng đến trang của từng pokemon (i là number riêng biệt của mỗi pokemon)
			chromedp.Sleep(5*time.Second), //Chỗ này chỉ là thời gian chờ mà thôi
			// Đống ở dưới là để trích xuất các data từ chỗ tương ứng, ví dụ cái id: Nó sẽ lấy phần tử có class là detail-header và detail-national-id, sau đó lấy innerText của nó và loại bỏ ký tự # rồi quăng vô numberStr
			chromedp.Evaluate(`document.querySelector(".detail-header .detail-national-id").innerText.replace("#", "")`, &numberStr),
			chromedp.Evaluate(`document.querySelector(".detail-panel-header").innerText`, &pokemon.Name),
			chromedp.Evaluate(`Array.from(document.querySelectorAll('.detail-types span.monster-type')).map(elem => elem.innerText.toLowerCase())`, &pokemon.Types),
			chromedp.Evaluate(`Array.from(document.querySelectorAll('.detail-stats-row span')).filter(span => span.innerText.includes('HP'))[0].nextElementSibling.innerText`, &hpStr),
			chromedp.Evaluate(`Array.from(document.querySelectorAll('.detail-stats-row span')).filter(span => span.innerText.includes('Attack'))[0].nextElementSibling.innerText`, &attackStr),
			chromedp.Evaluate(`Array.from(document.querySelectorAll('.detail-stats-row span')).filter(span => span.innerText.includes('Defense'))[0].nextElementSibling.innerText`, &defenseStr),
			chromedp.Evaluate(`Array.from(document.querySelectorAll('.detail-stats-row span')).filter(span => span.innerText.includes('Speed'))[0].nextElementSibling.innerText`, &speedStr),
			chromedp.Evaluate(`Array.from(document.querySelectorAll('.detail-stats-row span')).filter(span => span.innerText.includes('Sp Atk'))[0].nextElementSibling.innerText`, &spAtkStr),
			chromedp.Evaluate(`Array.from(document.querySelectorAll('.detail-stats-row span')).filter(span => span.innerText.includes('Sp Def'))[0].nextElementSibling.innerText`, &spDefStr),
		)
		if err != nil {
			log.Fatalf("Failed to extract data for Number %d: %v", i, err)
		}
		// Quăng đống đào đc vô chỗ
		pokemon.Number = strings.TrimSpace(numberStr)
		pokemon.Stats.HP, _ = strconv.Atoi(strings.TrimSpace(hpStr))
		pokemon.Stats.Attack, _ = strconv.Atoi(strings.TrimSpace(attackStr))
		pokemon.Stats.Defense, _ = strconv.Atoi(strings.TrimSpace(defenseStr))
		pokemon.Stats.Speed, _ = strconv.Atoi(strings.TrimSpace(speedStr))
		pokemon.Stats.SpAtk, _ = strconv.Atoi(strings.TrimSpace(spAtkStr))
		pokemon.Stats.SpDef, _ = strconv.Atoi(strings.TrimSpace(spDefStr))

		//Và rồi quăng vô list thôi
		pokemonList = append(pokemonList, pokemon)
		fmt.Printf("Crawled data for Pokemon Number %d\n", i)
	}

	// Tạo collector mới
	c := colly.NewCollector(
		colly.AllowedDomains("bulbapedia.bulbagarden.net"),
	)

	// Khởi tạo một map để lưu trữ dữ liệu EXP của các Pokémon. Key của map là số thứ tự của Pokémon
	expMap := make(map[string]string)

	// Với mỗi hàng trong bảng (trừ tiêu đề)
	c.OnHTML("table.roundy tbody tr:not(:first-child)", func(e *colly.HTMLElement) {
		Number := strings.Trim(e.ChildText("td:nth-child(1)"), "\n ")
		exp := strings.Trim(e.ChildText("td:nth-child(4)"), "\n ") // Cột exp điều chỉnh đúng
		Number = strings.TrimLeft(Number, "0") // Bỏ số 0

		if Number != "" && exp != "" {
			expMap[Number] = exp
		}
	})

	c.Visit("https://bulbapedia.bulbagarden.net/wiki/List_of_Pok%C3%A9mon_by_effort_value_yield_(Generation_IX)")

	// Quăng đống exp mới đào đc vô chỗ info poke
	for i := range pokemonList {
		if exp, found := expMap[pokemonList[i].Number]; found {
			pokemonList[i].Exp = exp
		}
	}

	c.OnError(func(_ *colly.Response, err error) {
		log.Println("Something went wrong:", err)
	})
//------------------------------------------------------------------------------------

	// Mã hóa dữ liệu
	pokemonJSON, err := json.MarshalIndent(pokemonList, "", "    ")
	if err != nil {
		fmt.Println("Error encoding Pokemon data to JSON:", err)
		return
	}

	// Viết dữ liệu vào file json
	err = os.WriteFile("pokedex.json", pokemonJSON, 0644)
	if err != nil {
		fmt.Println("Error writing JSON data to file:", err)
		return
	}
	fmt.Println("Pokemon data saved to pokedex.json")
}
