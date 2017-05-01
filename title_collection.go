package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
)

type TitleCollection struct {
	titles []string
	index  int
	max    int
	size   int
}

func (t *TitleCollection) Pop() string {
	var title string

	if t.index < t.size {
		title = t.titles[t.index]
		t.index = t.index + 1
	} else {
		t.titles = get_titles(t.max)
		t.size = len(t.titles)
		title = t.titles[0]
		t.index = 1
	}
	return title
}

func NewTitleCollection(size int) *TitleCollection {
	return &TitleCollection{
		titles: []string{},
		index:  0,
		max:    size,
		size:   0,
	}
}

func get_titles(count int) []string {
	fmt.Println("Getting titles", count)
	var titles []string
	req := "https://en.wikipedia.org/w/api.php?action=query&list=random&rnnamespace=0&rnlimit=" + strconv.Itoa(count) + "&format=json"
	resp, err := http.Get(req)

	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	var i map[string]interface{}
	err = json.Unmarshal(body, &i)
	m := i["query"].(map[string]interface{})["random"]
	if err != nil {
		fmt.Println(err)
	}
	for _, k := range m.([]interface{}) {
		r := k.(map[string]interface{})
		titles = append(titles, r["title"].(string))
	}
	return titles
}
