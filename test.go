package main

import (
	"log"

	"github.com/jmcvetta/napping"
)

func main() {
	urlName := "suin"      // ここは自分のものに
	password := "p@ssW0rd" // ここは自分のものに

	// 入力JSONのフォーマットを定義 & 初期化
	input := struct {
		UrlName  string `json:"url_name"` // JSONのキーと構造体のキーが一致しないものは、ここでマッピングする
		Password string `json:"password"` // 同上
	}{
		urlName,
		password,
	}

	// 正常時のJSONのフォーマットを定義 & 初期化
	// Qiitaからのレスポンスには url_name も含まれるが、必要ないのでここでは定義しない
	response := struct {
		id int
	}{}
	// var response []interface{}

	// エラー時のJSONのフォーマットを定義 & 初期化
	error := struct {
		Error string
	}{}

	resp, err := napping.Post("http://user-server:3000/auth", &input, &response, &error) // input, response, errorは参照を渡す

	if err != nil {
		panic(err)
	}

	log.Printf("status: %#v", resp.Status())
	log.Printf("id: %#v", response.id)
	// fmt.Println(response)
	log.Printf("error message: %#v", error.Error)
}
