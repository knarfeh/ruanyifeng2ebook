// Copyright © 2017 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	elastic "gopkg.in/olivere/elastic.v5"
)

// fetchCmd represents the fetch command
var fetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		main()
	},
}

func init() {
	RootCmd.AddCommand(fetchCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// fetchCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// fetchCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}

func main() {
	fmt.Println("ruanyifengeebook running...")
	URL := viper.GetString("URL")
	DAYTIMESTAMP := viper.GetString("DAY_TIME_STAMP")
	viper.SetDefault("ROUTINE_NUM", 30)
	ROUTINENUM := viper.GetInt("ROUTINE_NUM")
	var (
		maxRoutineNum = ROUTINENUM
		mutex         sync.Mutex
	)

	ESHOSTPORT := viper.GetString("ES_HOST_PORT")
	esClient, err := elastic.NewClient(elastic.SetURL(ESHOSTPORT))
	if err != nil {
		log.Printf("Unable to connect es")
		panic(err)
	}
	defer esClient.Stop()
	bulkRequest := esClient.Bulk()

	type esDoc struct {
		Title        string `json:"title"`
		Author       string `json:"author"`
		Content      string `json:"content"`
		DayTimeStamp string `json:"dayTimestamp"`
	}

	doc, err := goquery.NewDocument(URL)
	if err != nil {
		fmt.Println("Network issues...")
		log.Fatal(err)
	}

	ch := make(chan int, maxRoutineNum)

	documents := make([]*goquery.Document, 500)
	doc.Find(".content p").Each(func(i int, s *goquery.Selection) {
		title := s.Find("a").Text()
		href, _ := s.Find("a").Attr("href")
		articleHref := "http://www.aaronsw.com/weblog/" + href
		log.Printf("title %d: %s, href: %s\n", i, title, articleHref)

		ch <- 1
		go func() {
			documents[i], _ = goquery.NewDocument(articleHref)
			mutex.Lock()
			{
				documents[i].Find(".content").Each(func(i int, s *goquery.Selection) {
					articleTitle := s.Find("h1").Text()
					s.Find("h1").Remove()
					articleContent := s.Text()
					d := esDoc{
						Title:        articleTitle,
						Author:       "aaronsw",
						Content:      articleContent,
						DayTimeStamp: DAYTIMESTAMP,
					}
					{
						bulkData := elastic.NewBulkIndexRequest().Index("aaronsw").Type(URL + ":content").Id(articleTitle).Doc(d)
						bulkRequest = bulkRequest.Add(bulkData)
					}
				})
			}
			mutex.Unlock()
			<-ch
		}()
	})

	type metaData struct {
		Type     string `json:"type"`
		Title    string `json:"title"`
		BookDesp string `json:"book_desp"`
	}
	m := metaData{
		Type:     "aaronsw",
		Title:    "www-aaronsw-com",
		BookDesp: "RAW THOUGHT by Aaron Swartz",
	}
	bulkMetaData := elastic.NewBulkIndexRequest().Index("eebook").Type("metadata").Id(URL).Doc(m)
	bulkFinalRequest := bulkRequest.Add(bulkMetaData)

	bulkResponse, err := bulkFinalRequest.Do(context.TODO())
	if err != nil {
		fmt.Println("err???", err)
	}
	fmt.Println("bulkResponse: ", bulkResponse)
}
